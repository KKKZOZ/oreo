#!/bin/bash

timeoracle_port=8010

kill_process_on_port() {
    local port=$1
    local pid
    pid=$(lsof -t -i ":$port")
    if [ -n "$pid" ]; then
        echo "Port $port is occupied by process $pid. Terminating this process..."
        kill -9 "$pid"
    fi
}

main(){

    cd "$(dirname "$0")"

    kill_process_on_port "$timeoracle_port"
    echo "Starting timeoracle"
    nohup ./timeoracle -p "$timeoracle_port" -type hybrid >/dev/null 2>./timeoracle.log &

    sleep 1
    echo "Timeoracle started"
    lsof -i ":$timeoracle_port"
}

main "$@"
