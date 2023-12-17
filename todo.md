
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


### TxnManager

整个事务的提交操作应该由 Transaction 来决定，所以 Transaction 至少要知道以下信息：

+ 本次事务中涉及到的 datastore
+ 本次事务中的 globalDataStore 
  
在事务开始时，暂且认定 globalDataStore 也会参与该事务

并且每个 datastore 也需要知道当前的 Transaction:

+ 访问 txnId
+ 访问 TSR


### Abstraction

DataStore 中应该有对应数据库的驱动/连接，建议再单独包装一层

+ 数据库连接应该有一个数据库连接池进行维护