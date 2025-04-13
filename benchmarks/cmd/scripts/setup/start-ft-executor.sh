#!/bin/bash

##------------LOGGING------------##
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log() {
    local color=${2:-$NC}
    if [[ "${verbose}" = true ]]; then
        echo -e "${color}$1${NC}"
    fi
}
##------------LOGGING------------##

executor_port=8001
db_combinations=
wl_mode=
verbose=false

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
    -p | --port)
        executor_port="$2"
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

if [ $wl_mode == "ycsb" ]; then
    bc=./config/BenConfig_ycsb.yaml
else
    bc=./config/BenConfig_realistic.yaml
fi

kill_process_on_port() {
    local port=$1
    local pid
    pid=$(lsof -t -i ":$port")
    if [ -n "$pid" ]; then
        echo "Port $port is occupied by process $pid. Terminating this process..."
        kill -9 "$pid"
    fi
}

main() {

    cd "$(dirname "$0")"

    if [ ! -f "$bc" ]; then
        echo "Error: Benchmark Configuration file not found at $bc"
        exit 1
    fi

    if [ -z "$wl_mode" ]; then
        echo "Error: Workload mode must be specified using -wl or --workload"
        exit 1
    fi

    if [ ! -f "getip" ]; then
        echo "Error: getip binary not found. Please build it first."
        exit 1
    fi

    ip=$(./getip)

    kill_process_on_port "$executor_port"
    echo "Starting executor"
    nohup ./ft-executor -p "$executor_port" --advertise-addr "$ip" -w "$wl_mode" -bc "$bc" -db "$db_combinations" 2>./executor.log &

    sleep 3
    echo "Executor started"
    lsof -i ":$executor_port"
}

main "$@"
