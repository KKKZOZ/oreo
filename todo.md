
### version tag

如果 DataStore 对上层返回的只是应用层数据的话，那么版本号如何维护？

很多情况下都是 read-modify-write 的模式，我们暂且采用以下规则：

在 DataStore 中维护一个 map，versionMap:

+ key: record's key
+ value: record's version

在执行 write 操作时：

+ 如果 cache 中没有该数据，说明是第一次写入，从 versionMap 中获取：
  + 如果 versionMap 中没有对应的 key，说明这个数据是新创建的，把 version 设为 1
  + 如果 versionMap 中有对应的 key，则 `item.Version = versionMap[key] + 1`
+ 如果 cache 中已经有数据了，直接替换 value 即可