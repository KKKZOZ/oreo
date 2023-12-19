
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


在执行 delete 操作时：

+ 如果 cache 中没有该数据，说明是根据业务逻辑直接删除，需要执行一次 txn read 来确定版本号
+ 如果 cache 中有这项数据，直接修改 `isDeleted` 即可


### Transaction

整个事务的提交操作应该由 Transaction 来决定，所以 Transaction 至少要知道以下信息：

+ 本次事务中涉及到的 datastore
+ 本次事务中的 globalDataStore 

在事务开始时，暂且认定 globalDataStore 也会参与该事务

并且每个 datastore 也需要知道当前的 Transaction:

+ 访问 txnId
+ 访问 TSR

#### Transaction Commit

两个阶段：

Prepare：

+ 如果任何一个 datastore 失败了，对每个 datastore 都调用 `abort()`
  + `abort()` 失败时要不断重试
+ 对于那个引起失败的 datastore，其整个写集可以被分为两部分
  + updated：已经写入了 datastore 的 newItem，TxnId 等于当前事务的 id
  + to be updated: 还未写入，所以不用管
+ 所以：以 cache 中的 key 为列表，挨个去检查是否已经更改，更改了的话就rollback


### Abstraction

DataStore 中应该有对应数据库的驱动/连接，建议再单独包装一层

+ 数据库连接应该有一个数据库连接池进行维护


### 关于 delete

delete 操作在执行第二次时暂时默认为报错

delete 操作在论文中的描述是 "will only happen in the data store after commit"

相当与需要在数据库内部支持一个垃圾回收的机制，可以写但是没必要，直接在 record 上加一个 metadata:

+ `isDeleted`

在 Transaction Read 中，只需要在知道可见性后进行一步额外的处理就行


### 关于 record 的版本维护

一个 record 其实只需要两个版本:

+ Previous: 一定为 COMMITTED 的状态
+ Current

所以在 `conditionalUpdate` 时记得清理


### Test Case

#### Transaction

+ TestTwoTxnWriteThenRead


### TODO

#### 12.18

+ 完善 Transaction 的逻辑和代码，把单个事务的流程全部测试一遍
+ 寻找工具，研究一下测试覆盖率
+ TSR 写入的问题