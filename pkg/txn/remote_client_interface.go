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

type RemoteClient interface {
	Read(key string, ts time.Time) (DataItem, error)
	Prepare(itemList []DataItem, startTime time.Time, commitTime time.Time) (map[string]string, error)
	Commit(infoList []CommitInfo) error
}
