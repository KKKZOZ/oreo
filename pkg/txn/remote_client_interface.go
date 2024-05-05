package txn

import "time"

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
}

type RemoteClient interface {
	Read(key string, ts time.Time, config RecordConfig) (DataItem, error)
	Prepare(itemList []DataItem, startTime time.Time, commitTime time.Time, config RecordConfig) (map[string]string, error)
	Commit(infoList []CommitInfo) error
	Abort(keyList []string, txnId string) error
}
