package txn

import (
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

func (gk *GroupKey) String() string {
	return fmt.Sprintf(`GroupKey{
		Key: %s,
		TxnState: %d,
		TCommit: %d,
		}`, gk.Key, gk.TxnState, gk.TCommit)
}
