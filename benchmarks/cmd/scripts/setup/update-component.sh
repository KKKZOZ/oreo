#!/bin/bash

handle_error() {
    echo "Error: $1"
    exit 1
}

# cd "$(dirname "$0")" && cd ..

# echo "Building benchmarks..."

# go build . || handle_error "Failed to build benchmarks"

# mv cmd ./bin || handle_error "Failed to move cmd to bin"

PROJECT_ROOT="$(cd "$(dirname "$0")" && cd ../../../.. && pwd)"

mkdir -p "${PROJECT_ROOT}/benchmarks/cmd/bin"


components=("executor" "timeoracle")

for component in "${components[@]}"; do
    echo "Updating ${component}..."

    cd "${PROJECT_ROOT}/${component}" || handle_error "Failed to enter ${component} directory"

    if ! bash compile-dev.sh; then
        handle_error "Failed to compile ${component}"
    fi

    if [[ "$component" == "executor" ]]; then
        bash build-docker-image.sh
    fi

    echo "Successfully updated ${component}"
    echo
done

echo "Updating cassandra util"

cd "${PROJECT_ROOT}/pkg/datastore/cassandra/cmd/util" || handle_error "Failed to enter cassandra util directory"

if ! bash compile-dev.sh; then
    handle_error "Failed to compile cassandra util"
fi

echo "Successfully updated cassandra util"

echo "All components updated successfully"
