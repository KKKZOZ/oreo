#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

node_list=(node2 node3 node4)

log() {
    local color=${2:-$NC}
    echo -e "${color}$1${NC}"
}

main() {
    log "Start Fault Tolerance Process" "$GREEN"

    sleep 10
    for node in "${node_list[@]}"; do
        log "Stopping ft-executor-8002 on $node" "$GREEN"
        ssh -t "$node" "docker stop ft-executor-8002"
        sleep 0.3
    done

    sleep 6
    for node in "${node_list[@]}"; do
        log "Starting ft-executor-8002 on $node" "$GREEN"
        ssh -t "$node" "docker start ft-executor-8002"
        sleep 0.3
    done

    sleep 8

    log "Stopping primary timeoracle on ${node_list[0]}" "$GREEN"
    ssh -t "${node_list[0]}" "pkill -f 'ft-timeoracle -role primary'"

    sleep 8
    log "Stopping MongoDB1 on ${node_list[1]}" "$GREEN"
    ssh -t "${node_list[1]}" "docker rm -f mongo1"

    sleep 5
    log "Starting MongoDB1 on ${node_list[1]}" "$GREEN"
    ssh -t "${node_list[1]}" "bash /root/oreo-ben/ycsb-setup.sh MongoDB1"

    sleep 2
    log "Finish Fault Tolerance Process" "$GREEN"
}

main "$@"
