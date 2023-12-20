
### version tag

很多情况下都是 read-modify-write 的模式，对于版本号的维护，我们暂且采用以下规则：

在执行 write 操作时：

+ 如果 writeCache 中没有该数据，说明是第一次写入，从 readCache 中获取：
  + 如果 readCache 中没有对应的 key，说明这个数据是新创建的，把 version 设为 1
  + 如果 readCache 中有对应的 key，则 `item.Version = readCache[key].Version`
+ 如果 writeCache 中已经有数据了，直接替换 value 即可


在执行 delete 操作时：

+ 如果 writeCache 中有这项数据，说明版本号有对应，直接修改 `isDeleted` 即可
+ 如果 readCache 中有这项数据，说明版本号也有对应并且该数据还未修改过，直接修改 `isDeleted` 即可
+ 如果 writeCache 和 readCache 中都没有这个数据，说明是根据业务逻辑直接删除，需要执行一次 txn read 来确定版本号，否则会在 conditionalUpdate 时出错(没有正确的版本): TestDeleteWithoutRead



### Transaction

整个事务的提交操作应该由 Transaction 来决定，所以 Transaction 至少要知道以下信息：

+ 本次事务中涉及到的 datastore
+ 本次事务中的 globalDataStore 

在事务开始时，暂且认定 globalDataStore 也会参与该事务

并且每个 datastore 也需要知道当前的 Transaction:

+ 访问 txnId
+ 访问 TSR

#### Transaction Read

注意读过的数据也必须先保存在 cache 中：

+ 把读集和写集分离

Transaction Read:

+ if the record is in writeCache, use the cached record
+ if the record is in readCache, use the cached record
+ read from connectoin:
  + if the record is in COMMITTED:
    + if $T_valid < T_start$: OK
    + else go to previous one(by unmarshal the `Prev` field)
    + **abort** if it can not find a corresponding one
  + if the record is in PREPARED:
    + if TSR (Transaction Status Record) exists (indicates that the transaction did commit): the record is considerer COMMITTED(*roll forward*)
    + if there is no TSR:
      + if the record's $T_{lease\_time}$ has **not** expired (indicates that the transaction is running, for example, is executing another datastore's `conditionalUpdate`): read fails
      + if the record's $T_{lease\_time}$ has expired: rollback the record


#### Transaction Commit

两个阶段：

Prepare：

+ 如果任何一个 datastore 失败了，对每个 datastore 都调用 `abort()`
  + `abort()` 失败时要不断重试
+ 对于那个引起失败的 datastore，其整个写集可以被分为两部分
  + updated：已经写入了 datastore 的 newItem，TxnId 等于当前事务的 id
  + to be updated: 还未写入，所以不用管
+ 所以：以 cache 中的 key 为列表，挨个去检查是否已经更改，更改了的话就 rollback

假设一个慢事务 Prepare 阶段执行得非常慢（比如说有一个 datastore 由于网络情况，其 `conditionalUpdate` 执行得非常慢），导致前面的 record 的 $T_{lease\_time}$ 过期了，这时如果有一个快事务访问这个 record， 就会执行 `rollback` 流程，即把 `Prev` 中的记录提出来。

+ 如果快事务提前修改了慢事务后续要修改的值，慢事务需要 abort:
  + 总共 1，2，3，4，5个记录
  + slowTxn 读 12345
  + fastTxn 读 345
  + slowTxn 从 1 开始修改，修改到 4 时卡住
  + fastTxn 从 3 开始修改，发现记录的 lease time 已经过期，所以 rollback 后继续修改
  + fastTxn 能够正常提交
  + slowTxn 在修改 4 时出现异常，abort
  + postTxn 需要检查 1，2 处于原状态，3，4，5 处于 fastTxn 修改后的状态

如果慢事务此时进入 commit 阶段，打算提交，就会进入数据不一致的情况: TestSlowTransactionRecordExpiredConsistency


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

### 关于 TSR

不能直接调用 Transaction Write 的接口，因为这个操作是暂时写到 cache 中，并没有写到 data store中,所以要新增两个接口：

+ WriteTSR(key string)
+ DeleteTSR(key string)


### Test Case

#### Transaction

+ TestTxnWrite
+ TestTxnRead

注意这里的测试不应该与 datastore 的重复，所以单个 Transaction 的读写不用再测试了，应该聚焦于

+ 两个并行的 Transaction 之间的情况
  + 写冲突(TestSingleKeyWriteConflict)：读同一个数据，然后同时写，应该有一个成功有一个失败
  + 全局有序的写冲突(TestMultiKeyWriteConflict)：读两个数据，然后同时写，按照 AB 和 BA 的顺序执行写，只有一个事物能成功

可以按照 Postgres 的可见性检查来写测试样例：

事务自行修改的可见性：

+ TestReadOwnWrite

并发事务修改的可见性：

+ 在本事务读某个数据时，对应的并发事务已经修改了该数据但还未提交
  + TestRepeatableReadWhenAnotherUncommitted
+ 在本事务读某个数据时，对应的并发事务已经修改了该数据并且已经提交
  + TestRepeatableReadWhenAnotherCommitted


Abort 情况分为两种：

单数据源：

+ 由于业务逻辑需要 abort (TestTxnAbort)
  + 这种情况下下只会在 commit() 调用之前进行调用，对应的操作是 clear the cache
  + 需要测试本数据源上所有曾经写入的 record 是不是都没生效
+ conditionalUpdate 时发生冲突 (TestTxnAbortCausedByWriteConflict)
  + 事务还在 Prepare 阶段，需要**自行**把数据源上已经写入的数据 rollback
+ conditionalUpdate 时线程终止
  + 由下一个读到相关 record 的事务来进行 rollback
+ commit 阶段时线程终止
  + 由下一个读到相关 record 的事务来进行 roll forward

多数据源同理

### ConditionalUpdate 时的状态分析

> datastore 中的为 oldItem

如果 oldItem 的状态为 Committed:

+ 版本号一致
  + 可以直接修改
+ 版本号不一致
  + 检测到写写冲突（对方已提交），事务终止


如果 oldItem 的状态为 Prepared:

+ 如果有对应的 TSR
  + 当作为 COMMITTED，检测到写写冲突，事务终止
+ 如果没有对应的 TSR
  + $T_{least\_time}$ has **not** expired: 无法确定对方事务状态(可能还在进行中)，事务终止
  + $T_{least\_time}$ has expired: 无法确定对方事务状态，事务终止



### TODO

Transaction: State 可以用 StateMachine 来管理状态


读写集分离

+ 在读一个数据时，应该先从写集里面查找，再从读集里面查找

+ 在写一个数据时，先查找 writeCache 中是否已经有记录。