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
thread_load=10
tar_dir=iot


# Clear files
# rm iot-native.txt iot-oreo.txt iot-cg.txt
mkdir -p "$tar_dir"

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
# ./executor -p "$executor_port" -w iot -kvrocks localhost:6379 -mongo1 mongodb://localhost:27018 > /dev/null 2> executor.log &
./executor -p "$executor_port" -w iot -kvrocks 39.104.105.27:6379 -mongo1 mongodb://39.104.105.27:27018 > /dev/null 2> executor.log &
executor_pid=$!

verbose_echo "Starting time oracle"

LOG=DEBUG ./timeoracle -p "$timeoracle_port" -type counter > /dev/null 2> timeoracle.log &
time_oracle_pid=$!


if [ ! -f "iot-load" ]; then

  verbose_echo "Loading to IOT oreo"
  ben load iot oreo "$thread_load"

  verbose_echo "Loading to IOT native"
  ben load iot native "$thread_load"

  touch iot-load
else
  verbose_echo "Data has been already loaded"
fi

verbose_echo "Running IOT native"
go run . -d native -wl iot -wc ./workloads/iot.yaml -m run -ps native -t "$thread" > "$tar_dir/iot-native.txt"

verbose_echo "Running IOT Cherry Garcia"
go run . -d oreo -wl iot -wc ./workloads/iot.yaml -m run -ps oreo -ps cg -t "$thread" > "$tar_dir/iot-cg.txt"

verbose_echo "Running IOT oreo"
go run . -d oreo -wl iot -wc ./workloads/iot.yaml -m run -ps oreo -remote -t "$thread" > "$tar_dir/iot-oreo.txt"

native=$(sed -n '21p' "$tar_dir/iot-native.txt" | grep -E -o '[0-9.]+(ms|s)' | sed 's/\(ms\|s\)$//') 
cg=$(sed -n '21p' "$tar_dir/iot-cg.txt" | grep -E -o '[0-9.]+(ms|s)' | sed 's/\(ms\|s\)$//') 
oreo=$(sed -n '21p' "$tar_dir/iot-oreo.txt" | grep -E -o '[0-9.]+(ms|s)' | sed 's/\(ms\|s\)$//')

printf "%s:\nnative:%s\ncg    :%s\noreo  :%s\n" "$thread" "$native" "$cg" "$oreo"

relative_native=$(echo "scale=5;$oreo / $native" | bc)
relative_cg=$(echo "scale=5;$oreo / $cg" | bc)

printf "Oreo:native = %s\n" "$relative_native"
printf "Oreo:cg     = %s\n" "$relative_cg"


verbose_echo "Killing executor"
kill -s TERM $executor_pid
verbose_echo "Killing time oracle"
kill $time_oracle_pid

python3 timeoracle-latency-analyze.py