#!/bin/bash

nodes=(s1-ljy s3-ljy)


main(){
    cd "$(dirname "$0")" && cd ..

    for node in "${nodes[@]}"; do
        echo "Updating $node"
        echo "Updating executor and timeoracle"
        scp ./bin/executor ./bin/timeoracle $node:~/oreo-ben
        scp ./scripts/start-executor.sh ./scripts/start-timeoracle.sh $node:~/oreo-ben

        echo "Updating scripts"
        scp ./scripts/ycsb-setup.sh ./scripts/realistic-setup.sh $node:~/oreo-ben
        scp ./scripts/cassandra_util $node:~/oreo-ben
    done
}

main "$@"