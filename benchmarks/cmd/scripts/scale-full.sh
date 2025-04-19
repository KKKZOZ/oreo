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
threads=(64 80 96 112 138)
round_interval=2

thread=0
verbose=false
wl_mode=
db_combinations='MongoDB1,Cassandra'
skip=false

declare -g executor_pid
declare -g time_oracle_pid
remote=false
node_list=(node2 node3)
# node_list=(s1-ljy s3-ljy)
PASSWORD=kkkzoz

while [[ "$#" -gt 0 ]]; do
    case $1 in
    -wl | --workload)
        wl_mode="$2"
        shift
        ;;
    -t | --threads)
        threads=($2)
        shift
        ;;
    -v | --verbose) verbose=true ;;
    -r | --remote) remote=true ;;
    -s | --skip) skip=true ;;
    *)
        echo "Unknown parameter passed: $1"
        exit 1
        ;;
    esac
    shift
done

wl_type=scale
tar_dir=./data/scale
config_file="./workloads/${wl_mode}_${db_combinations}.yaml"
results_file="$tar_dir/${wl_mode}_${db_combinations}_benchmark_results.csv"
bc=./config/BenConfig_ycsb.yaml
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
    local mode=$1 profile=$2 thread=$3 output=$4 read_strategy=$5
    log "Running $wl_type-$wl_mode $profile thread=$thread readStrategy=$read_strategy" $BLUE
    go run . -d oreo-ycsb -wl "$db_combinations" -wc "$config_file" -bc "$bc" -m "$mode" -ps "$profile" -read "$read_strategy" -t "$thread" >"$output" 2>"$log_file"
}

load_data() {
    for profile in oreo; do
        log "Loading to ${wl_type} $profile" $BLUE
        go run . -d oreo-ycsb -wl "$db_combinations" -wc "$config_file" -bc "$bc" -m "load" -ps $profile -t "$thread_load"
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

    error_cnt=0

    # error_cnt=$((error_cnt + $(rg 'key not found in given RecordLen' "$file" | rg -o '[0-9]+' || echo 0)))

    # error_cnt=$((error_cnt + $(rg 'key not found prev is empty' "$file" | rg -o '[0-9]+' || echo 0)))

    echo "$duration $latency $error_cnt"

}

print_summary() {
    local thread=$1
    local opt1_duration=$2
    local opt2_duration=$3
    local opt3_duration=$4
    local opt4_duration=$5
    local opt1_p99=$6
    local opt2_p99=$7
    local opt3_p99=$8
    local opt4_p99=$9
    local opt1_ratio=${10}
    local opt2_ratio=${11}
    local opt3_ratio=${12}
    local opt4_ratio=${13}

    printf "%s:\nopt1:%s\nopt2:%s\nopt3:%s\nopt4:%s\n" "${thread}" "${opt1_duration}" "${opt2_duration}" "${opt3_duration}" "${opt4_duration}"

    local relative_opt2 relative_opt3 relative_opt4
    relative_opt2=$(bc <<<"scale=5;${opt2_duration} / ${opt1_duration}")
    relative_opt3=$(bc <<<"scale=5;${opt3_duration} / ${opt1_duration}")
    relative_opt4=$(bc <<<"scale=5;${opt4_duration} / ${opt1_duration}")

    printf "Opt2:opt1 = %s\nOpt3:opt1 = %s\nOpt4:opt1 = %s\n" "${relative_opt2}" "${relative_opt3}" "${relative_opt4}"
    printf "opt1 99th: %s\nopt2 99th: %s\nopt3 99th: %s\nopt4 99th: %s\n" "${opt1_p99}" "${opt2_p99}" "${opt3_p99}" "${opt4_p99}"
    printf "Read Error Count:\nopt1 = %s\nopt2 = %s\nopt3 = %s\nopt4 = %s\n" "${opt1_ratio}" "${opt2_ratio}" "${opt3_ratio}" "${opt4_ratio}"
    echo "---------------------------------"
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
    log "Setup timeoracle on node 2" $GREEN
    ssh -t ${node_list[0]} "echo '$PASSWORD' | sudo -S bash /root/oreo-ben/start-timeoracle.sh "
    # ssh -t ${node_list[0]} "echo '$PASSWORD' | sudo -S bash /home/liujinyi/oreo-ben/start-timeoracle.sh "

    # for node in "${node_list[@]}"; do
    #     log "Setup $node" $GREEN
    #     ssh -t $node "echo '$PASSWORD' | sudo -S bash /root/oreo-ben/start-executor.sh -wl ycsb -db $db_combinations"
    # done

    for node in "${node_list[@]}"; do
        log "Setup $node" $GREEN
        # ssh -t $node "echo '$PASSWORD' | sudo -S bash /root/oreo-ben/start-executor-docker.sh -l -p 8001 -wl ycsb -db $db_combinations"
        ssh -t "$node" "echo '$PASSWORD' | sudo -S bash /root/oreo-ben/start-ft-executor-docker.sh -p 8001 -wl ycsb -db $db_combinations -bc BenConfig_ycsb.yaml -r"
    done
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

    if [ $skip = false ]; then
        echo "Building the benchmark executable"
        go build .
        mv cmd ./bin
    fi

    tar_dir="$tar_dir/$wl_mode-$db_combinations"
    mkdir -p "$tar_dir"

    # Create/overwrite results file with header
    # echo "thread,operation,oreo,oreo-p,oreo-p_err" >"$results_file"
    operation=$(rg '^operationcount' "$config_file" | rg -o '[0-9.]+')

    log "Running benchmark for [$wl_type] workload with [$db_combinations] database combinations" $GREEN

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

    if [ ! -f "$tar_dir/${wl_type}-load" ]; then
        log "Ready to load data" $YELLOW
    else
        log "Data has been already loaded" $YELLOW
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

        # output="$tar_dir/$wl_type-$wl_mode-$db_combinations-native-p-$thread.txt"
        # run_workload "run" "native" "$thread" "$output" "p"

        # output="$tar_dir/$wl_type-$wl_mode-$db_combinations-cg-p-$thread.txt"
        # run_workload "run" "cg" "$thread" "$output" "p"

        output="$tar_dir/$wl_type-$wl_mode-$db_combinations-oreo-p-$thread.txt"
        run_workload "run" "oreo" "$thread" "$output" "p"

        # read -r native native_p99 native_err <<<"$(get_metrics "native" "p-$thread")"
        # read -r cg cg_p99 cg_err <<<"$(get_metrics "cg" "p-$thread")"
        read -r oreo oreo_p99 oreo_err <<<"$(get_metrics "oreo" "p-$thread")"

        echo "$thread,$operation,${oreo},${oreo_p99},${oreo_err}" >>"$results_file"

        # printf "native-p: %s\nnative-p_p99: %s\nnative-p_err: %s\n" ${native} ${native_p99} ${native_err}
        # printf "cg-p: %s\ncg-p_p99: %s\ncg-p_err: %s\n" ${cg} ${cg_p99} ${cg_err}
        printf "oreo-p: %s\noreo-p_p99: %s\noreo-p_err: %s\n" ${oreo} ${oreo_p99} ${oreo_err}

        # echo "$thread,$operation,$opt1_duration,$opt2_duration,$opt3_duration,$opt4_duration,$opt1_p99,$opt2_p99,$opt3_p99,$opt4_p99,$opt1_ratio,$opt2_ratio,$opt3_ratio,$opt4_ratio" >>"$results_file"

        # print_summary "${thread}" "${opt1_duration}" "${opt2_duration}" "${opt3_duration}" "${opt4_duration}" "${opt1_p99}" "${opt2_p99}" "${opt3_p99}" "${opt4_p99}" "${opt1_ratio}" "${opt2_ratio}" "${opt3_ratio}" "${opt4_ratio}"

    done

    clear_up
}

main "$@"
