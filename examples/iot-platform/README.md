# IoT Platform

基于 Oreo 分布式事务框架的物联网平台示例，展示了如何在多数据库环境中处理 IoT 设备数据。

# 启动步骤

## 1. 启动数据库服务

```powershell
docker compose -f './docker-compose.yml' up -d --build
```

## 2. 配置 Cassandra

连接到 Cassandra 容器并创建 keyspace：

```powershell
# 连接到 Cassandra 容器
docker exec -it cassandra cqlsh

# 创建 keyspace
CREATE KEYSPACE oreo WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
```

## 3. 启动时间预言机服务

```powershell
./ft-timeoracle -role primary -p 8012 -type hybrid -max-skew 50ms
```

## 4. 启动执行器实例

在不同的终端中分别启动：

```powershell
# Redis 和 MongoDB 执行器
./ft-executor -p 8002 -w ycsb --advertise-addr "localhost:8002" -bc "./executor-config.yaml" -db "Redis,MongoDB1"

# Cassandra 执行器
./ft-executor -p 8003 -w ycsb --advertise-addr "localhost:8003" -bc "./executor-config.yaml" -db "Cassandra"
```

## 5. 启动 IoT 平台服务

```powershell
go run main.go
```

服务将在 `http://localhost:8081` 启动。