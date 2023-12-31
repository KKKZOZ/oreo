package mongo

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type MongoItem struct {
	Key       string       `bson:"Key"`
	Value     string       `bson:"Value"`
	TxnId     string       `bson:"TxnId"`
	TxnState  config.State `bson:"TxnState"`
	TValid    time.Time    `bson:"TValid"`
	TLease    time.Time    `bson:"TLease"`
	Prev      string       `bson:"Prev"`
	LinkedLen int          `bson:"LinkedLen"`
	IsDeleted bool         `bson:"IsDeleted"`
	Version   int          `bson:"Version"`
}

func (m MongoItem) GetKey() string {
	return m.Key
}

func (r MongoItem) String() string {
	return fmt.Sprintf(`MongoItem{
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

func (r *MongoItem) Equal(other MongoItem) bool {
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

func (mi MongoItem) MarshalBinary() (data []byte, err error) {
	return json.Marshal(mi)
}

func (mi MongoItem) MarshalBSONValue() (bsontype.Type, []byte, error) {
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

func (mi *MongoItem) UnmarshalBSONValue(t bsontype.Type, raw []byte) error {
	var m map[string]interface{}

	err := bson.Unmarshal(raw, &m)
	if err != nil {
		return err
	}

	if value, ok := m["Key"]; ok {
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
