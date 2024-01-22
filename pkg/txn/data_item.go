package txn

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type DataItem struct {
	Key       string       `redis:"Key" bson:"_id"`
	Value     string       `redis:"Value" bson:"Value"`
	TxnId     string       `redis:"TxnId" bson:"TxnId"`
	TxnState  config.State `redis:"TxnState" bson:"TxnState"`
	TValid    time.Time    `redis:"TValid" bson:"TValid"`
	TLease    time.Time    `redis:"TLease" bson:"TLease"`
	Prev      string       `redis:"Prev" bson:"Prev"`
	LinkedLen int          `redis:"LinkedLen" bson:"LinkedLen"`
	IsDeleted bool         `redis:"IsDeleted" bson:"IsDeleted"`
	Version   int          `redis:"Version" bson:"Version"`
}

func (m DataItem) GetKey() string {
	return m.Key
}

func (r DataItem) String() string {
	return fmt.Sprintf(`DataItem{
    Key:       %s,
    Value:     %s,
    TxnId:     %s,
    TxnState:  %s,
    TValid:    %s,
    TLease:    %s,
    Prev:      %s,
	LinkedLen: %d,
    IsDeleted: %v,
    Version:   %d,
}`, r.Key, r.Value, r.TxnId, util.ToString(r.TxnState),
		r.TValid.Format(time.RFC3339), r.TLease.Format(time.RFC3339),
		r.Prev, r.LinkedLen, r.IsDeleted, r.Version)
}

func (r *DataItem) Equal(other DataItem) bool {
	return r.Key == other.Key &&
		r.Value == other.Value &&
		r.TxnId == other.TxnId &&
		r.TxnState == other.TxnState &&
		r.TValid.Equal(other.TValid) &&
		r.TLease.Equal(other.TLease) &&
		r.Prev == other.Prev &&
		r.LinkedLen == other.LinkedLen &&
		r.IsDeleted == other.IsDeleted &&
		r.Version == other.Version
}

func (r DataItem) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

func (mi DataItem) MarshalBSONValue() (bsontype.Type, []byte, error) {
	m := bson.M{
		"Key":       mi.Key,
		"Value":     mi.Value,
		"TxnId":     mi.TxnId,
		"TxnState":  mi.TxnState,
		"TValid":    mi.TValid.Format(time.RFC3339Nano),
		"TLease":    mi.TLease.Format(time.RFC3339Nano),
		"Prev":      mi.Prev,
		"LinkedLen": mi.LinkedLen,
		"IsDeleted": mi.IsDeleted,
		"Version":   mi.Version,
	}
	return bson.MarshalValue(m)
}

func (mi *DataItem) UnmarshalBSONValue(t bsontype.Type, raw []byte) error {
	var m map[string]interface{}

	err := bson.Unmarshal(raw, &m)
	if err != nil {
		return err
	}

	if value, ok := m["_id"]; ok {
		mi.Key = value.(string)
	}
	if value, ok := m["Value"]; ok {
		mi.Value = value.(string)
	}
	if value, ok := m["TxnId"]; ok {
		mi.TxnId = value.(string)
	}
	if value, ok := m["TxnState"]; ok {
		mi.TxnState = config.State(value.(int32))
	}
	if value, ok := m["TValid"]; ok {
		mi.TValid, err = time.Parse(time.RFC3339Nano, value.(string))
		if err != nil {
			return err
		}
	}
	if value, ok := m["TLease"]; ok {
		mi.TLease, err = time.Parse(time.RFC3339Nano, value.(string))
		if err != nil {
			return err
		}
	}
	if value, ok := m["Prev"]; ok {
		mi.Prev = value.(string)
	}
	if value, ok := m["LinkedLen"]; ok {
		mi.LinkedLen = int(value.(int32))
	}
	if value, ok := m["IsDeleted"]; ok {
		mi.IsDeleted = value.(bool)
	}
	if value, ok := m["Version"]; ok {
		version := value.(int32)
		mi.Version = int(version)
	}

	return nil
}
