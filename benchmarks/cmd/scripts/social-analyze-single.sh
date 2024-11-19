#!/bin/bash

# 检查是否提供了至少一个参数
if [ $# -lt 1 ]; then
    echo "Error: Missing required argument."
    echo "Usage: $0 <thread> [less-verbose]"
    exit 1
fi

if [ $# -eq 2 ]; then
    verbose=false
else
    verbose=true
fi

verbose_echo() {
    if [ "$verbose" = true ]; then
        echo "$@"
    fi
}

thread=$1
executor_port=8001
timeoracle_port=8010
thread_load=25
wl_type=social
tar_dir=$wl_type
bc=./BenConfig.yaml

# Go to the script root directory
cd "$(dirname "$0")" && cd ..

# Clear files
# rm iot-native.txt iot-oreo.txt iot-cg.txt
mkdir -p "$tar_dir"
# rm "$tar_dir/iot-*.txt"
rm "$tar_dir"/*.txt

pid=$(lsof -t -i ":$executor_port")
if [ -n "$pid" ]; then
    echo "Port $executor_port is occupied by process $pid. Terminating this process..."
    kill -9 "$pid"
fi

pid=$(lsof -t -i ":$timeoracle_port")
if [ -n "$pid" ]; then
    echo "Port $timeoracle_port is occupied by process $pid. Terminating this process..."
    kill -9 "$pid"
fi

verbose_echo "Starting executor"
# ./executor -p "$executor_port" -timeurl "http://localhost:$timeoracle_port" -w iot -kvrocks localhost:6379 -mongo1 mongodb://localhost:27018 >/dev/null 2>executor.log &
./bin/executor -p "$executor_port" -timeurl "http://localhost:$timeoracle_port" -w $wl_type -redis1 localhost:6379 -mongo1 mongodb://localhost:27018 -couch http://admin:password@localhost:5984 2>./log/executor.log &
executor_pid=$!

verbose_echo "Starting time oracle"

./bin/timeoracle -p "$timeoracle_port" -type hybrid >/dev/null 2>./log/timeoracle.log &
time_oracle_pid=$!

if [ ! -f "$tar_dir/$wl_type-load" ]; then

    # verbose_echo "Loading to $wl_type native"
    # go run . -d oreo -wl $wl_type -wc ./workloads/$wl_type.yaml -m load -ps native -t "$thread_load"

    # verbose_echo "Loading to $wl_type Cherry Garcia"
    # go run . -d oreo -wl $wl_type -wc ./workloads/$wl_type.yaml -m load -ps cg -t "$thread_load"

    verbose_echo "Loading to $wl_type oreo"
    go run . -d oreo -wl $wl_type -wc ./workloads/$wl_type.yaml -bc "bc" -m load -ps oreo -t "$thread_load"

    touch "$tar_dir/${wl_type}-load"
else
    verbose_echo "Data has been already loaded"
fi

# verbose_echo "Running $wl_type native"
# go run . -d oreo -wl $wl_type -wc ./workloads/$wl_type.yaml -bc "$bc" -m run -ps native -pprof -t "$thread" >"$tar_dir/$wl_type-native.txt"

# verbose_echo "Running $wl_type Cherry Garcia"
# go run . -d oreo -wl $wl_type -wc ./workloads/$wl_type.yaml -bc "$bc" -m run -ps cg -t "$thread" >"$tar_dir/$wl_type-cg.txt"

verbose_echo "Running $wl_type oreo"
go run . -d oreo -wl $wl_type -wc ./workloads/$wl_type.yaml -bc "$bc" -m run -ps oreo -t "$thread" >"$tar_dir/$wl_type-oreo.txt"

native=$(rg '^Run finished' "$tar_dir/$wl_type-native.txt" | rg -o '[0-9.]+')
cg=$(rg '^Run finished' "$tar_dir/$wl_type-cg.txt" | rg -o '[0-9.]+')
oreo=$(rg '^Run finished' "$tar_dir/$wl_type-oreo.txt" | rg -o '[0-9.]+')

native_p99=$(rg '^TXN\s' "$tar_dir/$wl_type-native.txt" | rg -o '\s99th\(us\): [0-9]+' | cut -d' ' -f3)
cg_p99=$(rg '^TXN\s' "$tar_dir/$wl_type-cg.txt" | rg -o '\s99th\(us\): [0-9]+' | cut -d' ' -f3)
oreo_p99=$(rg '^TXN\s' "$tar_dir/$wl_type-oreo.txt" | rg -o '\s99th\(us\): [0-9]+' | cut -d' ' -f3)

printf "%s:\nnative:%s\ncg    :%s\noreo  :%s\n" "$thread" "$native" "$cg" "$oreo"

relative_native=$(echo "scale=5;$oreo / $native" | bc)
relative_cg=$(echo "scale=5;$oreo / $cg" | bc)

printf "Oreo:native = %s\n" "$relative_native"
printf "Oreo:cg     = %s\n" "$relative_cg"
printf "native 99th: %s\ncg     99th: %s\noreo   99th: %s\n" "$native_p99" "$cg_p99" "$oreo_p99"

verbose_echo "Killing executor"
kill -s TERM $executor_pid
verbose_echo "Killing time oracle"
kill $time_oracle_pid

# python3 timeoracle-latency-analyze.py
