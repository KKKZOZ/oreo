#!/bin/bash

db=$1

if [ -z "$db" ]; then
    echo "Error: No database specified."
    exit 1
fi

if [ "$db" == "Redis" ]; then
    echo "Remove redis container"
    docker rm -f redis
    echo "Create new redis container"
    docker run --name redis --network=host --restart=always -d redis redis-server --requirepass "kkkzoz"
    # docker run --name redis -p 6379:6379 --restart=always -d redis redis-server --requirepass "kkkzoz"
    # docker run --name redis -p 6379:6379 --restart=always -d redis redis-server --requirepass password --save 60 1 --loglevel warning
elif [ "$db" == "MongoDB1" ]; then
    echo "Remove mongoDB1 container"
    docker rm -f mongo1
    echo "Create new mongoDB1 container"
    docker run -d \
        --name mongo1 \
        -p 27017:27017 \
        -e MONGO_INITDB_ROOT_USERNAME=admin \
        -e MONGO_INITDB_ROOT_PASSWORD=password \
        --restart=always \
        -v /data/mongo1_data:/data/db \
        mongo
elif [ "$db" == "MongoDB2" ]; then
    echo "Remove mongoDB2 container"
    docker rm -f mongo2
    echo "Create new mongoDB2 container"
    docker run -d \
        --name mongo2 \
        -p 27018:27017 \
        -e MONGO_INITDB_ROOT_USERNAME=admin \
        -e MONGO_INITDB_ROOT_PASSWORD=password \
        --restart=always \
        -v /data/mongo2_data:/data/db \
        mongo
elif [ "$db" == "KVRocks" ]; then
    echo "Remove kv-rocks container"
    docker rm -f kvrocks
    echo "Create new kv-rocks container"
    docker run --name kvrocks --restart=always -d -p 6666:6666 -v /data/kvrocks_data:/data apache/kvrocks --bind 0.0.0.0 --requirepass password
elif [ "$db" == "CouchDB" ]; then
    echo "Remove couchDB container"
    docker rm -f couch
    echo "Create new couchDB container"
    docker run -d --name couch --restart=always -e COUCHDB_USER=admin -e COUCHDB_PASSWORD=password -p 5984:5984 -v /data/couchdb_data:/opt/couchdb/data \ -d couchdb
elif [ "$db" == "Cassandra" ]; then
    echo "Remove cassandra container"
    docker rm -f cassandra
    echo "Create new cassandra container"
    # docker run -d --name cassandra -p 9042:9042 cassandra
    docker run -d \
        --name cassandra --network=host \
        -v /data/cassandra_data:/var/lib/cassandra/data \
        -v /data/cassandra_commitlog:/var/lib/cassandra/commitlog \
        -v /data/cassandra_saved_caches:/var/lib/cassandra/saved_caches \
        cassandra

    if [ ! -f "cassandra_util" ]; then
        echo "ERROR: cassandra_util not found"
        exit 1
    fi
    echo "Waiting for cassandra to start: 70 seconds"
    sleep 70
    echo "Setup cassandra"
    ./cassandra_util -op create
elif [ "$db" == "TiKV" ]; then
    echo "Restart TiKV binary"
    ./deploy-tikv.sh

else
    echo "Invalid database"
fi

# rm ../ycsb/ycsb-load

echo "YCSB setup complete"
