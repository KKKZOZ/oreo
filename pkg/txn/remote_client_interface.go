package txn

import (
	"github.com/oreo-dtx-lab/oreo/pkg/config"
)

type RemoteDataStrategy string

const (
	Normal       RemoteDataStrategy = "Normal"
	AssumeAbort  RemoteDataStrategy = "AssumeAbort"
	AssumeCommit RemoteDataStrategy = "AssumeCommit"
)

type ItemType string

const (
	NoneItem      ItemType = ""
	RedisItem     ItemType = "redis"
	MongoItem     ItemType = "mongo"
	CouchItem     ItemType = "couch"
	CassandraItem ItemType = "cassandra"
	DynamoDBItem  ItemType = "dynamodb"
	TiKVItem      ItemType = "tikv"
)

type NetworkItem struct {
	Item     DataItem
	DoCreate bool
}

type CommitInfo struct {
	Key     string
	Version string
}

type RecordConfig struct {
	// GlobalName                  string
	MaxRecordLen                int
	ReadStrategy                config.ReadStrategy
	ConcurrentOptimizationLevel int
	AblationLevel               int
}

type RemoteClient interface {
	Read(dsName string, key string, ts int64, config RecordConfig) (DataItem, RemoteDataStrategy, error)
	Prepare(dsName string, itemList []DataItem,
		startTime int64,
		config RecordConfig, validationMap map[string]PredicateInfo) (map[string]string, int64, error)
	Commit(dsName string, infoList []CommitInfo, TCommit int64) error
	Abort(dsName string, keyList []string, txnId string) error
}
