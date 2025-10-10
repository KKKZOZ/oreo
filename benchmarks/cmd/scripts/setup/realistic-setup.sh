#!/bin/bash

wl=
nodeId=

while [[ "$#" -gt 0 ]]; do
    case $1 in
    -wl | --workload)
        wl="$2"
        shift
        ;;
    -id | --node-id)
        nodeId="$2"
        shift
        ;;
    *)
        echo "Unknown parameter passed: $1"
        exit 1
        ;;
    esac
    shift
done

if [ -z "$wl" ] || [ -z "$nodeId" ]; then
    echo "Error: Missing arguments. Usage: $0 -wl <workload> -id <nodeId>"
    echo "-wl: iot, social, order"
    echo "-id: 2, 3"
    exit 1
fi

# Validate workload
if [[ ! "$wl" =~ ^(iot|hotel|social)$ ]]; then
    echo "Error: Invalid workload. Must be iot, hotel or social"
    exit 1
fi


deploy_mongodb() {
    echo "Remove MongoDB container"
    docker rm -f mongo2
    echo "Create new MongoDB container"
    docker run -d \
        --name mongo2 \
        -p 27018:27017 \
        -e MONGO_INITDB_ROOT_USERNAME=admin \
        -e MONGO_INITDB_ROOT_PASSWORD=password \
        --restart=always \
        mongo
}

deploy_redis() {
    echo "Remove Redis container"
    docker rm -f redis
    echo "Create new Redis container"
    docker run --name redis --network=host --restart=always -d redis redis-server --requirepass "kkkzoz"
    # docker run --name redis -p 6379:6379 --restart=always -d redis redis-server --requirepass "kkkzoz" --save ""
    # docker run --name redis -p 6379:6379 --restart=always -d redis redis-server --requirepass password --save 60 1 --loglevel warning
}

deploy_cassandra() {
    echo "Remove Cassandra container"
    docker rm -f cassandra
    echo "Create new Cassandra container"
    docker run -d --name cassandra -p 9042:9042 cassandra
    if [ ! -f "cassandra_util" ]; then
        echo "ERROR: cassandra_util not found"
        exit 1
    fi
    echo "Waiting for Cassandra to start: 70 seconds"
    sleep 70
    echo "Setup Cassandra"
    ./cassandra_util -op create
}

deploy_kvrocks() {
    echo "Remove KVRocks container"
    docker rm -f kvrocks
    echo "Create new KVRocks container"
    docker run --name kvrocks --restart=always -d -p 6666:6666 apache/kvrocks --bind 0.0.0.0 --requirepass password
}

# Deploy based on workload and nodeId
if [ "$wl" == "iot" ]; then
    if [ "$nodeId" == "2" ]; then
        deploy_redis
    elif [ "$nodeId" == "3" ]; then
        deploy_mongodb
    fi
elif [ "$wl" == "hotel" ]; then
    if [ "$nodeId" == "2" ]; then
        deploy_redis
    elif [ "$nodeId" == "3" ]; then
        deploy_mongodb
    elif [ "$nodeId" == "4" ]; then
        deploy_cassandra
    fi
elif [ "$wl" == "social" ]; then
    if [ "$nodeId" == "2-1" ]; then
        deploy_redis
    elif [ "$nodeId" == "2-2" ]; then
        deploy_kvrocks
    elif [ "$nodeId" == "3" ]; then
        deploy_mongodb
    elif [ "$nodeId" == "4" ]; then
        deploy_cassandra
    fi
fi

# rm -f "../$wl/$wl-load"

echo "Deployment complete for $wl workload on node$nodeId"
