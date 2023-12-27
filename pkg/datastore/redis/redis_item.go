package redis

import (
	"fmt"
	"time"

	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
)

type RedisItem struct {
	Key       string       `redis:"Key"`
	Value     string       `redis:"Value"`
	TxnId     string       `redis:"TxnId"`
	TxnState  config.State `redis:"TxnState"`
	TValid    time.Time    `redis:"TValid"`
	TLease    time.Time    `redis:"TLease"`
	Prev      string       `redis:"Prev"`
	IsDeleted bool         `redis:"IsDeleted"`
	Version   int          `redis:"Version"`
}

func (m RedisItem) GetKey() string {
	return m.Key
}

func (r RedisItem) String() string {
	return fmt.Sprintf(`RedisItem{
    Key:       %s,
    Value:     %s,
    TxnId:     %s,
    TxnState:  %s,
    TValid:    %s,
    TLease:    %s,
    Prev:      %s,
    IsDeleted: %v,
    Version:   %d,
}`, r.Key, r.Value, r.TxnId, util.ToString(r.TxnState),
		r.TValid.Format(time.RFC3339), r.TLease.Format(time.RFC3339),
		r.Prev, r.IsDeleted, r.Version)
}

func (r *RedisItem) Equal(other RedisItem) bool {
	return r.Key == other.Key &&
		r.Value == other.Value &&
		r.TxnId == other.TxnId &&
		r.TxnState == other.TxnState &&
		r.TValid.Equal(other.TValid) &&
		r.TLease.Equal(other.TLease) &&
		r.Prev == other.Prev &&
		r.IsDeleted == other.IsDeleted &&
		r.Version == other.Version
}
