package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/oreo-dtx-lab/oreo/pkg/datastore/tikv"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

func main() {
	// 生成随机 key
	key := randomString()
	tikvItem := tikv.NewTiKVItem(txn.ItemOptions{
		Key:          key,
		Value:        "value1",
		GroupKeyList: "txn1",
		TValid:       time.Now().UnixMicro(),
		TLease:       time.Now(),
	})

	// 创建连接
	conn := tikv.NewTiKVConnection(&tikv.ConnectionOptions{
		PDAddrs: []string{"127.0.0.1:2379"},
	})
	err := conn.Connect()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Connected to TiKV\n\n\n\n")
	time.Sleep(1 * time.Second)
	fmt.Printf("Start testing\n")

	// 写入数据
	_, err = conn.ConditionalUpdate(tikvItem.Key(), tikvItem, true)
	if err != nil {
		panic(err)
	}
	fmt.Println("First put success")

	// 读取数据
	resItem, err := conn.GetItem(tikvItem.Key())
	if err != nil {
		panic(err)
	}
	fmt.Println(resItem)

	resItemStr, _ := json.Marshal(resItem)

	// 并发更新测试
	num := 100
	var wg sync.WaitGroup
	wg.Add(num)
	successNum := 0
	var mu sync.Mutex

	for i := 0; i < num; i++ {
		item := &tikv.TiKVItem{
			KKey:          resItem.Key(),
			KValue:        resItem.Value(),
			KGroupKeyList: resItem.GroupKeyList(),
			KTxnState:     resItem.TxnState(),
			KTValid:       resItem.TValid(),
			KTLease:       resItem.TLease(),
			KPrev:         resItem.Prev(),
			KLinkedLen:    resItem.LinkedLen(),
			KIsDeleted:    resItem.IsDeleted(),
			KVersion:      resItem.Version(),
		}
		go func(ii int) {
			defer wg.Done()
			idx := fmt.Sprintf("value-%d", ii)
			item.SetValue(idx)
			item.SetPrev(string(resItemStr))
			newVer, err := conn.ConditionalUpdate(item.Key(), item, false)
			if err == nil {
				mu.Lock()
				successNum++
				fmt.Printf("success id: %d  newVer: %s\n", ii, newVer)
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("successNum: %d\n", successNum)

	resItem, err = conn.GetItem(tikvItem.Key())
	if err != nil {
		panic(err)
	}
	fmt.Println(resItem)

	// KV Test
	// Test Conn.Get() and Conn.AtomicCreate()
	successNum = 0
	wg.Add(num)
	key = "GroupKey"
	_ = conn.Delete(key)

	for i := 0; i < num; i++ {
		go func(idx int) {
			defer wg.Done()
			value := fmt.Sprintf("value-%d", idx)
			_, err := conn.AtomicCreate(key, value)
			if err == nil {
				mu.Lock()
				successNum++
				fmt.Printf("success ID: %d\n", idx)
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("successNum: %d\n", successNum)

	value, err := conn.Get(key)
	if err != nil {
		log.Fatalf("Get failed: %v", err)
	}
	fmt.Printf("Get value: %s\n", value)
}

func randomString() string {
	return "item" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
