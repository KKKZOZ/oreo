package txn

import (
	"time"

	"github.com/kkkzoz/oreo/pkg/config"
)

type RemoteDataType string

const (
	Normal       RemoteDataType = "Normal"
	AssumeAbort  RemoteDataType = "AssumeAbort"
	AssumeCommit RemoteDataType = "AssumeCommit"
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
	MaxRecordLen int
	ReadStrategy config.ReadStrategy
}

type RemoteClient interface {
	Read(key string, ts time.Time, config RecordConfig) (DataItem, RemoteDataType, error)
	Prepare(itemList []DataItem, startTime time.Time, commitTime time.Time, config RecordConfig) (map[string]string, error)
	Commit(infoList []CommitInfo) error
	Abort(keyList []string, txnId string) error
}
