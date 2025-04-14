# README

## ft-timeoracle

两个 Time oracle 实例由脚本启动, 但 HAProxy 服务需要手动开启:

```shell
sudo yum install -y haproxy

sudo cp haproxy.cfg /etc/haproxy/haproxy.cfg

sudo systemctl start haproxy

sudo systemctl status haproxy

sudo systemctl restart haproxy

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

# Load data
./ft-full.sh -r

# Run
./ft-full.sh -r -l
```

- 注意第一轮 load 数据时 ft-full.sh 中 ft-executor 部署时没有加 `-l`
