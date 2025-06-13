package txn

import (
	"encoding/json"
	"fmt"

	"github.com/oreo-dtx-lab/oreo/pkg/config"
)

type GroupKey struct {
	Key string
	GroupKeyItem
}

type GroupKeyItem struct {
	TxnState config.State
	TCommit  int64
}

func (item *GroupKeyItem) MarshalBinary() ([]byte, error) {
	return json.Marshal(item)
}

func (item *GroupKeyItem) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, item)
}

func NewGroupKeyItem(txnState config.State, tCommit int64) GroupKeyItem {
	return GroupKeyItem{
		TxnState: txnState,
		TCommit:  tCommit,
	}
}

func NewGroupKey(key string, txnState config.State, tCommit int64) *GroupKey {
	return &GroupKey{
		Key: key,
		GroupKeyItem: GroupKeyItem{
			TxnState: txnState,
			TCommit:  tCommit,
		},
	}
}

func (gk *GroupKey) String() string {
	return fmt.Sprintf(`GroupKey{
		Key: %s,
		TxnState: %d,
		TCommit: %d,
		}`, gk.Key, gk.TxnState, gk.TCommit)
}

func CommittedForAll(groupKeys []GroupKey) bool {
	for _, gk := range groupKeys {
		if gk.TxnState != config.COMMITTED {
			return false
		}
	}
	return true
}

func AtLeastOneAborted(groupKeys []GroupKey) bool {
	for _, gk := range groupKeys {
		if gk.TxnState == config.ABORTED {
			return true
		}
	}
	return false
}

// func MakeGroupKeyListFromUrls(urls []string) string {
// 	list := ""
// 	for _, url := range urls {
// 		list += url + ","
// 	}
// 	return list
// }
