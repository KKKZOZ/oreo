#!/bin/bash

HOST=$1

rsync -avP root@$HOST:~/oreo/benchmarks/cmd/data/ ~/Projects/oreo/benchmarks/cmd/data/

# rsync -avP root@$HOST:~/oreo ~/Projects/oreo2
