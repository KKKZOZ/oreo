#!/bin/bash

docker rm -f redis

docker rm -f mongodb

docker run --name redis -p 6379:6379 --restart=always -d redis

docker run -d \
    --name mongodb \
    -p 27017:27017 \
    -e MONGO_INITDB_ROOT_USERNAME=admin \
    -e MONGO_INITDB_ROOT_PASSWORD=password \
    --restart=always \
    mongo
