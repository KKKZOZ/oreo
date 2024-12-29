package network

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/logger"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	"github.com/valyala/fasthttp"
)

var _ txn.RemoteClient = (*Client)(nil)

type Client struct {
	ExecutorAddrMap map[string][]string
	mutex           sync.Mutex
	curIndexMap     map[string]int
}

const ALL = "ALL"

func NewClient(executorAddrMap map[string][]string) *Client {
	// addrList := make([]string, 0)

	// for _, serverAddr := range serverAddrList {
	// 	serverAddr = "http://" + serverAddr
	// 	addrList = append(addrList, serverAddr)
	// }
	curIndexMap := make(map[string]int)
	for dsName := range executorAddrMap {
		curIndexMap[dsName] = 0
	}
	return &Client{
		ExecutorAddrMap: executorAddrMap,
		curIndexMap:     curIndexMap,
	}
}

func (c *Client) GetServerAddr(dsName string) string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	executorAddrList, ok := c.ExecutorAddrMap[dsName]
	if !ok {
		if alt, ok := c.ExecutorAddrMap[ALL]; ok {
			executorAddrList = alt
		} else {
			log.Fatalf("GetExecutorAddr: dsName %v not found in ExecutorAddrMap", dsName)
		}
	}

	curIndex, ok := c.curIndexMap[dsName]
	if !ok {
		if alt, ok := c.curIndexMap[ALL]; ok {
			curIndex = alt
		} else {
			log.Fatalf("GetExecutorAddr: dsName %v not found in curIndexMap", dsName)
		}
	}

	if curIndex >= len(executorAddrList) {
		curIndex = 0
	}
	addr := executorAddrList[curIndex]
	c.curIndexMap[dsName] = curIndex + 1
	return addr
}

func (c *Client) Read(dsName string, key string, ts int64, cfg txn.RecordConfig) (txn.DataItem, txn.RemoteDataStrategy, string, error) {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}

	data := ReadRequest{
		DsName:    dsName,
		Key:       key,
		StartTime: ts,
		Config:    cfg,
	}
	jsonData, _ := json2.Marshal(data)

	reqUrl := c.GetServerAddr(dsName) + "/read"

	// Create a new POST request using fasthttp
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(reqUrl)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBody(jsonData)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := fasthttp.Do(req, resp)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, txn.Normal, "", errors.New("unexpected status code")
	}

	body := resp.Body()

	var response ReadResponse
	err = json2.Unmarshal(body, &response)
	if err != nil {
		log.Fatal(err)
	}

	if response.Status == "OK" {
		return response.Data, response.DataStrategy, response.GroupKey, nil
	} else {
		errMsg := response.ErrMsg
		return nil, txn.Normal, "", errors.New(errMsg)
	}
}

func (c *Client) Prepare(dsName string, itemList []txn.DataItem,
	startTime int64, cfg txn.RecordConfig,
	validationMap map[string]txn.PredicateInfo) (map[string]string, int64, error) {
	debugStart := time.Now()

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}

	data := PrepareRequest{
		DsName:        dsName,
		ItemType:      GetItemType(dsName),
		ItemList:      itemList,
		StartTime:     startTime,
		Config:        cfg,
		ValidationMap: validationMap,
	}

	jsonData, err := json2.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("Prepare request(JSON DATA): %v\n", string(jsonData))

	reqUrl := c.GetServerAddr(dsName) + "/prepare"

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(reqUrl)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBody(jsonData)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	debugMsg := fmt.Sprintf("HttpClient.Do(Prepare) in %v", dsName)
	logger.Log.Debugw("Before "+debugMsg, "LatencyInFunc", time.Since(debugStart), "Topic", "CheckPoint")
	err = fasthttp.Do(req, resp)
	logger.Log.Debugw("After "+debugMsg, "LatencyInFunc", time.Since(debugStart), "Topic", "CheckPoint")
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, 0, errors.New("unexpected status code")
	}

	body := resp.Body()

	var response PrepareResponse
	err = json2.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Prepare call resp Unmarshal error: %v\nbody:\n%v", err, string(body))
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
	jsonData, _ := json2.Marshal(data)

	reqUrl := c.GetServerAddr(dsName) + "/commit"

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(reqUrl)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBody(jsonData)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := fasthttp.Do(req, resp)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return errors.New("unexpected status code")
	}

	body := resp.Body()

	var response Response[string]
	err = json2.Unmarshal(body, &response)
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

func (c *Client) Abort(dsName string, keyList []string, groupKeyList string) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}

	data := AbortRequest{
		DsName:       dsName,
		KeyList:      keyList,
		GroupKeyList: groupKeyList,
	}
	jsonData, _ := json2.Marshal(data)

	reqUrl := c.GetServerAddr(dsName) + "/abort"

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(reqUrl)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBody(jsonData)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := fasthttp.Do(req, resp)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return errors.New("unexpected status code")
	}

	body := resp.Body()

	var response Response[string]
	err = json2.Unmarshal(body, &response)
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
