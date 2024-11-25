#!/bin/bash

node1=s1-ljy
PASSWORD=kkkzoz


main(){

    ssh -t $node1 "echo '$PASSWORD' | sudo -S bash -c '\
        bash /home/liujinyi/oreo-ben/start-timeoracle.sh && \
        bash /home/liujinyi/oreo-ben/start-executor.sh -wl ycsb -db Redis,Cassandra
    '"

}

main "$@"