package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/kkkzoz/oreo/pkg/datastore/cassandra"
	"github.com/kkkzoz/oreo/pkg/txn"
)

func main() {
	// 生成随机 key
	key := randomString()
	cassandraItem := cassandra.NewCassandraItem(txn.ItemOptions{
		Key:          key,
		Value:        "value1",
		GroupKeyList: "txn1",
		TValid:       time.Now().UnixMicro(),
		TLease:       time.Now(),
	})

	// 创建连接
	conn := cassandra.NewCassandraConnection(&cassandra.ConnectionOptions{
		Hosts:    []string{"localhost"},
		Keyspace: "oreo",
	})
	err := conn.Connect()
	if err != nil {
		panic(err)
	}

	// 写入数据
	_, err = conn.PutItem(cassandraItem.Key(), cassandraItem)
	if err != nil {
		panic(err)
	}
	fmt.Println("First put success")

	// 读取数据
	resItem, err := conn.GetItem(cassandraItem.Key())
	if err != nil {
		panic(err)
	}
	fmt.Println(resItem)

	// 更新数据
	resItem.SetValue("value2")
	_, err = conn.PutItem(cassandraItem.Key(), resItem)
	if err != nil {
		panic(err)
	}
	resItem, err = conn.GetItem(cassandraItem.Key())
	if err != nil {
		panic(err)
	}
	fmt.Println("Update success")

	// 并发更新测试
	num := 100
	var wg sync.WaitGroup
	wg.Add(num)
	successNum := 0
	var mu sync.Mutex

	for i := 0; i < num; i++ {
		item := resItem
		go func(ii int) {
			defer wg.Done()
			idx := fmt.Sprintf("%d", ii)
			item.SetValue(idx)
			newVer, err := conn.ConditionalUpdate(item.Key(), item, false)
			if err == nil {
				mu.Lock()
				successNum++
				fmt.Printf("success: %s\n", newVer)
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("successNum: %d\n", successNum)
}

func randomString() string {
	return "item" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
