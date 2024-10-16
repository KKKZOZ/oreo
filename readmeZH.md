# Oreo: High-Performance and Scalable Transactions across Heterogeneous NoSQL Data Stores

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


<div align="center">

![Logo](./assets/img/logo.png)
</div>
## Oreo

Oreo: High-Performance and Scalable Transactions across Heterogeneous NoSQL Data Stores

This repository is dedicated to sharing the implementation of Oreo for the ASPLOS 2025 paper entitled: Oreo: High-Performance and Scalable Transactions across Heterogeneous NoSQL Data Stores.



## Oreo Structure

![Project Structure](./assets/img/sys-arch.png)



## Project Structure

+ `./benchmarks`：所有关于 Benchmark 测试的代码
+ `./executor`：Stateless Executor 的代码
+ `./integration`：集成测试的代码
+ `./internal`：一些自己使用的内部类
+ `./pkg`：所有关于 Oreo 的代码





## Evaluation

### Command Line Parameters

下面给出 benchmark 应用程序中所有命令参数的说明

你可以切换到`./benchmarks/cmd`路径下， 通过 `go build .` 生成对应的二进制文件，或者通过 `go run .` 直接编译运行

```bash
Usage of ./benchmarks/cmd:
  -d string
        DB type
  -help
        Show help
  -m string
        Mode: load or run (default "load")
  -ps string
        Preset configuration for evaluation
  -read string
        Read Strategy (default "p")
  -remote
        Run in remote mode (for Oreo series)
  -t int
        Thread number (default 1)
  -trace
        Enable trace
  -wc string
        Workload configuration path
  -wl string
        Workload type
```

下面是对每个选项的详细说明

`-d`：必要参数，选取本次测试中运行的数据库类型，可选的类型有：

+ `redis`：原生的 Redis 数据操作
  + 适配的工作负载为 `ycsb`
  + 使用的数据库地址为 `RedisDBAddr`
+ `oreo-redis`：Oreo 框架下的 Redis 数据操作，支持分布式 ACID 事务
  + 适配的工作负载为 `ycsb`
  + 使用的数据库地址为 `OreoRedisAddr`
+ `mongo`：原生的 MongoDB 数据操作
  + 适配的工作负载为 `ycsb`
  + 使用的数据库地址为 `MongoDBAddr1`
+ `oreo-mongo`：Oreo 框架下的 MongoDB 数据操作，支持分布式 ACID 事务
  + 适配的工作负载为 `ycsb`
  + 使用的数据库地址为 `OreoMongoDBAddr1`
+ `native-rm`：原生的数据操作，数据库组合为 Redis-MongoDB
  + 适配的工作负载为 `multi-ycsb`
  + 使用的数据库地址为 `RedisDBAddr` 和 `MongoDBAddr1`
+ `native-mm`：原生的数据操作，数据库组合为 MongoDB-MongoDB
  + 适配的工作负载为 `multi-ycsb`
  + 使用的数据库地址为 `MongoDBAddr1` 和 `MongoDBAddr2`

+ `oreo-rm`：支持分布式 ACID 事务，数据库组合为 Redis-MongoDB
  + 适配的工作负载为 `multi-ycsb`
  + 使用的数据库地址为 ` OreoRedisAddr` 和 `OreoMongoDBAddr1`
+ `oreo-mm`：支持分布式 ACID 事务，数据库组合为 MongoDB-MongoDB
  + 适配的工作负载为 `multi-ycsb`
  + 使用的数据库地址为 ` OreoMongoDBAddr1` 和 `OreoMongoDBAddr2`



`-m` ：必要参数，设定本次执行的模式，可以为 `load` 或者 `run`

+ `load` 指令给对应的数据库类型加载对应数据
+ `run` 指令开启性能基准测试



`-ps`：可选参数，默认值为空，设置预定的参数，具体的参数可以在 `main.go` 中查看

+ `cg`：Cherry Garcia 协议对应的参数，如果想要测试 Cherry Garcia，请务必指定此选项
+ `native`：原生数据操作对应的参数



`-read`：可选参数，默认值为 `p`，设定 Oreo 框架下事务读的策略

+ `p`：采用悲观读的策略
+ `ac`：采用 Assume-Committed 的策略
+ `aa`：采用 Assume-Aborted 的策略



`-remote`：是否使用 Stateless Executor 进行数据操作，Executor 的地址可以在 `./benchmarks/pkg/config` 中配置



`-t`：必要参数，本次测试采用的线程数



`-trace`：是否启用 Golang 自带的 trace 分析



`-wc`：必要参数，本次测试的工作负载配置路径



下面提供了几个命令行的示例：

```bash
# Load data to redis using 100 threads
go run . -d redis -wl ycsb -wc ./workloads/workloada.yaml -m load -t 100
# Load data to oreo-redis using 100 threads
go run . -d oreo-redis -wl ycsb -wc ./workloads/workloada.yaml -m load -t 100

# Running with native-rm under workload A
go run . -d native-rm -wl multi-ycsb -wc ./workloads/workloada.yaml -m run -ps native -t 128
# Running with oreo-mm under workload F
go run . -d oreo-mm -wl multi-ycsb -wc ./workloads/workloadf.yaml -m run -remote -t 128
# Running with oreo-mm under workload F in Cherry Garcia
go run . -d oreo-mm -wl multi-ycsb -wc ./workloads/workloada -m run -ps cg -t 128
```

### Getting Started

本节将介绍如何从零开始运行一个 Benchmark。

第一步，我们需要将整个仓库克隆下来

```bash
git clone git@github.com:oreo-dtx-lab/oreo.git
```

第二步，修改相关的配置

1. 数据库地址，用户名以及密码，位于 `./benchmarks/cmd/main.go` 中
2. Stateless Executor 地址，位于 `./benchmarks/pkg/config/config.go` 中

第三步，编译 Stateless Executor

```bash
cd ./executor
go build .
```

第四步，将 Executor 分发到数据库节点上并运行

> Executor 也有几个简单的命令行参数：
>
> + `-p`：设置 HTTP 服务器的端口
> + `-r1`：指定 `redis1` 数据库的地址
> + `-m1`：指定 `mongo1` 数据库的地址
> + `-m2`：指定 `mongo2` 数据库的地址

```bash
./executor -m1 mongodb://localhost:27017 -m2 mongodb://localhost:27018
```

第五步，确定自己要运行的工作负载，检查其配置，工作负载的配置文件在 `./benchmarks/cmd/workloads` 路径下，采用 yaml 格式编写，下面是一份工作负载的模版

```yaml
# Total number of records to be generated
recordcount: 1000000

# Total number of operations to be performed
operationcount: 100000

# Number of operations in each transaction group
txnoperationgroup: 6

# Proportions of different operations
readproportion: 0.5 # Proportion of read operations
updateproportion: 0.5 # Proportion of update operations
insertproportion: 0 # Proportion of insert operations
scanproportion: 0 # Proportion of scan operations
readmodifywriteproportion: 0 # Proportion of read-modify-write operations

# Proportions of operations on different databases
redis1proportion: 0.5 # Proportion of operations on Redis instance 1
mongo1proportion: 0.5 # Proportion of operations on MongoDB instance 1
mongo2proportion: 0 # Proportion of operations on MongoDB instance 2
couchdbproportion: 0 # Proportion of operations on CouchDB
```

第六步，向数据库中加载数据，注意不同类型的数据库对应的地址是不一样的

> 注意，数据库组合类型如 `oreo-mm`，`oreo-rm`，`native-mm`，`native-rm` 不支持直接加载数据，请手动分为两个数据库独立运行加载指令，如 `oreo-rm` 请运行
>
> ```bash
> # Load data to redis
> go run . -d oreo-redis -wl ycsb -wc ./workloads/workloada.yaml -m load -t 100
> # Load data to mongo
> go run . -d oreo-mongo -wl ycsb -wc ./workloads/workloada.yaml -m load -t 100
> ```



```bash
# Load data to redis using 100 threads
go run . -d redis -wl ycsb -wc ./workloads/workloada.yaml -m load -t 100
# Load data to oreo-redis using 100 threads
go run . -d oreo-redis -wl ycsb -wc ./workloads/workloada.yaml -m load -t 100
```

第七步，运行 Benchmark，下面提供了一些例子以供参考

```bash
# Running with native-rm under workload A
go run . -d native-rm -wl multi-ycsb -wc ./workloads/workloada.yaml -m run -ps native -t 128
# Running with oreo-mm under workload F
go run . -d oreo-mm -wl multi-ycsb -wc ./workloads/workloadf.yaml -m run -remote -t 128
# Running with oreo-mm under workload F in Cherry Garcia
go run . -d oreo-mm -wl multi-ycsb -wc ./workloads/workloada.yaml -m run -ps cg -t 128
# Running with oreo-redis under purewrite workload using Assume-Abort strategy
go run . -d oreo-redis -wl ycsb -wc ./workloads/purewrite.yaml -m run -remote -read aa -t 128
```

第八步，等待 Benchmark 完成，你应该会得到类似于下面的输出：

```bash
-----------------
DBType: oreo-mm
Mode: run
WorkloadType: multi-ycsb
ThreadNum: 128
Remote Mode: true
Read Strategy: p
ConcurrentOptimizationLevel: 2
AsyncLevel: 2
MaxOutstandingRequest: 5
MaxRecordLength: 3
HTTPAdditionalLatency: 3ms ConnAdditionalLatency: 0s
LeaseTime: 100ms
ZipfianConstant: 0.9
-----------------
Start to run benchmark
----------------------------------
Run finished, takes 9.403273025s
COMMIT - Takes(s): 9.4, Count: 9528, OPS: 1017.9, Avg(us): 25863, Min(us): 8296, Max(us): 100223, 50th(us): 24175, 90th(us): 37567, 95th(us): 42975, 99th(us): 56383, 99.9th(us): 77183, 99.99th(us): 99391
COMMIT_ERROR - Takes(s): 9.4, Count: 7112, OPS: 759.9, Avg(us): 20556, Min(us): 4516, Max(us): 91903, 50th(us): 19151, 90th(us): 31583, 95th(us): 35839, 99th(us): 47615, 99.9th(us): 65535, 99.99th(us): 87999
READ   - Takes(s): 9.4, Count: 97334, OPS: 10355.1, Avg(us): 7597, Min(us): 6, Max(us): 85631, 50th(us): 5951, 90th(us): 13351, 95th(us): 16975, 99th(us): 25935, 99.9th(us): 39359, 99.99th(us): 57951
READ_ERROR - Takes(s): 9.3, Count: 2666, OPS: 285.3, Avg(us): 12612, Min(us): 3968, Max(us): 79039, 50th(us): 9943, 90th(us): 23375, 95th(us): 29071, 99th(us): 40735, 99.9th(us): 60511, 99.99th(us): 79039
Start  - Takes(s): 9.4, Count: 16768, OPS: 1782.8, Avg(us): 58, Min(us): 20, Max(us): 5515, 50th(us): 32, 90th(us): 62, 95th(us): 82, 99th(us): 593, 99.9th(us): 2723, 99.99th(us): 4759
TOTAL  - Takes(s): 9.4, Count: 247132, OPS: 26280.6, Avg(us): 11487, Min(us): 1, Max(us): 186879, 50th(us): 4119, 90th(us): 50623, 95th(us): 70079, 99th(us): 91391, 99.9th(us): 115391, 99.99th(us): 139519
TXN    - Takes(s): 9.4, Count: 9528, OPS: 1017.8, Avg(us): 71551, Min(us): 31552, Max(us): 166399, 50th(us): 69695, 90th(us): 91135, 95th(us): 98687, 99th(us): 114687, 99.9th(us): 136447, 99.99th(us): 153343
TXN_ERROR - Takes(s): 9.4, Count: 7112, OPS: 760.0, Avg(us): 68862, Min(us): 27232, Max(us): 157567, 50th(us): 66815, 90th(us): 88959, 95th(us): 96831, 99th(us): 113663, 99.9th(us): 134527, 99.99th(us): 155775
TxnGroup - Takes(s): 9.4, Count: 16640, OPS: 1773.8, Avg(us): 70293, Min(us): 22240, Max(us): 186879, 50th(us): 68607, 90th(us): 90623, 95th(us): 98559, 99th(us): 115327, 99.9th(us): 140415, 99.99th(us): 162431
UPDATE - Takes(s): 9.4, Count: 97334, OPS: 10353.6, Avg(us): 7, Min(us): 1, Max(us): 3601, 50th(us): 4, 90th(us): 6, 95th(us): 7, 99th(us): 17, 99.9th(us): 1314, 99.99th(us): 2675
Error Summary:

  Operation:  COMMIT
       Error   Count
       -----   -----
prepare phase failed: Remote prepare failed
  version mismatch  6281
prepare phase failed: Remote prepare failed
rollForward failed
  version mismatch  515
prepare phase failed: Remote prepare failed
rollback failed
  version mismatch  315
prepare phase failed: Remote prepare failed
  read failed due to unknown txn status  1

                             Operation:   READ
                                  Error  Count
                                  -----  -----
  read failed due to unknown txn status   1333
rollForward failed
  version mismatch  826
rollback failed
  version mismatch  507
```

测试的输出主要分为三个部分：

+ 本次测试的配置信息

+ 性能测量信息
  + 这部分统计了测试中每个操作的用时数据，包括平均值/最大值/最小值/50分位/95分位/99分位/99.9分位/99.99分位的数据

+ 错误记录信息
  + 这部分统计了本次测试中所有事务操作出错的次数和原因



### Note

+ 如果需要测试 Cherry Garcia，请使用 `-ps cg` 参数
+ 程序默认的模拟延迟为 3ms，如果你需要调整模拟延迟，请修改 `./benchmark/pkg/config/config.go` 中的数值













## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.