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
thread_load=50
wl_type=iot
tar_dir=$wl_type

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
./executor -p "$executor_port" -timeurl "http://localhost:$timeoracle_port" -w iot -kvrocks localhost:6379 -mongo1 mongodb://localhost:27018 >/dev/null 2>executor.log &
# ./executor -p "$executor_port" -w iot -kvrocks 39.104.105.27:6379 -mongo1 mongodb://39.104.105.27:27018 > /dev/null 2> executor.log &
executor_pid=$!

verbose_echo "Starting time oracle"

./timeoracle -p "$timeoracle_port" -type hybrid >/dev/null 2>timeoracle.log &
time_oracle_pid=$!

if [ ! -f "iot-load" ]; then

  verbose_echo "Loading to IOT native"
  # ben load iot native "$thread_load"
  go run . -d oreo -wl iot -wc ./workloads/iot.yaml -m load -ps native -t "$thread_load"

  verbose_echo "Loading to IOT Cherry Garcia"
  # ben load iot native "$thread_load"
  go run . -d oreo -wl iot -wc ./workloads/iot.yaml -m load -ps cg -t "$thread_load"

  verbose_echo "Loading to IOT oreo"
  # ben load iot oreo "$thread_load"
  go run . -d oreo -wl iot -wc ./workloads/iot.yaml -m load -ps oreo -t "$thread_load"

  touch iot-load
else
  verbose_echo "Data has been already loaded"
fi

verbose_echo "Running IOT native"
go run . -d oreo -wl iot -wc ./workloads/iot.yaml -m run -ps native -t "$thread" >"$tar_dir/iot-native.txt"

verbose_echo "Running IOT Cherry Garcia"
go run . -d oreo -wl iot -wc ./workloads/iot.yaml -m run -ps cg -t "$thread" >"$tar_dir/iot-cg.txt"

verbose_echo "Running IOT oreo"
go run . -d oreo -wl iot -wc ./workloads/iot.yaml -m run -ps oreo -remote -t "$thread" >"$tar_dir/iot-oreo.txt"

native=$(rg '^Run finished' "$tar_dir/iot-native.txt" | rg -o '[0-9.]+')
cg=$(rg '^Run finished' "$tar_dir/iot-cg.txt" | rg -o '[0-9.]+')
oreo=$(rg '^Run finished' "$tar_dir/iot-oreo.txt" | rg -o '[0-9.]+')

native_p99=$(rg '^TXN\s' "$tar_dir/iot-native.txt" | rg -o '\s99th\(us\): [0-9]+' | cut -d' ' -f3)
cg_p99=$(rg '^TXN\s' "$tar_dir/iot-cg.txt" | rg -o '\s99th\(us\): [0-9]+' | cut -d' ' -f3)
oreo_p99=$(rg '^TXN\s' "$tar_dir/iot-oreo.txt" | rg -o '\s99th\(us\): [0-9]+' | cut -d' ' -f3)

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
