package tikv

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-errors/errors"
	"github.com/oreo-dtx-lab/oreo/internal/util"
	oreoconfig "github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	"github.com/tikv/client-go/v2/config"
	"github.com/tikv/client-go/v2/rawkv"
)

var _ txn.Connector = (*TiKVConnection)(nil)

type TiKVConnection struct {
	client       *rawkv.Client
	config       ConnectionOptions
	hasConnected bool
}

type ConnectionOptions struct {
	PDAddrs []string
}

func NewTiKVConnection(config *ConnectionOptions) *TiKVConnection {
	if config == nil {
		config = &ConnectionOptions{
			PDAddrs: []string{"127.0.0.1:2379"},
		}
	}
	return &TiKVConnection{
		config:       *config,
		hasConnected: false,
	}
}

func (c *TiKVConnection) Connect() error {
	if c.hasConnected {
		return nil
	}

	client, err := rawkv.NewClient(context.Background(), c.config.PDAddrs, config.Security{})
	if err != nil {
		return err
	}

	client.SetAtomicForCAS(true)
	cfg := config.GetGlobalConfig()
	cfg.TiKVClient.GrpcConnectionCount = 1000
	cfg.TiKVClient.OverloadThreshold = 5000
	cfg.TiKVClient.MaxBatchSize = 1
	cfg.TiKVClient.BatchWaitSize = 1
	cfg.TiKVClient.MaxBatchWaitTime = 0

	config.StoreGlobalConfig(cfg)
	c.client = client
	c.hasConnected = true

	// warm up
	for i := 0; i < 10; i++ {
		_, _ = c.client.Get(context.Background(), []byte("warmup"))
	}

	return nil
}

func (c *TiKVConnection) GetItem(key string) (txn.DataItem, error) {
	if !c.hasConnected {
		return &TiKVItem{}, fmt.Errorf("not connected to TiKV")
	}
	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	value, err := c.client.Get(context.Background(), []byte(key))
	if err != nil {
		return &TiKVItem{}, err
	}
	if value == nil {
		return &TiKVItem{}, errors.New(txn.KeyNotFound)
	}

	var item TiKVItem
	err = json.Unmarshal(value, &item)
	if err != nil {
		return &TiKVItem{}, errors.New("failed to unmarshal item")
	}
	return &item, nil
}

func (c *TiKVConnection) PutItem(key string, value txn.DataItem) (string, error) {
	if !c.hasConnected {
		return "", fmt.Errorf("not connected to TiKV")
	}
	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	data, err := json.Marshal(value)
	if err != nil {
		return "", errors.New("failed to marshal item")
	}

	err = c.client.Put(context.Background(), []byte(key), data)
	if err != nil {
		return "", errors.New(fmt.Sprintf("PutItem key %s failed, err: %v", key, err))
	}
	return "", nil
}

func (c *TiKVConnection) ConditionalUpdate(
	key string,
	value txn.DataItem,
	doCreate bool,
) (string, error) {
	if !c.hasConnected {
		return "", fmt.Errorf("not connected to TiKV")
	}
	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	ctx := context.Background()
	newVer := util.AddToString(value.Version(), 1)
	value.SetVersion(newVer)
	newData, err := json.Marshal(value)
	if err != nil {
		return "", errors.New("failed to marshal item")
	}

	if doCreate {
		if value.Version() != "1" {
			return "", errors.New(txn.VersionMismatch)
		}

		// 使用 CompareAndSwap 确保键不存在时才创建
		_, ok, err := c.client.CompareAndSwap(ctx, []byte(key), nil, newData)
		if err != nil {
			return "", errors.New(
				fmt.Sprintf("ConditionalUpdate(doCreate) key %s failed, err: %v", key, err),
			)
		}
		if ok {
			return newVer, nil
		} else {
			return "", errors.New("key exists")
		}
	}

	// 使用 CompareAndSwap 确保原子更新
	_, ok, err := c.client.CompareAndSwap(ctx, []byte(key), []byte(value.Prev()), newData)
	if err != nil {
		return "", errors.New(fmt.Sprintf("ConditionalUpdate key %s failed, err: %v", key, err))
	}
	if ok {
		// fmt.Printf("ConditionalUpdate key %s success, newVersion: %s, value: %s\n", key, newVer, value)
		return newVer, nil
	} else {
		return "", errors.New(txn.VersionMismatch)
	}
}

func (c *TiKVConnection) ConditionalCommit(
	key string,
	version string,
	tCommit int64,
) (string, error) {
	if !c.hasConnected {
		return "", fmt.Errorf("not connected to TiKV")
	}
	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	ctx := context.Background()

	// 获取当前值
	currentValue, err := c.client.Get(ctx, []byte(key))
	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to get current value: %v", err))
	}
	if currentValue == nil {
		return "", errors.New(txn.KeyNotFound)
	}

	var item TiKVItem
	err = json.Unmarshal(currentValue, &item)
	if err != nil {
		return "", errors.New("failed to unmarshal item")
	}

	if item.Version() != version {
		return "", errors.New(txn.VersionMismatch)
	}

	// 更新状态
	item.KTxnState = oreoconfig.COMMITTED
	item.KTValid = tCommit

	// 序列化新值
	newData, err := json.Marshal(item)
	if err != nil {
		return "", errors.New("failed to marshal item")
	}

	// 原子更新
	_, ok, err := c.client.CompareAndSwap(ctx, []byte(key), currentValue, newData)
	if err != nil {
		return "", errors.New(fmt.Sprintf("ConditionalCommit key %s failed, err: %v", key, err))
	}
	if ok {
		return version, nil
	} else {
		return "", errors.New(txn.VersionMismatch)
	}
}

func (c *TiKVConnection) AtomicCreate(name string, value any) (string, error) {
	if !c.hasConnected {
		return "", fmt.Errorf("not connected to TiKV")
	}
	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	ctx := context.Background()
	strValue := util.ToString(value)

	// 使用 CompareAndSwap 确保原子创建
	_, ok, err := c.client.CompareAndSwap(ctx, []byte(name), nil, []byte(strValue))
	if err != nil {
		return "", err
	}
	if ok {
		return "", nil
	} else {
		return "", errors.New(txn.KeyExists)
	}
}

func (c *TiKVConnection) Get(name string) (string, error) {
	if !c.hasConnected {
		return "", fmt.Errorf("not connected to TiKV")
	}
	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	value, err := c.client.Get(context.Background(), []byte(name))
	if err != nil {
		return "", errors.New(fmt.Sprintf("get key %s failed, err: %v", name, err))
	}
	if value == nil {
		return "", errors.New(txn.KeyNotFound)
	}
	return string(value), nil
}

func (c *TiKVConnection) Put(name string, value interface{}) error {
	if !c.hasConnected {
		return fmt.Errorf("not connected to TiKV")
	}
	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	strValue := util.ToString(value)
	err := c.client.Put(context.Background(), []byte(name), []byte(strValue))
	if err != nil {
		return errors.New(fmt.Sprintf("put key %s failed, err: %v", name, err))
	}
	return nil
}

func (c *TiKVConnection) Delete(name string) error {
	if !c.hasConnected {
		return fmt.Errorf("not connected to TiKV")
	}
	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	err := c.client.Delete(context.Background(), []byte(name))
	if err != nil {
		return errors.New(fmt.Sprintf("delete key %s failed, err: %v", name, err))
	}
	return nil
}
