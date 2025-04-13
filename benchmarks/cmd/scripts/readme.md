# README

## ft-timeoracle

两个 Time oracle 实例由脚本启动, 但 HAProxy 服务需要手动开启:

```shell
haproxy -f ./haproxy.cfg
```

## ft-executor

实验时需要手动停止和启动 `oreo-ft-executor` container

```shell
docker stop ft-executor-8002

docker start ft-executor-8002

```

## ft-full.sh

```shell
./ft-full.sh -r
```
