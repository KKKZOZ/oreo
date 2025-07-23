package network

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/kkkzoz/oreo/pkg/datastore/cassandra"
	"github.com/kkkzoz/oreo/pkg/datastore/couchdb"
	"github.com/kkkzoz/oreo/pkg/datastore/dynamodb"
	"github.com/kkkzoz/oreo/pkg/datastore/mongo"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/datastore/tikv"
	"github.com/kkkzoz/oreo/pkg/txn"
)

var json2 = jsoniter.ConfigCompatibleWithStandardLibrary

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
	GroupKey     string
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
	type TempResponse struct {
		Status       string
		ErrMsg       string
		DataStrategy txn.RemoteDataStrategy
		ItemType     txn.ItemType        `json:"ItemType"`
		Data         jsoniter.RawMessage `json:"Data"`
	}

	var aux TempResponse
	if err := json2.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal basic fields: %v", err)
	}

	r.Status = aux.Status
	r.ErrMsg = aux.ErrMsg
	r.DataStrategy = aux.DataStrategy
	r.ItemType = aux.ItemType

	switch r.ItemType {
	case txn.RedisItem:
		var redisItem redis.RedisItem
		if err := json2.Unmarshal(aux.Data, &redisItem); err != nil {
			return err
		}
		r.Data = &redisItem
	case txn.MongoItem:
		var mongoItem mongo.MongoItem
		if err := json2.Unmarshal(aux.Data, &mongoItem); err != nil {
			return err
		}
		r.Data = &mongoItem
	case txn.CouchItem:
		var couchItem couchdb.CouchDBItem
		if err := json2.Unmarshal(aux.Data, &couchItem); err != nil {
			return err
		}
		r.Data = &couchItem
	case txn.CassandraItem:
		var cassandraItem cassandra.CassandraItem
		if err := json2.Unmarshal(aux.Data, &cassandraItem); err != nil {
			return err
		}
		r.Data = &cassandraItem
	case txn.DynamoDBItem:
		var dynamoDBItem dynamodb.DynamoDBItem
		if err := json2.Unmarshal(aux.Data, &dynamoDBItem); err != nil {
			return err
		}
		r.Data = &dynamoDBItem
	case txn.TiKVItem:
		var tikvItem tikv.TiKVItem
		if err := json2.Unmarshal(aux.Data, &tikvItem); err != nil {
			return err
		}
		r.Data = &tikvItem
	case txn.NoneItem:
		r.Data = nil
	default:
		return fmt.Errorf("[network.go - ReadResponse] unsupported data type: %v", r.ItemType)
	}

	return nil
}

func (p *PrepareRequest) UnmarshalJSON(data []byte) error {
	type TempRequest struct {
		DsName        string                       `json:"DsName"`
		ValidationMap map[string]txn.PredicateInfo `json:"ValidationMap"`
		ItemType      txn.ItemType                 `json:"ItemType"`
		StartTime     int64                        `json:"StartTime"`
		Config        txn.RecordConfig             `json:"Config"`
		ItemList      jsoniter.RawMessage          `json:"ItemList"`
	}

	var aux TempRequest
	if err := json2.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal basic fields: %v", err)
	}

	p.DsName = aux.DsName
	p.ValidationMap = aux.ValidationMap
	p.ItemType = aux.ItemType
	p.StartTime = aux.StartTime
	p.Config = aux.Config

	// fmt.Printf("Item Type: %v\n", p.ItemType)
	// fmt.Printf("Item List: %v\n", string(aux.ItemList))

	switch p.ItemType {
	case txn.RedisItem:
		var redisItemList []redis.RedisItem
		if err := json2.Unmarshal(aux.ItemList, &redisItemList); err != nil {
			return err
		}
		p.ItemList = make([]txn.DataItem, len(redisItemList))
		for i, it := range redisItemList {
			item := it
			p.ItemList[i] = &item
		}
	case txn.MongoItem:
		var mongoItemList []mongo.MongoItem
		if err := json2.Unmarshal(aux.ItemList, &mongoItemList); err != nil {
			return err
		}
		p.ItemList = make([]txn.DataItem, len(mongoItemList))
		for i, it := range mongoItemList {
			item := it
			p.ItemList[i] = &item
		}
	case txn.CouchItem:
		var couchItemList []couchdb.CouchDBItem
		if err := json2.Unmarshal(aux.ItemList, &couchItemList); err != nil {
			return err
		}
		p.ItemList = make([]txn.DataItem, len(couchItemList))
		for i, it := range couchItemList {
			item := it
			p.ItemList[i] = &item
		}
	case txn.CassandraItem:
		var cassandraItemList []cassandra.CassandraItem
		if err := json2.Unmarshal(aux.ItemList, &cassandraItemList); err != nil {
			return err
		}
		p.ItemList = make([]txn.DataItem, len(cassandraItemList))
		for i, it := range cassandraItemList {
			item := it
			p.ItemList[i] = &item
		}
	case txn.DynamoDBItem:
		var dynamoDBItemList []dynamodb.DynamoDBItem
		if err := json2.Unmarshal(aux.ItemList, &dynamoDBItemList); err != nil {
			return err
		}
		p.ItemList = make([]txn.DataItem, len(dynamoDBItemList))
		for i, it := range dynamoDBItemList {
			item := it
			p.ItemList[i] = &item
		}
	case txn.TiKVItem:
		var tikvItemList []tikv.TiKVItem
		if err := json2.Unmarshal(aux.ItemList, &tikvItemList); err != nil {
			return err
		}
		p.ItemList = make([]txn.DataItem, len(tikvItemList))
		for i, it := range tikvItemList {
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

func GetItemType(dsName string) txn.ItemType {
	switch dsName {
	case "redis1", "Redis":
		return txn.RedisItem
	case "mongo1", "mongo2", "MongoDB", "MongoDB1", "MongoDB2":
		return txn.MongoItem
	case "CouchDB":
		return txn.CouchItem
	case "KVRocks":
		return txn.RedisItem
	case "Cassandra":
		return txn.CassandraItem
	case "DynamoDB":
		return txn.DynamoDBItem
	case "TiKV":
		return txn.TiKVItem
	default:
		return ""
	}
}
