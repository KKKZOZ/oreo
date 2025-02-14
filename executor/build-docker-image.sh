#!/bin/bash

echo "Building the Go executable..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o executor

echo "Building the Docker image..."
docker build -t oreo-executor .

echo "Saving the Docker image..."
docker save oreo-executor:latest | gzip >oreo-executor-image.tar.gz
mv oreo-executor-image.tar.gz ../benchmarks/cmd/bin/

echo "Cleaning up..."
rm executor
