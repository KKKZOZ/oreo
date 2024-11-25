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
thread_load=50
threads=(8 16 32 48 64 80 96)
round_interval=5
bc=./BenConfig.yaml

wl_type=
verbose=false
remote=false
node2=s1-ljy
node3=s3-ljy
PASSWORD=kkkzoz

while [[ "$#" -gt 0 ]]; do
    case $1 in
    -wl | --workload)
        wl_type="$2"
        shift
        ;;
    -t | --threads)
        threads=($2)
        shift
        ;;
    -v | --verbose) verbose=true ;;
    -r | --remote) remote=true ;;
    *)
        echo "Unknown parameter passed: $1"
        exit 1
        ;;
    esac
    shift
done

# Validate workload
if [[ ! "$wl_type" =~ ^(iot|social|order)$ ]]; then
    echo "Error: Invalid workload. Must be iot, social or order"
    exit 1
fi

tar_dir=./data/$wl_type
config_file="./workloads/${wl_type}.yaml"
results_file="$tar_dir/${wl_type}_benchmark_results.csv"

log() {
    local color=${2:-$NC}
    if [[ "${verbose}" = true ]]; then
        echo -e "${color}$1${NC}"
    fi
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
    log "Running $wl_type $profile thread=$thread" $BLUE
    go run . -d oreo -wl $wl_type -wc "./workloads/$wl_type.yaml" -bc "$bc" -m $mode -ps $profile -t "$thread" >"$output"
}

load_data() {
    for profile in native cg oreo; do
        log "Loading to ${wl_type} $profile" $BLUE
        LOG=ERROR ./bin/cmd -d oreo -wl "$wl_type" -wc "$config_file" -bc "$bc" -m "load" -ps $profile -t "$thread_load"
        # run_workload "load" "$profile" "$thread_load" "/dev/null"
    done
    touch "$tar_dir/${wl_type}-load"
}

get_metrics() {
    local profile=$1 thread=$2
    local duration
    local file="$tar_dir/$wl_type-$profile-$thread.txt"
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
    local thread=$1 native_duration=$2 cg_duration=$3 oreo_duration=$4 native_p99=$5 cg_p99=$6 oreo_p99=$7
    local native_ratio=$8 cg_ratio=$9 oreo_ratio=${10}

    printf "%s:\nnative:%s\ncg    :%s\noreo  :%s\n" "${thread}" "${native_duration}" "${cg_duration}" "${oreo_duration}"

    local relative_native relative_cg
    relative_native=$(bc <<<"scale=5;${oreo_duration} / ${native_duration}")
    relative_cg=$(bc <<<"scale=5;${oreo_duration} / ${cg_duration}")

    printf "Oreo:native = %s\nOreo:cg     = %s\n" "${relative_native}" "${relative_cg}"
    printf "native 99th: %s\ncg     99th: %s\noreo   99th: %s\n" "${native_p99}" "${cg_p99}" "${oreo_p99}"
    printf "Error ratio:\nnative = %s\ncg = %s\noreo = %s\n" "${native_ratio}" "${cg_ratio}" "${oreo_ratio}"
    echo "---------------------------------"
}

clear_up() {
    if [ "$remote" = false ]; then
        log "Killing executor" $RED
        kill $executor_pid
        log "Killing time oracle" $RED
        kill $time_oracle_pid
    fi
}

deploy_local(){
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

deploy_remote(){
    log "Setup node 2" $GREEN
    ssh -t $node2 "echo '$PASSWORD' | sudo -S bash /home/liujinyi/oreo-ben/start-timeoracle.sh && sudo -S bash /home/liujinyi/oreo-ben/start-executor.sh -wl $wl_type -db $db_combinations"

    log "Setup node 3" $GREEN
    ssh -t $node3 "echo '$PASSWORD' | sudo -S bash /home/liujinyi/oreo-ben/start-executor.sh -wl $wl_type -db $db_combinations"
}

main() {

    # Go to the script root directory
    cd "$(dirname "$0")" && cd ..

    # check if config file exists
    if [ ! -f "$config_file" ]; then
        handle_error "Config file $config_file does not exist"
    fi

    if [ -z "$wl_type" ]; then
        handle_error "Workload type is not provided"
    fi

    log "Building the benchmark" $GREEN
    go build .
    mv cmd ./bin

    mkdir -p "$tar_dir"

    # Create/overwrite results file with header
    echo "thread,operation,native,cg,oreo,native_p99,cg_p99,oreo_p99,native_err,cg_err,oreo_err" >"$results_file"
    operation=$(rg '^operationcount' "$config_file" | rg -o '[0-9.]+')

    log "Running benchmark for [$wl_type] workload with [$db_combinations] database combinations" $YELLOW

    if [ "$remote" = true ]; then
        log "Running remotely" $BLUE
        deploy_remote
    else
        log "Running locally" $BLUE
        deploy_local
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

    if [ ! -f "$tar_dir/${wl_type}-load" ]; then
        load_data
    fi

    for thread in "${threads[@]}"; do

        for profile in native cg oreo; do
            output="$tar_dir/$wl_type-$profile-$thread.txt"
            run_workload "run" "$profile" "$thread" "$output"
        done

        read -r native_duration native_p99 native_ratio <<<"$(get_metrics "native" "$thread")"
        read -r cg_duration cg_p99 cg_ratio <<<"$(get_metrics "cg" "$thread")"
        read -r oreo_duration oreo_p99 oreo_ratio <<<"$(get_metrics "oreo" "$thread")"

        echo "$thread,$operation,$native_duration,$cg_duration,$oreo_duration,$native_p99,$cg_p99,$oreo_p99,$native_ratio,$cg_ratio,$oreo_ratio" >>"$results_file"

        print_summary "${thread}" "${native_duration}" "${cg_duration}" "${oreo_duration}" "${native_p99}" "${cg_p99}" "${oreo_p99}" "${native_ratio}" "${cg_ratio}" "${oreo_ratio}"

        sleep $round_interval
    done

    clear_up 
}

main "$@"
