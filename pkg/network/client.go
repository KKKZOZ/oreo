package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/logger"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

var _ txn.RemoteClient = (*Client)(nil)

var HttpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        6000,
		MaxIdleConnsPerHost: 1000,
		MaxConnsPerHost:     1000,
	},
}

type Client struct {
	ServerAddrList []string
	mutex          sync.Mutex
	curIndex       int
}

func NewClient(serverAddrList []string) *Client {

	addrList := make([]string, 0)

	for _, serverAddr := range serverAddrList {
		serverAddr = "http://" + serverAddr
		addrList = append(addrList, serverAddr)
	}
	return &Client{
		ServerAddrList: addrList,
	}
}

func (c *Client) GetServerAddr() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.curIndex >= len(c.ServerAddrList) {
		c.curIndex = 0
	}
	addr := c.ServerAddrList[c.curIndex]
	c.curIndex++
	return addr
}

func (c *Client) Read(dsName string, key string, ts int64, cfg txn.RecordConfig) (txn.DataItem, txn.RemoteDataStrategy, error) {

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}

	data := ReadRequest{
		DsName:    dsName,
		Key:       key,
		StartTime: ts,
		Config:    cfg,
	}
	json_data, _ := json.Marshal(data)

	reqUrl := c.GetServerAddr() + "/read"

	// Create a new POST request
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(json_data))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := HttpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var response ReadResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatal(err)
	}
	if response.Status == "OK" {
		return response.Data, response.DataStrategy, nil
	} else {
		errMsg := response.ErrMsg
		return nil, txn.Normal, errors.New(errMsg)
	}
}

func (c *Client) Prepare(dsName string, itemList []txn.DataItem,
	startTime int64, cfg txn.RecordConfig,
	validationMap map[string]txn.PredicateInfo) (map[string]string, int64, error) {

	debugStart := time.Now()

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}

	// itemArr := make([]redis.RedisItem, 0)
	// for _, item := range itemList {
	// 	redisItem, ok := item.(*redis.RedisItem)
	// 	if !ok {
	// 		return nil, errors.New("unexpected data type")
	// 	}
	// 	itemArr = append(itemArr, *redisItem)
	// }
	data := PrepareRequest{
		DsName:        dsName,
		ItemType:      c.getItemType(dsName),
		ItemList:      itemList,
		StartTime:     startTime,
		Config:        cfg,
		ValidationMap: validationMap,
	}
	json_data, _ := json.Marshal(data)

	reqUrl := c.GetServerAddr() + "/prepare"
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(json_data))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	debugMsg := fmt.Sprintf("HttpClient.Do(Prepare) in %v", dsName)
	resp, err := HttpClient.Do(req)
	logger.Log.Debugw("After "+debugMsg, "LatencyInFunc", time.Since(debugStart), "Topic", "CheckPoint")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var response PrepareResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Prepare call resp Unmarshal error: %v\nbody: %v", err, string(body))
	}
	if response.Status == "OK" {
		return response.VerMap, response.TCommit, nil
	} else {
		errMsg := response.ErrMsg
		return nil, 0, errors.New(errMsg)
	}
}

func (c *Client) Commit(dsName string, infoList []txn.CommitInfo, tCommit int64) error {

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}

	data := CommitRequest{
		DsName:  dsName,
		List:    infoList,
		TCommit: tCommit,
	}
	json_data, _ := json.Marshal(data)

	reqUrl := c.GetServerAddr() + "/commit"

	// Create a new POST request
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(json_data))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := HttpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var response Response[string]
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Commit call resp Unmarshal error: %v\nbody: %v", err, string(body))
	}
	if response.Status == "OK" {
		return nil
	} else {
		errMsg := response.ErrMsg
		return errors.New(errMsg)
	}
}

func (c *Client) Abort(dsName string, keyList []string, txnId string) error {

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}

	data := AbortRequest{
		DsName:  dsName,
		KeyList: keyList,
		TxnId:   txnId,
	}
	json_data, _ := json.Marshal(data)

	reqUrl := c.GetServerAddr() + "/abort"

	// Create a new POST request
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(json_data))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := HttpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var response Response[string]
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Abort call resp Unmarshal error: %v\nbody: %v", err, string(body))
	}
	if response.Status == "OK" {
		return nil
	} else {
		errMsg := response.ErrMsg
		return errors.New(errMsg)
	}
}

func (c *Client) getItemType(dsName string) txn.ItemType {
	switch dsName {
	case "redis1", "Redis":
		return txn.RedisItem
	case "mongo1", "mongo2", "MongoDB":
		return txn.MongoItem
	case "CouchDB":
		return txn.CouchItem
	case "KVRocks":
		return txn.RedisItem
	default:
		return ""
	}
}
