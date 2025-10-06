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

show_usage() {
    echo "Usage: $0 [type]"
    echo "  type: timeoracle type (default: hybrid)"
    echo "Example:"
    echo "  $0           # Use default type: hybrid"
    echo "  $0 hybrid    # Explicitly set type to hybrid"
    echo "  $0 physical  # Set type to physical"
}

main() {
    cd "$(dirname "$0")"
    
    # Parse type parameter, default to "hybrid" if not provided
    local type="${1:-hybrid}"
    
    # Show help if requested
    if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
        show_usage
        exit 0
    fi
    
    kill_process_on_port "$timeoracle_port"
    
    echo "Starting timeoracle with type: $type"
    nohup ./timeoracle -p "$timeoracle_port" -type "$type" >/dev/null 2>./timeoracle.log &
    
    sleep 1
    
    echo "Timeoracle started"
    lsof -i ":$timeoracle_port"
}

main "$@"