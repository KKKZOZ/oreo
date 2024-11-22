#!/bin/sh

node2=s1-ljy
node3=s3-ljy

cd "$(dirname "$0")"

bash update-component.sh

cd ..

echo "Create directory on Node 2"
ssh -n $node2 "mkdir oreo-ben"

echo "Create directory on Node 3"
ssh -n $node3 "mkdir oreo-ben"

echo "Send executor and timeoracle to Node 2"
scp ./bin/executor ./bin/timeoracle $node2:~/oreo-ben

echo "Send executor and timeoracle to Node 3"
scp ./bin/executor ./bin/timeoracle $node3:~/oreo-ben
