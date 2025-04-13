# README

## ft-timeoracle

两个 Time oracle 实例由脚本启动, 但 HAProxy 服务需要手动开启:

```shell
sudo yum install haproxy

sudo cp haproxy.cfg /etc/haproxy/haproxy.cfg

sudo systemctl start haproxy

sudo systemctl status haproxy

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
# node2
./ycsb-setup.sh Redis
# node3
./ycsb-setup.sh MongoDB1
# node4
./ycsb-setup.sh Cassandra

./ft-full.sh -r
```
