#!/bin/bash

executor_port=8001
timeoracle_port=8010
thread_load=50
threads=(32 64 96)
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
    -v | --verbose) verbose=true ;;
    -r | --remote) remote=true ;;
    *)
        echo "Unknown parameter passed: $1"
        exit 1
        ;;
    esac
    shift
done

tar_dir=$wl_type
config_file="./workloads/${wl_type}.yaml"
results_file="$tar_dir/${wl_type}_benchmark_results.csv"

log() {
    [[ "${verbose}" = true ]] && echo "$@"
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
    log "Running $wl_type $profile thread=$thread"
    go run . -d oreo -wl $wl_type -wc "./workloads/$wl_type.yaml" -bc "$bc" -m $mode -ps $profile -t "$thread" >"$output"
}

load_data() {
    for profile in native cg oreo; do
        log "Loading to ${wl_type} $profile"
        run_workload "load" "$profile" "$thread_load" "/dev/null"
    done
    touch "$tar_dir/${wl_type}-load"
}

get_metrics() {
    local profile=$1 thread=$2
    local duration
    duration=$(rg '^Run finished' "$tar_dir/$wl_type-$profile-$thread.txt" | rg -o '[0-9.]+')
    local duration
    latency=$(rg '^TXN\s' "$tar_dir/$wl_type-$profile-$thread.txt" | rg -o '\s99th\(us\): [0-9]+' | cut -d' ' -f3)
    echo "$duration $latency"

}

print_summary() {
    local thread=$1 native_duration=$2 cg_duration=$3 oreo_duration=$4 native_p99=$5 cg_p99=$6 oreo_p99=$7

    printf "%s:\nnative:%s\ncg    :%s\noreo  :%s\n" "${thread}" "${native_duration}" "${cg_duration}" "${oreo_duration}"

    local relative_native relative_cg
    relative_native=$(bc <<<"scale=5;${oreo_duration} / ${native_duration}")
    relative_cg=$(bc <<<"scale=5;${oreo_duration} / ${cg_duration}")

    printf "Oreo:native = %s\nOreo:cg     = %s\n" "${relative_native}" "${relative_cg}"
    printf "native 99th: %s\ncg     99th: %s\noreo   99th: %s\n" "${native_p99}" "${cg_p99}" "${oreo_p99}"
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
    log "Setup node 2"
    ssh -t $node2 "echo '$PASSWORD' | sudo -S bash /home/liujinyi/oreo-ben/start-timeoracle.sh && sudo -S bash /home/liujinyi/oreo-ben/start-executor.sh -wl $wl_type -db $db_combinations"

    log "Setup node 3"
    ssh -t $node3 "echo '$PASSWORD' | sudo -S bash /home/liujinyi/oreo-ben/start-executor.sh -wl $wl_type -db $db_combinations"

    read -p "Do you want to continue? (y/n): " continue_choice
    if [[ "$continue_choice" != "y" && "$continue_choice" != "Y" ]]; then
        echo "Exiting the script."
        exit 0
    fi
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

    echo "Building the benchmark"
    go build .
    mv cmd ./bin

    mkdir -p "$tar_dir"

    # Create/overwrite results file with header
    echo "thread,operation,native,cg,oreo,native_p99,cg_p99,oreo_p99,native_err,cg_err,oreo_err" >"$results_file"
    operation=$(rg '^operationcount' "$config_file" | rg -o '[0-9.]+')

    printf "Running benchmark for [%s] workload with [%s] database combinations\n" "$wl_type" "$db_combinations"

    if [ "$remote" = true ]; then
        echo "Running remotely"
        deploy_remote
    else
        echo "Running locally"
        deploy_local
    fi

    # Load data if needed
    if [ ! -f "$tar_dir/${wl_type}-load" ]; then
        load_data
    else
        log "Data has been already loaded"
    fi

    for thread in "${threads[@]}"; do

        for profile in native cg oreo; do
            output="$tar_dir/$wl_type-$profile-$thread.txt"
            run_workload "run" "$profile" "$thread" "$output"
        done

        read -r native_duration native_p99 <<<"$(get_metrics "native" "$thread")"
        read -r cg_duration cg_p99 <<<"$(get_metrics "cg" "$thread")"
        read -r oreo_duration oreo_p99 <<<"$(get_metrics "oreo" "$thread")"

        eecho "$thread,$operation,$native_duration,$cg_duration,$oreo_duration,$native_p99,$cg_p99,$oreo_p99,$native_ratio,$cg_ratio,$oreo_ratio" >>"$results_file"

        print_summary "${thread}" "${native_duration}" "${cg_duration}" "${oreo_duration}" "${native_p99}" "${cg_p99}" "${oreo_p99}"
    done

    clear_up 
}

main "$@"
