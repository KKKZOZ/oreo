#!/bin/bash

# nodes=(s1-ljy s3-ljy)
nodes=(node2 node3)

main() {
    cd "$(dirname "$0")" && cd ../..
    mkdir -p ./oreo-ben/config

    cp ./scripts/setup/* ./oreo-ben/
    cp ./config/* ./oreo-ben/config
    cp ./bin/* ./oreo-ben

    tar -czf ./oreo-ben.tar.gz ./oreo-ben

    for node in "${nodes[@]}"; do
        scp ./oreo-ben.tar.gz "$node":~
        ssh -t "$node" "rm -rf ~/oreo-ben"
        ssh -t "$node" "tar -xzf ~/oreo-ben.tar.gz && rm ~/oreo-ben.tar.gz"
    done

    rm ./oreo-ben.tar.gz && rm -rf ./oreo-ben

    # for node in "${nodes[@]}"; do
    #     echo "Updating $node"
    #     ssh -t $node "mkdir -p ~/oreo-ben/config"

    #     echo "Updating executor and timeoracle"
    #     scp ./bin/executor ./bin/timeoracle $node:~/oreo-ben
    #     scp ./bin/oreo-executor-image.tar.gz $node:~/oreo-ben

    #     scp ./scripts/start-executor.sh ./scripts/start-executor-docker.sh ./scripts/start-timeoracle.sh $node:~/oreo-ben
    #     scp -r ./config/* $node:~/oreo-ben/config

    #     echo "Updating scripts"
    #     scp ./scripts/ycsb-setup.sh ./scripts/realistic-setup.sh ./scripts/opt-setup.sh ./scripts/read-setup.sh $node:~/oreo-ben
    #     scp ./scripts/cassandra_util $node:~/oreo-ben
    # done
}

main "$@"
