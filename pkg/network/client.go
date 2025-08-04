package network

import (
	"errors"
	"fmt"
	"time"

	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/discovery"
	"github.com/kkkzoz/oreo/pkg/logger" // Use provided logger for client ops
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/valyala/fasthttp"
)

// Ensure RegistryClient implements the RemoteClient interface
var _ txn.RemoteClient = (*Client)(nil)

// --- Registry Data Structures ---

// Client embeds both the registry server functionality and the client
// logic for discovering and communicating with executor instances.
// Refactored Client that supports multiple service discovery methods
type Client struct {
	httpClient       *fasthttp.Client
	serviceDiscovery discovery.ServiceDiscovery
}

// Constants
const (
	ALL = "ALL" // Special DsName indicating an instance handles all datastores
	// Default timeout for requests to executors if not configured otherwise
	defaultRequestTimeout = 30000 * time.Millisecond
)

// Add a unified client creation function
func NewClient(config *discovery.ServiceDiscoveryConfig) (*Client, error) {
	if config == nil {
		config = &discovery.ServiceDiscoveryConfig{
			Type: discovery.HTTPDiscovery,
			HTTP: &discovery.HTTPDiscoveryConfig{
				RegistryPort: ":9000",
			},
		}
	}

	sd, err := createServiceDiscovery(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		httpClient:       &fasthttp.Client{},
		serviceDiscovery: sd,
	}, nil
}

// GetServerAddr returns a server address for the given datastore name
func (rc *Client) GetServerAddr(dsName string) (string, error) {
	if rc == nil {
		return "", fmt.Errorf("client is nil")
	}
	if rc.serviceDiscovery == nil {
		return "", fmt.Errorf("no service discovery configured")
	}
	return rc.serviceDiscovery.GetService(dsName)
}

// Helper to get timeout value (can be extended to read from config)
func getRequestTimeout() time.Duration {
	// TODO: Read this value from config.System.NetworkRequestTimeout if available
	// For now, use the default.
	// Example:
	// if config.System.NetworkRequestTimeout > 0 {
	//     return config.System.NetworkRequestTimeout
	// }
	return defaultRequestTimeout
}

// Read sends a read request with timeout.
func (rc *Client) Read(
	dsName string,
	key string,
	ts int64,
	cfg txn.RecordConfig,
) (txn.DataItem, txn.RemoteDataStrategy, string, error) {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}

	addr, err := rc.GetServerAddr(dsName)
	if err != nil {
		return nil, txn.Normal, "", fmt.Errorf(
			"failed to get executor address for read dsName '%s': %w",
			dsName,
			err,
		)
	}
	reqUrl := "http://" + addr + "/read"
	logger.Log.Debugw("Executing Read request", "url", reqUrl, "dsName", dsName, "key", key)

	reqData := ReadRequest{DsName: dsName, Key: key, StartTime: ts, Config: cfg}
	jsonData, err := json2.Marshal(reqData)
	if err != nil {
		logger.Log.Errorw("Failed to marshal Read request body", "error", err)
		return nil, txn.Normal, "", fmt.Errorf("failed to marshal read request: %w", err)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(reqUrl)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBody(jsonData)

	// Execute with timeout
	timeout := getRequestTimeout()
	err = rc.httpClient.DoTimeout(req, resp, timeout)
	if err != nil {
		// Check specifically for timeout error
		if errors.Is(err, fasthttp.ErrTimeout) {
			logger.Log.Errorw(
				"Read HTTP request timed out",
				"url",
				reqUrl,
				"timeout",
				timeout,
				"error",
				err,
			)
			return nil, txn.Normal, "", fmt.Errorf(
				"request to executor %s timed out after %v: %w",
				reqUrl,
				timeout,
				err,
			)
		}
		// Handle other potential errors (connection refused, DNS error, etc.)
		logger.Log.Errorw("Failed to execute Read HTTP request", "url", reqUrl, "error", err)
		return nil, txn.Normal, "", fmt.Errorf(
			"http request to executor %s failed: %w",
			reqUrl,
			err,
		)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		errMsg := fmt.Sprintf("executor %s returned status %d for read", addr, resp.StatusCode())
		logger.Log.Warnw(errMsg, "url", reqUrl, "responseBody", string(resp.Body()))
		// return nil, txn.Normal, "", errors.New(errMsg)
	}

	var response ReadResponse
	err = json2.Unmarshal(resp.Body(), &response)
	if err != nil {
		logger.Log.Errorw(
			"Failed to unmarshal Read response body",
			"url",
			reqUrl,
			"body",
			string(resp.Body()),
			"error",
			err,
		)
		return nil, txn.Normal, "", fmt.Errorf("unmarshal read response error: %w", err)
	}

	if response.Status == "OK" {
		return response.Data, response.DataStrategy, response.GroupKey, nil
	} else {
		errMsg := response.ErrMsg
		logger.Log.Warnw("Read operation failed on executor (application error)", "url", reqUrl, "error", errMsg)
		return nil, txn.Normal, "", errors.New(errMsg)
	}
}

// Prepare sends a prepare request with timeout.
func (rc *Client) Prepare(dsName string, itemList []txn.DataItem,
	startTime int64, cfg txn.RecordConfig,
	validationMap map[string]txn.PredicateInfo,
) (map[string]string, int64, error) {
	debugStart := time.Now()
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}

	addr, err := rc.GetServerAddr(dsName)
	if err != nil {
		return nil, 0, fmt.Errorf(
			"failed to get executor address for prepare dsName '%s': %w",
			dsName,
			err,
		)
	}
	reqUrl := "http://" + addr + "/prepare"
	logger.Log.Debugw(
		"Executing Prepare request",
		"url",
		reqUrl,
		"dsName",
		dsName,
		"itemCount",
		len(itemList),
	)

	reqData := PrepareRequest{
		DsName:        dsName,
		ItemType:      GetItemType(dsName),
		ItemList:      itemList,
		StartTime:     startTime,
		Config:        cfg,
		ValidationMap: validationMap,
	}
	jsonData, err := json2.Marshal(reqData)
	if err != nil {
		logger.Log.Errorw("Failed to marshal Prepare request body", "error", err)
		return nil, 0, fmt.Errorf("failed to marshal prepare request: %w", err)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(reqUrl)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBody(jsonData)

	// Execute with timeout
	timeout := getRequestTimeout()
	debugMsg := fmt.Sprintf("HttpClient.DoTimeout(Prepare) to %s", reqUrl)
	logger.Log.Debugw(
		"Before "+debugMsg,
		"LatencyInFunc",
		time.Since(debugStart),
		"Topic",
		"CheckPoint",
		"Timeout",
		timeout,
	)
	err = rc.httpClient.DoTimeout(req, resp, timeout)
	logger.Log.Debugw(
		"After "+debugMsg,
		"LatencyInFunc",
		time.Since(debugStart),
		"Topic",
		"CheckPoint",
	)
	if err != nil {
		if errors.Is(err, fasthttp.ErrTimeout) {
			logger.Log.Errorw(
				"Prepare HTTP request timed out",
				"url",
				reqUrl,
				"timeout",
				timeout,
				"error",
				err,
			)
			return nil, 0, fmt.Errorf(
				"request to executor %s timed out after %v: %w",
				reqUrl,
				timeout,
				err,
			)
		}
		logger.Log.Errorw("Failed to execute Prepare HTTP request", "url", reqUrl, "error", err)
		return nil, 0, fmt.Errorf("http request to executor %s failed: %w", reqUrl, err)
	}

	// if resp.StatusCode() != fasthttp.StatusOK {
	// 	errMsg := fmt.Sprintf("executor %s returned status %d for prepare,err: %s", addr, resp.StatusCode(), string(resp.Body()))
	// 	logger.Log.Warnw(errMsg, "url", reqUrl, "responseBody", string(resp.Body()))
	// 	return nil, 0, errors.New(errMsg)
	// }

	var response PrepareResponse
	err = json2.Unmarshal(resp.Body(), &response)
	if err != nil {
		logger.Log.Errorw(
			"Failed to unmarshal Prepare response body",
			"url",
			reqUrl,
			"body",
			string(resp.Body()),
			"error",
			err,
		)
		return nil, 0, fmt.Errorf("unmarshal prepare response error: %w", err)
	}

	if response.Status == "OK" {
		return response.VerMap, response.TCommit, nil
	} else {
		errMsg := response.ErrMsg
		logger.Log.Warnw("Prepare operation failed on executor (application error)", "url", reqUrl, "error", errMsg)
		return nil, 0, errors.New(errMsg)
	}
}

// Commit sends a commit request with timeout.
func (rc *Client) Commit(dsName string, infoList []txn.CommitInfo, tCommit int64) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}

	addr, err := rc.GetServerAddr(dsName)
	if err != nil {
		return fmt.Errorf("failed to get executor address for commit dsName '%s': %w", dsName, err)
	}
	reqUrl := "http://" + addr + "/commit"
	logger.Log.Debugw(
		"Executing Commit request",
		"url",
		reqUrl,
		"dsName",
		dsName,
		"infoCount",
		len(infoList),
	)

	reqData := CommitRequest{DsName: dsName, List: infoList, TCommit: tCommit}
	jsonData, err := json2.Marshal(reqData)
	if err != nil {
		logger.Log.Errorw("Failed to marshal Commit request body", "error", err)
		return fmt.Errorf("failed to marshal commit request: %w", err)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(reqUrl)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBody(jsonData)

	// Execute with timeout
	timeout := getRequestTimeout()
	err = rc.httpClient.DoTimeout(req, resp, timeout)
	if err != nil {
		if errors.Is(err, fasthttp.ErrTimeout) {
			logger.Log.Errorw(
				"Commit HTTP request timed out",
				"url",
				reqUrl,
				"timeout",
				timeout,
				"error",
				err,
			)
			return fmt.Errorf("request to executor %s timed out after %v: %w", reqUrl, timeout, err)
		}
		logger.Log.Errorw("Failed to execute Commit HTTP request", "url", reqUrl, "error", err)
		return fmt.Errorf("http request to executor %s failed: %w", reqUrl, err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		errMsg := fmt.Sprintf("executor %s returned status %d for commit", addr, resp.StatusCode())
		logger.Log.Warnw(errMsg, "url", reqUrl, "responseBody", string(resp.Body()))
		return errors.New(errMsg)
	}

	var response Response[string] // Generic response structure
	err = json2.Unmarshal(resp.Body(), &response)
	if err != nil {
		logger.Log.Errorw(
			"Failed to unmarshal Commit response body",
			"url",
			reqUrl,
			"body",
			string(resp.Body()),
			"error",
			err,
		)
		return fmt.Errorf("unmarshal commit response error: %w", err)
	}

	if response.Status == "OK" {
		return nil
	} else {
		errMsg := response.ErrMsg
		logger.Log.Warnw("Commit operation failed on executor (application error)", "url", reqUrl, "error", errMsg)
		return errors.New(errMsg)
	}
}

// Abort sends an abort request with timeout.
func (rc *Client) Abort(dsName string, keyList []string, groupKeyList string) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}

	addr, err := rc.GetServerAddr(dsName)
	if err != nil {
		return fmt.Errorf("failed to get executor address for abort dsName '%s': %w", dsName, err)
	}
	reqUrl := "http://" + addr + "/abort"
	logger.Log.Debugw(
		"Executing Abort request",
		"url",
		reqUrl,
		"dsName",
		dsName,
		"keyCount",
		len(keyList),
	)

	reqData := AbortRequest{DsName: dsName, KeyList: keyList, GroupKeyList: groupKeyList}
	jsonData, err := json2.Marshal(reqData)
	if err != nil {
		logger.Log.Errorw("Failed to marshal Abort request body", "error", err)
		return fmt.Errorf("failed to marshal abort request: %w", err)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(reqUrl)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBody(jsonData)

	// Execute with timeout
	timeout := getRequestTimeout()
	err = rc.httpClient.DoTimeout(req, resp, timeout)
	if err != nil {
		if errors.Is(err, fasthttp.ErrTimeout) {
			logger.Log.Errorw(
				"Abort HTTP request timed out",
				"url",
				reqUrl,
				"timeout",
				timeout,
				"error",
				err,
			)
			return fmt.Errorf("request to executor %s timed out after %v: %w", reqUrl, timeout, err)
		}
		logger.Log.Errorw("Failed to execute Abort HTTP request", "url", reqUrl, "error", err)
		return fmt.Errorf("http request to executor %s failed: %w", reqUrl, err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		errMsg := fmt.Sprintf("executor %s returned status %d for abort", addr, resp.StatusCode())
		logger.Log.Warnw(errMsg, "url", reqUrl, "responseBody", string(resp.Body()))
		return errors.New(errMsg)
	}

	var response Response[string] // Generic response structure
	err = json2.Unmarshal(resp.Body(), &response)
	if err != nil {
		logger.Log.Errorw(
			"Failed to unmarshal Abort response body",
			"url",
			reqUrl,
			"body",
			string(resp.Body()),
			"error",
			err,
		)
		return fmt.Errorf("unmarshal abort response error: %w", err)
	}

	if response.Status == "OK" {
		return nil
	} else {
		errMsg := response.ErrMsg
		logger.Log.Warnw("Abort operation failed on executor (application error)", "url", reqUrl, "error", errMsg)
		return errors.New(errMsg)
	}
}

func createServiceDiscovery(
	config *discovery.ServiceDiscoveryConfig,
) (discovery.ServiceDiscovery, error) {
	switch config.Type {
	case discovery.HTTPDiscovery:
		return discovery.NewHTTPServiceDiscovery(
			config.HTTP.RegistryPort,
			config.HTTP.RegistryServerURL,
		)
	case discovery.EtcdDiscovery:
		registryConfig := discovery.DefaultRegistryConfig()
		return discovery.NewEtcdServiceDiscovery(
			config.Etcd.Endpoints,
			config.Etcd.KeyPrefix,
			registryConfig,
		)
	default:
		return nil, fmt.Errorf("unsupported service discovery type: %s", config.Type)
	}
}
