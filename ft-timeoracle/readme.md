# README

A simple implementation of primary-backup time oracle

## How to run

```shell
go run . -role primary -p 8010 -type hybrid -max-skew 50ms

go run . -role backup -p 8011 -type hybrid -max-skew 50ms \
             -primary-addr http://localhost:8010 \
             -health-check-interval 2s \
             -health-check-timeout 1s \
             -failure-threshold 3

```
