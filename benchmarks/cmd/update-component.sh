#!/bin/bash

echo "updating executor"
bash ../../executor/compile-dev.sh

ehco "updating timeoracle"
bash ../../timeoracle/compile-dev.sh