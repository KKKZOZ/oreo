# Benchmakr Note

## 10-21

- 把 iot-setup 中的 docker container 改为 redis 进行测试
- iot-workload 代码中保持不变，我们先把 OreoKVRocksAddr 改为 Redis 的地址进行测试
    - Bash 脚本中 KVRocks 的地址也需要暂时改为 Redis 的
- 启动独立的中心化 Time Oracle，记得确定地址和配置是否正确
    - 采用 Hybrid，physicalTimeInterval 为 10ms
- 确定一下 Benchmark 的 Config 是否符合预期
    - AdditionalLatency: 5ms 应该没有问题
- 每轮脚本跑完后记得检查一下错误率，确保整个执行过程是正常的

### Time Oracle 测试

- 切换 TimeOracle 类型后记得 ./iot-setup
- cg 的 TimeOracle 类型必须统一
- 测试 hybrid 在 physicalInterval = 1 and 10 的情况


physicalInterval = 10

```bash
❯ bash iot-analyze-single.sh 8
Starting executor
Starting time oracle
Data has been already loaded
8:
native:8.28436328
cg    :20.96166045
oreo  :9.66813186
Oreo:native = 1.16703
Oreo:cg     = .46122
Killing executor
Killing time oracle
   Max Latency (µs)  Min Latency (µs)  Average Latency (µs)  Total Count
0           698.796             1.002              2.750021        16104
❯ bash iot-analyze-single.sh 16 -v
16:
native:4.33229704
cg    :10.38561054
oreo  :4.99619336
Oreo:native = 1.15324
Oreo:cg     = .48106
   Max Latency (µs)  Min Latency (µs)  Average Latency (µs)  Total Count
0           605.814             1.001              2.586606        16029
❯ bash iot-analyze-single.sh 32 -v
32:
native:2.24876970
cg    :5.31330361
oreo  :2.91972948
Oreo:native = 1.29836
Oreo:cg     = .54951
   Max Latency (µs)  Min Latency (µs)  Average Latency (µs)  Total Count
0           620.613               1.0              2.306105        15743
❯ bash iot-analyze-single.sh 64 -v
64:
native:1.13969009
cg    :2.76668489
oreo  :2.22907627
Oreo:native = 1.95586
Oreo:cg     = .80568
   Max Latency (µs)  Min Latency (µs)  Average Latency (µs)  Total Count
0           188.663               1.0              2.093707        15628
❯ bash iot-analyze-single.sh 96 -v
96:
native:0.79483884
cg    :2.04855043
oreo  :2.22632751
Oreo:native = 2.80097
Oreo:cg     = 1.08678
   Max Latency (µs)  Min Latency (µs)  Average Latency (µs)  Total Count
0           214.773               1.0              2.030376        15416
```

- 跑一次 local simple

```bash
8:
native:20.73855582
cg    :53.38732254
oreo  :23.12912925
Oreo:native = 1.11527
Oreo:cg     = .43323
   Max Latency (µs)  Min Latency (µs)  Average Latency (µs)  Total Count
0            23.826             1.001              2.566441          145
❯ bash iot-analyze-single.sh 16 -lv
16:
native:10.72302435
cg    :25.36335983
oreo  :11.90949951
Oreo:native = 1.11064
Oreo:cg     = .46955
   Max Latency (µs)  Min Latency (µs)  Average Latency (µs)  Total Count
0            21.736               1.0              2.588164          122
❯ bash iot-analyze-single.sh 32 -lv
32:
native:5.61391675
cg    :12.76052492
oreo  :6.76155421
Oreo:native = 1.20442
Oreo:cg     = .52988
   Max Latency (µs)  Min Latency (µs)  Average Latency (µs)  Total Count
0            41.597             1.001              2.587135          185
❯ bash iot-analyze-single.sh 64 -lv
64:
native:2.72449313
cg    :6.63818913
oreo  :4.79146287
Oreo:native = 1.75866
Oreo:cg     = .72180
   Max Latency (µs)  Min Latency (µs)  Average Latency (µs)  Total Count
0           533.614             1.004              7.756182          121
❯ bash iot-analyze-single.sh 96 -lv
96:
native:1.87952274
cg    :4.83403319
oreo  :4.44653516
Oreo:native = 2.36577
Oreo:cg     = .91983
   Max Latency (µs)  Min Latency (µs)  Average Latency (µs)  Total Count
0            24.353             1.006              2.093827          231
```


可以发现，32 -> 64 之后额外开销立刻变大：

- 个人感觉是服务器性能问题
   - 把数据库分隔开试试
      - 在 s1 中部署数据库
      - 防火墙问题，需要查看阿里云控制台
      - 顺便修改所有的脚本，把端口统一到某个区间中，遵循 port_allcation 的规则
   - 本地跑跑？[14:54]