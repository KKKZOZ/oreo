#!/bin/bash

nodeId=

while [[ "$#" -gt 0 ]]; do
    case $1 in
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

deploy_mongodb2() {
    echo "Remove mongoDB2 container"
    docker rm -f mongo2
    echo "Create new mongoDB2 container"
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
    docker run --name redis -p 6379:6379 --restart=always -d redis
    # docker run --name redis -p 6379:6379 --restart=always -d redis redis-server --requirepass password --save 60 1 --loglevel warning
}

main(){
    if [ "$nodeId" == "2" ]; then
        deploy_redis
    elif [ "$nodeId" == "3" ]; then
        deploy_mongodb2
    fi
}

main "$@"