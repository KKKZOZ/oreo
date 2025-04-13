#!/bin/bash

echo "Building the Go executable..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o ft-executor

echo "Building the Docker image..."
docker build -t oreo-ft-executor .

echo "Saving the Docker image..."
docker save oreo-ft-executor:latest | gzip >oreo-ft-executor-image.tar.gz
mv oreo-ft-executor-image.tar.gz ../benchmarks/cmd/bin/

echo "Cleaning up..."
rm ft-executor
