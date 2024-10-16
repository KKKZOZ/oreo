## 安装 Docker

dnf config-manager --add-repo=https://mirrors.cloud.tencent.com/docker-ce/linux/centos/docker-ce.repo

dnf list docker-ce

dnf install -y docker-ce --nobest

systemctl start docker

docker info

### Docker 命令

docker ps --format "table {{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"

docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"

vim /etc/docker/daemon.json

{
  "registry-mirrors": ["https://mirror.ccs.tencentyun.com"]
}

systemctl restart docker


## 启动数据库

docker run  --name redis-native -p 6379:6379 --restart=always -d redis redis-server --requirepass @ljy123456 --save 60 1 --loglevel warning

docker run --name redis-oreo -p 6380:6379 --restart=always -d redis redis-server --requirepass @ljy123456 --save 60 1 --loglevel warning

docker run -d \
  --name mongo-native \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=admin \
  --restart=always \
  mongo

docker run -d \
  --name mongo-oreo \
  -p 27018:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=admin \
  --restart=always \
  mongo

docker run -d -p 5432:5432 --restart=always --name="apiary-postgres" --env POSTGRES_PASSWORD=dbos --env POSTGRES_MAX_CONNECTIONS=20000 postgres:latest




## 安装 Golang

wget https://go.dev/dl/go1.22.3.linux-amd64.tar.gz

rm -rf /usr/local/go && tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz

vim .bashrc

export PATH=$PATH:/usr/local/go/bin

source .bashrc

go env -w GOPROXY=https://goproxy.cn,direct


## 安装 Java

tar -xzvf jdk-17_linux-x64.tar.gz

export JAVA_HOME=/root/jdk-17.0.11

## 配置 ssh

cat id_rsa.pub.windows >> .ssh/authorized_keys

yum install htop

ssh root@124.223.5.240


---

go run . -d oreo-redis -wl ycsb -t 10 -m load -wc ./workloads/workloada


go run . -d oreo-redis -wl ycsb -wc ./workloads/workloada -m run -t 64 -ps cg

go run . -d oreo-redis -wl ycsb -wc ./workloads/workloadg -m run -t 64 -ps cg


## 常用命令

lsof -i :6379

tar -xzvf oreo.tar.gz

### 步骤

在 Node 1 中安装 golang,传输 oreo，在 /benchmark/cmd/main.go 中配置 Addr

然后本地编译 component

通过 WinSCP 传给其他两个节点

记得 Load 数据，修改 RemoteAddressList

## Topology

go run main.go -p 8000 -r1 localhost:6380 -m1 mongodb://localhost:27018

### Node1
124.223.5.240
172.17.16.6

ssh root@124.223.5.240


#### Component 启动配置

go run main.go -p 8000 -r1 10.206.0.3:6380 -m1 mongodb://10.206.0.4:27018

go run main.go -p 8001 -m1 mongodb://172.17.0.6:27018 -m2 mongodb://172.17.0.14:27018

### DB1
119.45.235.108
172.17.16.2

+ 在 Redis-Mongo 中，作为 Redis1
+ 在 Mongo-Mongo 中，作为 Mongo1
+ 在 Epoxy 中，作为 Postgres 和 Mongo1
Redis1 6379,6380

Mongo1 27017,27018

#### Component 启动配置


+ For Redis-Mongo
./component -p 8000 -r1 localhost:6380 -m1 mongodb://172.17.16.8:27018
+ For Mongo-Mongo
./component -p 8000 -m1 mongodb://localhost:27018 -m2 mongodb://172.17.16.8:27018

### DB2
119.45.202.48
172.17.16.8

+ 在 Redis-Mongo 中，作为 Mongo1
Mongo1 27017,27018
+ 在 Mongo-Mongo 中，作为 Mongo2

#### Component 启动配置

+ For Redis-Mongo
./component -p 8000 -r1 172.17.16.2:6380 -m1 mongodb://localhost:27018
+ For Mongo-Mongo
./component -p 8000 -m1 mongodb://172.17.16.2:27018 -m2 mongodb://localhost:27018





# 实验

## Load 数据

go run . -d redis -wl ycsb -wc ./workloads/workloada.yaml -m load -t 100
go run . -d oreo-redis -wl ycsb -wc ./workloads/workloada.yaml -m load -t 100

go run . -d mongo -wl ycsb -wc ./workloads/workloada.yaml -m load -t 100
go run . -d oreo-mongo -wl ycsb -wc ./workloads/workloada.yaml -m load -t 100

在给 Mongo2 Load 数据时，需要特别注意

### Epoxy 

129.211.181.125

POST 129.211.181.125:7081/load?recordCount=1000000&threadNum=100&mongoIdx=1

POST 129.211.181.125:7081/load?recordCount=1000000&threadNum=100&mongoIdx=2

## Performance of Distributed Transactions

### Workload F

#### Epoxy

129.211.181.125:7081/workloadf?operationCount=100000&threadNum=64

#### Cherry Garcia

go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -ps cg -t 128


go run . -d oreo-mm -wl multi-ycsb -wc ./workloads/workloada -m run -ps cg -t 128

#### Oreo

go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -t 64


go run . -d oreo-mm -wl multi-ycsb -wc ./workloads/workloada -m run -t 8

#### Oreo-Remote

go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -remote -t 8


go run . -d oreo-mm -wl multi-ycsb -wc ./workloads/workloada -m run -remote -t 128

#### Native

go run . -d native-rm -wl multi-ycsb -wc ./workloads/workloada -m run -ps native -t 128


go run . -d native-mm -wl multi-ycsb -wc ./workloads/workloada -m run -ps native -t 128

## Effectiveness of Optimizations

### Performance Optimization

#### C0A1

go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -read p -ps cg -t 64

#### C0A2

go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -read p -ps cg -t 64


#### C1A1

go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -remote -t 64

#### C2A1

go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -remote -t 64

#### C2A2

go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -remote -t 64

### Protocol Optimization

#### Cherry Garcia
go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -read p -ps cg -t 128

#### Oreo-P using remote
go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -remote -read p -t 128

#### Oreo-AA using remote
go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -read aa -remote -t 128

#### Oreo-AC using remote
go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -read ac -remote -t 128


## High Latency

### Cherry Garcia

go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -ps cg -t 128


### Oreo Remote

go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -read p -remote -t 128

## High Write

### Cherry Garcia

go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -ps cg -t 64

### Oreo

go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -read p -remote -t 64


## Scalability

go run . -d oreo-mm -wl multi-ycsb -wc ./workloads/workloada -m run -read p -remote -t 128

go run . -d native-mm -wl multi-ycsb -wc ./workloads/workloada -m run -ps native -t 128

--------

# Dev Test

go run . -d oreo-redis -wl ycsb -wc ./workloads/workloada -m run -t 128

go run . -d oreo-mongo -wl ycsb -wc ./workloads/workloada -m run -t 128

go run . -d oreo-mm -wl multi-ycsb -wc ./workloads/workloada -m run -read p -t 128


### Single Test

#### Load

go run . -d redis -wl ycsb -wc ./workloads/workloada -m load -t 100

go run . -d oreo-redis -wl ycsb -wc ./workloads/workloada -m load -t 100

go run . -d mongo -wl ycsb -wc ./workloads/workloada -m load -t 100

go run . -d oreo-mongo -wl ycsb -wc ./workloads/workloada -m load -t 100

#### Run

go run . -d oreo-redis -wl ycsb -wc ./workloads/workloada -m run -t 64

go run . -d oreo-mongo -wl ycsb -wc ./workloads/workloada -m run -t 64


### rm 系列

go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -t 1

go run . -d oreo-rm -wl multi-ycsb -wc ./workloads/workloada -m run -t 64

go run . -d native-rm -wl multi-ycsb -wc ./workloads/workloada -m run -t 64


