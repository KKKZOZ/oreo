#!/bin/bash

executor_port=8001
timeoracle_port=8010
thread_load=50

db_combinations=
thread=

wl_type=ycsb
tar_dir=$wl_type
results_file="$tar_dir/${wl_type}_benchmark_results.csv"

if [ $# -eq 1 ]; then
    verbose=false
else
    verbose=true
fi

log() {
    [[ "${verbose}" = true ]] && echo "$@"
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
    log "Running $wl_type $profile thread=$thread"
    ./bin/cmd -d oreo-ycsb -wl $wl_type -wc "./workloads/$wl_type.yaml" -m $mode -ps $profile -t "$thread" >"$output"
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
    log "Killing executor"
    kill $executor_pid
    log "Killing time oracle"
    kill $time_oracle_pid
}

main() {
    cd "$(dirname "$0")" && cd ..

    echo "Building the benchmark"
    go build .
    mv cmd ./bin

    # Create/overwrite results file with header
    echo "thread,operation,native,cg,oreo,native_p99,cg_p99,oreo_p99" >"$results_file"

    operation=$(rg '^operationcount' ./workloads/$wl_type.yaml | rg -o '[0-9.]+')

    mkdir -p "$tar_dir"

    kill_process_on_port "$executor_port"
    kill_process_on_port "$timeoracle_port"

    log "Starting executor"
    ./bin/executor -p "$executor_port" -timeurl "http://localhost:$timeoracle_port" -w $wl_type -kvrocks localhost:6379 -mongo1 mongodb://localhost:27018 2>./log/executor.log &
    executor_pid=$!

    log "Starting time oracle"
    ./bin/timeoracle -p "$timeoracle_port" -type hybrid >/dev/null 2>./log/timeoracle.log &
    time_oracle_pid=$!

    # Load data if needed
    if [ ! -f "$tar_dir/${wl_type}-load" ]; then
        load_data
    else
        log "Data has been already loaded"
    fi
}

main "$@"
