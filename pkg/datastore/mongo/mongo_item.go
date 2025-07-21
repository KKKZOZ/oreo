package mongo

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

var _ txn.DataItem = (*MongoItem)(nil)

type MongoItem struct {
	MKey          string       `bson:"_id" json:"Key"`
	MValue        string       `bson:"Value" json:"Value"`
	MGroupKeyList string       `bson:"GroupKeyList" json:"GroupKeyList"`
	MTxnState     config.State `bson:"TxnState" json:"TxnState"`
	MTValid       int64        `bson:"TValid" json:"TValid"`
	MTLease       time.Time    `bson:"TLease" json:"TLease"`
	MPrev         string       `bson:"Prev" json:"Prev"`
	MLinkedLen    int          `bson:"LinkedLen" json:"LinkedLen"`
	MIsDeleted    bool         `bson:"IsDeleted" json:"IsDeleted"`
	MVersion      string       `bson:"Version" json:"Version"`
}

func NewMongoItem(options txn.ItemOptions) *MongoItem {

	if options.Value == nil {
		options.Value = ""
	}

	return &MongoItem{
		MKey:          options.Key,
		MValue:        options.Value.(string),
		MGroupKeyList: options.GroupKeyList,
		MTxnState:     options.TxnState,
		MTValid:       options.TValid,
		MTLease:       options.TLease,
		MPrev:         options.Prev,
		MLinkedLen:    options.LinkedLen,
		MIsDeleted:    options.IsDeleted,
		MVersion:      options.Version,
	}
}

func (m *MongoItem) Key() string {
	return m.MKey
}

func (m *MongoItem) Value() string {
	return m.MValue
}

func (m *MongoItem) SetValue(value string) {
	m.MValue = value
}

func (m *MongoItem) GroupKeyList() string {
	return m.MGroupKeyList
}

func (m *MongoItem) SetGroupKeyList(groupKeyList string) {
	m.MGroupKeyList = groupKeyList
}

func (m *MongoItem) TxnState() config.State {
	return m.MTxnState
}

func (m *MongoItem) SetTxnState(state config.State) {
	m.MTxnState = state
}

func (m *MongoItem) TValid() int64 {
	return m.MTValid
}

func (m *MongoItem) SetTValid(tValid int64) {
	m.MTValid = tValid
}

func (m *MongoItem) TLease() time.Time {
	return m.MTLease
}

func (m *MongoItem) SetTLease(tLease time.Time) {
	m.MTLease = tLease
}

func (m *MongoItem) Prev() string {
	return m.MPrev
}

func (m *MongoItem) SetPrev(prev string) {
	m.MPrev = prev
}

func (m *MongoItem) LinkedLen() int {
	return m.MLinkedLen
}

func (m *MongoItem) SetLinkedLen(linkedLen int) {
	m.MLinkedLen = linkedLen
}

func (m *MongoItem) IsDeleted() bool {
	return m.MIsDeleted
}

func (m *MongoItem) SetIsDeleted(isDeleted bool) {
	m.MIsDeleted = isDeleted
}

func (m *MongoItem) Version() string {
	return m.MVersion
}

func (m *MongoItem) SetVersion(version string) {
	m.MVersion = version
}

func (r MongoItem) String() string {
	return fmt.Sprintf(`MongoItem{
    Key:       %s,
    Value:     %s,
    GroupKeyList:     %s,
    TxnState:  %s,
    TValid:    %d,
    TLease:    %s,
    Prev:      %s,
	LinkedLen: %d,
    IsDeleted: %v,
    Version:   %s,
}`, r.MKey, r.MValue, r.MGroupKeyList, util.ToString(r.MTxnState),
		r.MTValid, r.MTLease.Format(time.RFC3339),
		r.MPrev, r.MLinkedLen, r.MIsDeleted, r.MVersion)
}

func (r *MongoItem) Empty() bool {
	return r.MKey == "" && r.MValue == "" &&
		r.MGroupKeyList == "" && r.MTxnState == config.State(0) &&
		r.MTValid == 0 && r.MTLease.IsZero() &&
		r.MPrev == "" && r.MLinkedLen == 0 &&
		!r.MIsDeleted && r.MVersion == ""
}

func (r *MongoItem) Equal(other txn.DataItem) bool {
	return r.MKey == other.Key() &&
		r.MValue == other.Value() &&
		r.MGroupKeyList == other.GroupKeyList() &&
		r.MTxnState == other.TxnState() &&
		r.MTValid == other.TValid() &&
		r.MTLease.Equal(other.TLease()) &&
		r.MPrev == other.Prev() &&
		r.MLinkedLen == other.LinkedLen() &&
		r.MIsDeleted == other.IsDeleted() &&
		r.MVersion == other.Version()
}

func (mi MongoItem) MarshalBinary() (data []byte, err error) {
	return json.Marshal(mi)
}

func (mi MongoItem) MarshalBSONValue() (bsontype.Type, []byte, error) {
	m := bson.M{
		"Key":          mi.MKey,
		"Value":        mi.MValue,
		"GroupKeyList": mi.MGroupKeyList,
		"TxnState":     mi.MTxnState,
		"TValid":       mi.MTValid,
		"TLease":       mi.MTLease.Format(time.RFC3339Nano),
		"Prev":         mi.MPrev,
		"LinkedLen":    mi.MLinkedLen,
		"IsDeleted":    mi.MIsDeleted,
		"Version":      mi.MVersion,
	}
	return bson.MarshalValue(m)
}

func (mi *MongoItem) UnmarshalBSONValue(t bsontype.Type, raw []byte) error {
	var m map[string]interface{}

	err := bson.Unmarshal(raw, &m)
	if err != nil {
		return err
	}

	if value, ok := m["_id"]; ok {
		mi.MKey = value.(string)
	}
	if value, ok := m["Value"]; ok {
		mi.MValue = value.(string)
	}
	if value, ok := m["GroupKeyList"]; ok {
		mi.MGroupKeyList = value.(string)
	}
	if value, ok := m["TxnState"]; ok {
		mi.MTxnState = config.State(value.(int32))
	}
	if value, ok := m["TValid"]; ok {
		// mi.MTValid, err = time.Parse(time.RFC3339Nano, value.(string))
		// if err != nil {
		// 	return err
		// }
		mi.MTValid = value.(int64)
	}
	if value, ok := m["TLease"]; ok {
		mi.MTLease, err = time.Parse(time.RFC3339Nano, value.(string))
		if err != nil {
			return err
		}
	}
	if value, ok := m["Prev"]; ok {
		mi.MPrev = value.(string)
	}
	if value, ok := m["LinkedLen"]; ok {
		mi.MLinkedLen = int(value.(int32))
	}
	if value, ok := m["IsDeleted"]; ok {
		mi.MIsDeleted = value.(bool)
	}
	if value, ok := m["Version"]; ok {
		mi.MVersion = value.(string)
	}

	return nil
}
