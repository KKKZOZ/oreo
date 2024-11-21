#!/bin/bash

executor_port=8001
timeoracle_port=8010
thread_load=100
threads=(16 32 48 64 80 96)
# threads=(8 16 32)

db_combinations=
thread=0
verbose=false
wl_mode=

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
    -v | --verbose) verbose=true ;;
    *)
        echo "Unknown parameter passed: $1"
        exit 1
        ;;
    esac
    shift
done

wl_type=ycsb
tar_dir=ycsb
config_file="./workloads/${wl_mode}_${db_combinations}.yaml"
results_file="$tar_dir/${wl_mode}_${db_combinations}_benchmark_results.csv"
bc=./BenConfig.yaml

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
    log "Running $wl_type-$wl_mode $profile thread=$thread"
    ./bin/cmd -d oreo-ycsb -wl "$db_combinations" -wc "$config_file" -bc "$bc" -m "$mode" -ps "$profile" -t "$thread" >"$output"
}

load_data() {
    for profile in native cg oreo; do
        # for profile in native cg oreo; do
        log "Loading to ${wl_type} $profile"
        ./bin/cmd -d oreo-ycsb -wl "$db_combinations" -wc "$config_file" -bc "$bc" -m "load" -ps $profile -t "$thread_load"
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
    log "Killing executor"
    kill $executor_pid
    log "Killing time oracle"
    kill $time_oracle_pid
}

# 基础命令部分
base_cmd="./bin/executor -p $executor_port -timeurl http://localhost:$timeoracle_port -w $wl_type"

# 数据库连接配置
declare -A db_configs=(
    ["TiKV"]="-tikv localhost:2379"
    ["KVRocks"]="-kvrocks localhost:6666"
    ["MongoDB"]="-mongo1 mongodb://172.24.58.116:27018"
    ["Redis"]="-redis1 localhost:6379"
    ["CouchDB"]="-couch http://admin:password@localhost:5984"
    ["Cassandra"]="-cas localhost"
    ["DynamoDB"]="-dynamodb http://localhost:8000"
)

# 构建命令的函数
build_command() {
    local final_cmd="$base_cmd"
    # 将 db_combinations 按逗号分割
    IFS=',' read -ra DBS <<<"$db_combinations"

    # 遍历选择的数据库
    for db in "${DBS[@]}"; do
        # 去除可能存在的空格
        db=$(echo "$db" | xargs)
        # 如果这个数据库在配置中存在，则添加相应的参数
        if [[ -n "${db_configs[$db]}" ]]; then
            final_cmd="$final_cmd ${db_configs[$db]}"
        fi
    done
    echo "$final_cmd"
}

main() {
    cd "$(dirname "$0")" && cd ..

    # check if config file exists
    if [ ! -f "$config_file" ]; then
        handle_error "Config file $config_file does not exist"
    fi

    echo "Building the benchmark executable"
    go build .
    mv cmd ./bin

    tar_dir="$tar_dir/$wl_mode-$db_combinations"
    mkdir -p "$tar_dir"

    # Create/overwrite results file with header
    echo "thread,operation,native,cg,oreo,native_p99,cg_p99,oreo_p99" >"$results_file"
    operation=$(rg '^operationcount' "$config_file" | rg -o '[0-9.]+')

    kill_process_on_port "$executor_port"
    kill_process_on_port "$timeoracle_port"

    log "Starting executor"
    LOG=ERROR ./bin/executor -p "$executor_port" -w $wl_type -bc "$bc" -db "$db_combinations" 2>./log/executor.log &
    # env LOG=ERROR $(build_command) 2>./log/executor.log &
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

    for thread in "${threads[@]}"; do

        for profile in native cg oreo; do
            output="$tar_dir/$wl_type-$wl_mode-$db_combinations-$profile-$thread.txt"
            run_workload "run" "$profile" "$thread" "$output"
        done

        read -r native_duration native_p99 native_ratio <<<"$(get_metrics "native" "$thread")"
        read -r cg_duration cg_p99 cg_ratio <<<"$(get_metrics "cg" "$thread")"
        read -r oreo_duration oreo_p99 oreo_ratio <<<"$(get_metrics "oreo" "$thread")"

        echo "$thread,$operation,$native_duration,$cg_duration,$oreo_duration,$native_p99,$cg_p99,$oreo_p99" >>"$results_file"

        print_summary "${thread}" "${native_duration}" "${cg_duration}" "${oreo_duration}" "${native_p99}" "${cg_p99}" "${oreo_p99}" "${native_ratio}" "${cg_ratio}" "${oreo_ratio}"
    done

    clear_up
}

main "$@"
