package txn

import (
	"time"

	"github.com/kkkzoz/oreo/pkg/config"
)

type RemoteDataStrategy string

const (
	Normal       RemoteDataStrategy = "Normal"
	AssumeAbort  RemoteDataStrategy = "AssumeAbort"
	AssumeCommit RemoteDataStrategy = "AssumeCommit"
)

type ItemType string

const (
	NoneItem  ItemType = ""
	RedisItem ItemType = "redis"
	MongoItem ItemType = "mongo"
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
	GlobalName   string
	MaxRecordLen int
	ReadStrategy config.ReadStrategy
}

type RemoteClient interface {
	Read(dsName string, key string, ts time.Time, config RecordConfig) (DataItem, RemoteDataStrategy, error)
	Prepare(dsName string, itemList []DataItem,
		startTime time.Time, commitTime time.Time,
		config RecordConfig, validationMap map[string]PredicateInfo) (map[string]string, error)
	Commit(dsName string, infoList []CommitInfo) error
	Abort(dsName string, keyList []string, txnId string) error
}
