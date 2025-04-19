#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

executor_port=8001
timeoracle_port=8010
thread_load=100
# threads=(8 16 32 48 64 80 96)
round_interval=5
threads=(80)

db_combinations=Redis,MongoDB1,Cassandra
thread=0
verbose=false
remote=false
skip=false
loaded=false
limited=false
wl_mode=

declare -g executor_pid
declare -g time_oracle_pid
node_list=(node2 node3 node4)
# node_list=(s1-ljy s3-ljy)
PASSWORD=kkkzoz

while [[ "$#" -gt 0 ]]; do
    case $1 in
    -wl | --workload)
        wl_mode="$2"
        shift
        ;;
    -db | --db)
        db_combinations="$2"
        shift
        ;;
    -t | --threads)
        threads=($2)
        shift
        ;;
    -v | --verbose) verbose=true ;;
    -r | --remote) remote=true ;;
    -s | --skip) skip=true ;;
    -l | --limited) limited=true ;;
    *)
        echo "Unknown parameter passed: $1"
        exit 1
        ;;
    esac
    shift
done

wl_type=ycsb
tar_dir=./data/ft
config_file="./workloads/ft/ft.yaml"
results_file="$tar_dir/ft_benchmark_results.csv"
bc=./config/BenConfig_ft.yaml
log_file="$tar_dir/benchmark.log"

log() {
    local color=${2:-$NC}
    if [[ "${verbose}" = true ]]; then
        echo -e "${color}$1${NC}"
    fi
}

handle_error() {
    echo "Error: $1"
    exit 1
}

kill_process_on_port() {
    local port=$1
    local pid
    pid=$(lsof -t -i ":$port")
    if [ -n "$pid" ]; then
        echo "Port $port is occupied by process $pid. Terminating this process..."
        kill -9 "$pid"
    fi
}

run_workload() {
    local mode=$1 profile=$2 thread=$3 output=$4
    log "Running $wl_type-$wl_mode $profile thread=$thread" $BLUE
    ./bin/cmd -ft -d oreo-ycsb -wl "$db_combinations" -wc "$config_file" -bc "$bc" -m "$mode" -ps "$profile" -t "$thread" >"$output" 2>"$log_file"
}

load_data() {
    for profile in oreo; do
        log "Loading to ${wl_type} $profile" $BLUE
        ./bin/cmd -ft -d oreo-ycsb -wl "$db_combinations" -wc "$config_file" -bc "$bc" -m "load" -ps $profile -t "$thread_load"
        # run_workload "load" "$profile" "$thread_load" "/dev/null"
    done
    touch "$tar_dir/${wl_type}-load"
}

get_metrics() {
    local profile=$1 thread=$2
    local duration
    local file="$tar_dir/$wl_type-$wl_mode-$db_combinations-$profile-$thread.txt"
    duration=$(rg '^Run finished' "$file" | rg -o '[0-9.]+')
    local duration
    latency=$(rg '^TXN\s' "$file" | rg -o '\s99th\(us\): [0-9]+' | cut -d' ' -f3)

    success_cnt=$(rg 'COMMIT ' "$file" | rg -o 'Count: [0-9]+' | cut -d ' ' -f 2)
    error_cnt=0
    if [ "$profile" != "native" ]; then
        error_cnt=$(rg 'COMMIT_ERROR ' "$file" | rg -o 'Count: [0-9]+' | cut -d ' ' -f 2)
    fi
    total=$((success_cnt + error_cnt))
    ratio=$(bc <<<"scale=4;$error_cnt / $total")
    echo "$duration $latency $ratio"

}

print_summary() {
    local thread=$1 oreo_duration=$2 oreo_p99=$3 oreo_ratio=$4

    printf "%s:\n" "${thread}"
    printf "Oreo: %s\n" "${oreo_duration}"
    printf "Oreo 99th: %s\n" "${oreo_p99}"
    printf "Oreo error ratio: %s\n" "${oreo_ratio}"
    # printf "%s:\nnative:%s\ncg    :%s\noreo  :%s\n" "${thread}" "${native_duration}" "${cg_duration}" "${oreo_duration}"

    # local relative_native relative_cg
    # relative_native=$(bc <<<"scale=5;${oreo_duration} / ${native_duration}")
    # relative_cg=$(bc <<<"scale=5;${oreo_duration} / ${cg_duration}")

    # printf "Oreo:native = %s\nOreo:cg     = %s\n" "${relative_native}" "${relative_cg}"
    # printf "native 99th: %s\ncg     99th: %s\noreo   99th: %s\n" "${native_p99}" "${cg_p99}" "${oreo_p99}"
    # printf "Error ratio:\nnative = %s\ncg = %s\noreo = %s\n" "${native_ratio}" "${cg_ratio}" "${oreo_ratio}"
    # echo "---------------------------------"
}

clear_up() {
    if [ "$remote" = false ]; then
        log "Killing executor"
        kill $executor_pid
        log "Killing time oracle"
        kill $time_oracle_pid
    fi
}

deploy_local() {
    kill_process_on_port "$executor_port"
    kill_process_on_port "$timeoracle_port"

    log "Starting executor"
    ./bin/executor -p "$executor_port" -w $wl_type -bc "$bc" -db "$db_combinations" 2>./log/executor.log &
    # env LOG=ERROR $(build_command) 2>./log/executor.log &
    executor_pid=$!

    log "Starting time oracle"
    ./bin/timeoracle -p "$timeoracle_port" -type hybrid >/dev/null 2>./log/timeoracle.log &
    time_oracle_pid=$!
}

deploy_remote() {
    log "Setup timeoracle on node 2" "$GREEN"
    ssh -t ${node_list[0]} "echo '$PASSWORD' | sudo -S bash /root/oreo-ben/start-ft-timeoracle.sh "
    ssh -t ${node_list[0]} "sudo systemctl restart haproxy"

    if [ "$limited" = true ]; then
       for node in "${node_list[@]}"; do
            log "Setup $node" "$GREEN"
            ssh -t "$node" "echo '$PASSWORD' | sudo -S bash /root/oreo-ben/start-ft-executor-docker.sh -p 8001 -wl $wl_type -db $db_combinations -bc BenConfig_ft.yaml -r -l"
            ssh -t "$node" "echo '$PASSWORD' | sudo -S bash /root/oreo-ben/start-ft-executor-docker.sh -p 8002 -wl $wl_type -db $db_combinations -bc BenConfig_ft.yaml -l"
        done
    else 
        for node in "${node_list[@]}"; do
            log "Setup $node" "$GREEN"
            ssh -t "$node" "echo '$PASSWORD' | sudo -S bash /root/oreo-ben/start-ft-executor-docker.sh -p 8001 -wl $wl_type -db $db_combinations -r"
            ssh -t "$node" "echo '$PASSWORD' | sudo -S bash /root/oreo-ben/start-ft-executor-docker.sh -p 8002 -wl $wl_type -db $db_combinations"
        done
    fi
}

main() {
    cd "$(dirname "$0")" && cd ..

    # check if config file exists
    if [ ! -f "$config_file" ]; then
        handle_error "Config file $config_file does not exist"
    fi

    if [ ! -f "$bc" ]; then
        handle_error "Config file $bc does not exist"
    fi

    echo "Building the benchmark executable"
    go build .
    mv cmd ./bin

    tar_dir="$tar_dir/$wl_mode-$db_combinations"
    mkdir -p "$tar_dir"

    # Create/overwrite results file with header
    echo "thread,operation,oreo,oreo_p99,oreo_err" >"$results_file"
    operation=$(rg '^operationcount' "$config_file" | rg -o '[0-9.]+')

    log "Running benchmark for [$wl_type] workload with [$db_combinations] database combinations" $YELLOW

    if [ "$skip" = true ]; then
        log "Skipping deployment" $YELLOW
    else
        if [ "$remote" = true ]; then
            echo "Running remotely"
            deploy_remote
        else
            echo "Running locally"
            deploy_local
        fi
    fi

    if [ "$loaded" = true ]; then
        log "Skipping data loading" $YELLOW
    else
        if [ ! -f "$tar_dir/${wl_type}-load" ]; then
            log "Ready to load data" $YELLOW
        else
            log "Data has been already loaded" $YELLOW
        fi
    fi

    read -p "Do you want to continue? (y/n): " continue_choice
    if [[ "$continue_choice" != "y" && "$continue_choice" != "Y" ]]; then
        echo "Exiting the script."
        exit 0
    fi

    # Load data if needed
    if [ ! -f "$tar_dir/${wl_type}-load" ]; then
        load_data
    fi

    for thread in "${threads[@]}"; do

        for profile in oreo; do
            output="$tar_dir/$wl_type-$wl_mode-$db_combinations-$profile-$thread.txt"
            run_workload "run" "$profile" "$thread" "$output"
        done

        # read -r native_duration native_p99 native_ratio <<<"$(get_metrics "native" "$thread")"
        # read -r cg_duration cg_p99 cg_ratio <<<"$(get_metrics "cg" "$thread")"
        read -r oreo_duration oreo_p99 oreo_ratio <<<"$(get_metrics "oreo" "$thread")"

        echo "$thread,$operation,$oreo_duration,$oreo_p99,$oreo_ratio" >>"$results_file"

        print_summary "${thread}" "${oreo_duration}" "${oreo_p99}" "${oreo_ratio}"

        # sleep $round_interval
    done

    mv timeline.csv "./data/ft/timeline.csv"

    python3 ./analysis/process_timeline_logs.py --file ./data/ft/timeline.csv --span 100

    clear_up
}

main "$@"
