# README

Simple implementation of fault tolerant executor

TODO: deprecate `-w` parameter, put `-db` parameter into config file.

## How to run

```shell
# HTTP registry
go run . -p 8001 --advertise-addr localhost:8001 -registry http -w social -bc config-example.yaml

# Etcd registry
go run . -p 8001 --advertise-addr localhost:8001 -registry etcd -w social -bc config-example.yaml
```

When using the HTTP registry you can supply multiple instances for failover by either
listing them in the configuration (`registry_addrs`) or passing a comma-separated
value to `--registry-addr`.
