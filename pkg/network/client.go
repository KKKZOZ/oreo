package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/txn"
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
	ServerAddr string
}

func NewClient(serverAddr string) *Client {
	serverAddr = "http://" + serverAddr
	return &Client{
		ServerAddr: serverAddr,
	}
}

func (c *Client) Read(key string, ts time.Time) (txn.DataItem, error) {
	// startTime := time.Now()
	// defer func() {
	// 	fmt.Printf("Read request latency: %v\n", time.Since(startTime))
	// }()

	data := ReadRequest{
		Key:       key,
		StartTime: ts,
	}
	json_data, _ := json.Marshal(data)

	reqUrl := c.ServerAddr + "/read"

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
		return response.Data, nil
	} else {
		errMsg := response.ErrMsg
		return nil, errors.New(errMsg)
	}
}

func (c *Client) Prepare(itemList []txn.DataItem,
	startTime time.Time, commitTime time.Time) (map[string]string, error) {
	// sTime := time.Now()
	// defer func() {
	// 	fmt.Printf("Prepare request latency: %v\n", time.Since(sTime))
	// }()

	itemArr := make([]redis.RedisItem, 0)
	for _, item := range itemList {
		redisItem, ok := item.(*redis.RedisItem)
		if !ok {
			return nil, errors.New("unexpected data type")
		}
		itemArr = append(itemArr, *redisItem)
	}
	data := PrepareRequest{
		ItemList:   itemArr,
		StartTime:  startTime,
		CommitTime: commitTime,
	}
	json_data, _ := json.Marshal(data)

	// fmt.Printf("Prepare request: %v\n", data)

	reqUrl := c.ServerAddr + "/prepare"
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
	var response Response[map[string]string]
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Prepare call resp Unmarshal error: %v\nbody: %v", err, string(body))
	}
	if response.Status == "OK" {
		return response.Data, nil
	} else {
		errMsg := response.ErrMsg
		return nil, errors.New(errMsg)
	}
}

func (c *Client) Commit(infoList []txn.CommitInfo) error {
	// startTime := time.Now()
	// defer func() {
	// 	fmt.Printf("Commit request latency: %v\n", time.Since(startTime))
	// }()

	data := CommitRequest{
		List: infoList,
	}
	json_data, _ := json.Marshal(data)

	reqUrl := c.ServerAddr + "/commit"

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

func (c *Client) Abort(keyList []string, txnId string) error {
	data := AbortRequest{
		KeyList: keyList,
		TxnId:   txnId,
	}
	json_data, _ := json.Marshal(data)

	reqUrl := c.ServerAddr + "/abort"

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
