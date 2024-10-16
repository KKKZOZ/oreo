#!/bin/bash

executor_port=8001
timeoracle_port=8010
thread_load=20
tar_dir=iot


if [ $# -lt 1 ]; then
  threads=(8 16 32 64 96)
else
  threads=("$1")
fi


# Clear files
# rm iot-native.txt iot-oreo.txt
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

echo "Starting executor"
./executor -p "$executor_port" -w iot -kvrocks localhost:6666 -mongo1 mongodb://localhost:27018 > /dev/null 2> executor-log.log &
executor_pid=$!

echo "Starting time oracle"
./timeoracle > /dev/null &
time_oracle_pid=$!


if [ ! -f "iot-load" ]; then

  echo "Loading to IOT oreo"
  ben load iot oreo "$thread_load"

  echo "Loading to IOT native"
  ben load iot native "$thread_load"

  touch iot-load
else
  echo "Data has been already loaded"
fi

for thread in "${threads[@]}"
do

  echo "Running IOT native thread=$thread"
  go run . -d native -wl iot -wc ./workloads/iot.yaml -m run -ps native -t "$thread" > "$tar_dir/iot-native-$thread.txt"

  echo "Running IOT Cherry Garcia thread=$thread"
  go run . -d oreo -wl iot -wc ./workloads/iot.yaml -m run -ps oreo -ps cg -t "$thread" > "$tar_dir/iot-cg-$thread.txt"

  echo "Running IOT oreo thread=$thread"
  go run . -d oreo -wl iot -wc ./workloads/iot.yaml -m run -ps oreo -remote -t "$thread" > "$tar_dir/iot-oreo-$thread.txt"

  native=$(sed -n '21p' "$tar_dir/iot-native-$thread.txt" | grep -E -o '[0-9.]+(ms|s)' | sed 's/\(ms\|s\)$//') 
  cg=$(sed -n '21p' "$tar_dir/iot-cg-$thread.txt" | grep -E -o '[0-9.]+(ms|s)' | sed 's/\(ms\|s\)$//') 
  oreo=$(sed -n '21p' "$tar_dir/iot-oreo-$thread.txt" | grep -E -o '[0-9.]+(ms|s)' | sed 's/\(ms\|s\)$//')

  printf "%s:\nnative:%s\ncg:%s\noreo:%s\n" "$thread" "$native" "$cg" "$oreo"

  relative_native=$(echo "scale=5;$oreo / $native" | bc)
  relative_cg=$(echo "scale=5;$oreo / $cg" | bc)

  printf "Oreo:native = %s\n" "$relative_native"
  printf "Oreo:cg = %s\n" "$relative_cg"
  echo "---------------------------------"

done

echo "Killing executor"
kill $executor_pid
echo "Killing time oracle"
kill $time_oracle_pid
