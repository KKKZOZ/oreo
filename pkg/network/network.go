package network

import (
	"encoding/json"
	"fmt"

	"github.com/oreo-dtx-lab/oreo/pkg/datastore/couchdb"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/mongo"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/redis"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Response[T any] struct {
	Status string
	ErrMsg string
	Data   T
}

type ReadResponse struct {
	Status       string
	ErrMsg       string
	DataStrategy txn.RemoteDataStrategy
	ItemType     txn.ItemType
	Data         txn.DataItem
}

type ReadRequest struct {
	DsName    string
	Key       string
	StartTime int64
	Config    txn.RecordConfig
}

type PrepareRequest struct {
	DsName        string
	ValidationMap map[string]txn.PredicateInfo
	ItemType      txn.ItemType
	ItemList      []txn.DataItem
	StartTime     int64
	Config        txn.RecordConfig
}

type PrepareResponse struct {
	Status  string
	ErrMsg  string
	TCommit int64
	VerMap  map[string]string
}

type CommitRequest struct {
	DsName  string
	List    []txn.CommitInfo
	TCommit int64
}

type AbortRequest struct {
	DsName  string
	KeyList []string
	// TxnId   string
	GroupKeyList string
}

func (r *ReadResponse) UnmarshalJSON(data []byte) error {
	type Alias ReadResponse
	aux := &struct {
		Data json.RawMessage `json:"Data"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch r.ItemType {
	case txn.RedisItem:
		var redisItem redis.RedisItem
		if err := json.Unmarshal(aux.Data, &redisItem); err != nil {
			return err
		}
		r.Data = &redisItem
	case txn.MongoItem:
		var mongoItem mongo.MongoItem
		if err := json.Unmarshal(aux.Data, &mongoItem); err != nil {
			return err
		}
		r.Data = &mongoItem
	case txn.CouchItem:
		var couchItem couchdb.CouchDBItem
		if err := json.Unmarshal(aux.Data, &couchItem); err != nil {
			return err
		}
		r.Data = &couchItem
	case txn.NoneItem:
		r.Data = nil
	default:
		return fmt.Errorf("[network.go - ReadResponse] unsupported data type: %v", r.ItemType)
	}

	return nil
}

func (p *PrepareRequest) UnmarshalJSON(data []byte) error {
	type Alias PrepareRequest
	aux := &struct {
		ItemList json.RawMessage `json:"ItemList"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch p.ItemType {
	case txn.RedisItem:
		var redisItemList []redis.RedisItem
		if err := json.Unmarshal(aux.ItemList, &redisItemList); err != nil {
			return err
		}
		p.ItemList = make([]txn.DataItem, len(redisItemList))
		for i, it := range redisItemList {
			item := it
			p.ItemList[i] = &item
		}
	case txn.MongoItem:
		var mongoItemList []mongo.MongoItem
		if err := json.Unmarshal(aux.ItemList, &mongoItemList); err != nil {
			return err
		}
		p.ItemList = make([]txn.DataItem, len(mongoItemList))
		for i, it := range mongoItemList {
			item := it
			p.ItemList[i] = &item
		}
	case txn.NoneItem:
		p.ItemList = nil
	default:
		return fmt.Errorf("[network.go - PrepareRequest]unsupported data type: %v", p.ItemType)
	}

	return nil
}
