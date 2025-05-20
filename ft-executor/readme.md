# README

Simple implementation of fault tolerant executor

## How to run

```shell
go run . -p 8001 --advertise-addr localhost:8001 --registry-addr http://localhost:9000 -w social -bc config.yaml

go run . -p 8001 --advertise-addr localhost:8001 -w social -bc config.yaml
```
