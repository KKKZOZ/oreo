#!/bin/bash

timeoracle_primary_port=8010
timeoracle_backup_port=8011
max_skew=1s

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

    kill_process_on_port "$timeoracle_primary_port"
    kill_process_on_port "$timeoracle_backup_port"

    echo "Starting primary timeoracle"

    nohup ./ft-timeoracle -role primary -p 8010 -type hybrid -max-skew "$max_skew" >/dev/null 2>./primary_timeoracle.log &

    sleep 1
    echo "Primary timeoracle started"
    lsof -i ":$timeoracle_primary_port"

    echo "Starting backup timeoracle"

    nohup ./ft-timeoracle -role backup -p 8011 -type hybrid -max-skew "$max_skew" \
        -primary-addr http://localhost:"$timeoracle_primary_port" \
        -health-check-interval 0.25s \
        -health-check-timeout 0.25s \
        -failure-threshold 1 \
        >/dev/null 2>./backup_timeoracle.log &

    sleep 1
    echo "Backup timeoracle started"
    lsof -i ":$timeoracle_backup_port"
}

main "$@"
