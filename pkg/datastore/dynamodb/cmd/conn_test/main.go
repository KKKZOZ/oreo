package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/dynamodb"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

func main() {
	// 生成随机 key
	key := randomString()
	dynamoItem := dynamodb.NewDynamoDBItem(txn.ItemOptions{
		Key:          key,
		Value:        "value1",
		GroupKeyList: "txn1",
		TValid:       time.Now().UnixMicro(),
		TLease:       time.Now(),
	})

	// 创建连接，使用 dynamodb-local
	conn := dynamodb.NewDynamoDBConnection(&dynamodb.ConnectionOptions{
		Region:    "",
		TableName: "oreo",
		Endpoint:  "http://localhost:8000", // DynamoDB Local endpoint
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			"dummy",
			"dummy",
			"",
		)),
	})
	err := conn.Connect()
	if err != nil {
		panic(err)
	}

	// 写入数据
	_, err = conn.PutItem(dynamoItem.Key(), dynamoItem)
	if err != nil {
		panic(err)
	}
	fmt.Println("First put success")

	// 读取数据
	resItem, err := conn.GetItem(dynamoItem.Key())
	if err != nil {
		panic(err)
	}
	fmt.Println(resItem)

	// 更新数据
	resItem.SetValue("value2")
	_, err = conn.PutItem(dynamoItem.Key(), resItem)
	if err != nil {
		panic(err)
	}
	resItem, err = conn.GetItem(dynamoItem.Key())
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

	// KV Test
	// Test Conn.Get() and Conn.AtomicCreate()
	successNum = 0
	wg.Add(num)
	key = "GroupKey"
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
