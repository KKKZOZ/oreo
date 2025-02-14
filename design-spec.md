# Design Spec

## Datastore Template

为了适配新的数据抽象, 同时兼容旧的数据抽象, 我们需要把 `transaction.go` 和 `datastore-template.go` 目前依赖的嵌套逻辑提取为新的数据抽象, 主要有以下接口:

```go
type DataItemNext interface {
    TxnMetadata
    TxnValueOperator
}

type TxnMetadata interface {
    Key() string

    GroupKeyList() string
    SetGroupKeyList(string)

    TxnState() config.State
    SetTxnState(config.State)

    Version() string
    SetVersion(string)

    TValid() int64
    SetTValid(int64)

    TLease() time.Time
    SetTLease(time.Time)

    IsDeleted() bool
    SetIsDeleted(bool)
    Empty() bool
}

type TxnValueOperator interface {
    ParseValue(any) error
    // need to serialize the value to string first
    SetValue(any) error

    UpdateMetadata(DataItemNext, int64, time.Time) error
    GetValidItem(TStart int64) (DataItemNext, bool)
    // if prev is empty, return "key not found" Error
    // if deserialize failed, return "deserialize failed" Error
    GetPrevItem() (DataItemNext, error)
}
```

## Connector

Connector 与 DataItemNext 之间的交互应该注意单一职责原则

DataItemNext 的自定义写入逻辑应该在 Item 自身中维护

Connector 只负责调用对应函数, 并传递相关参数 (比如 `rdb`)

Connector 中相关的接口有:

+ `GetItem(key string) (DataItem, error)`
+ `PutItem(key string, value DataItem) (string, error)`
+ `ConditionalUpdate(key string, value DataItem, doCreate bool) (string, error)`
+ `ConditionalCommit(key string, version string, tCommit int64) (string, error)`
+ `AtomicCreate(name string, value any) (string, error)`

需要修改为下面的接口:

+ `GetItem`
