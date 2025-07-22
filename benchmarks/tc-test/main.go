// combined_client_test.go
package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gocql/gocql" // Cassandra 客户端
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo" // MongoDB 客户端
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	iterations = 3               // 每个测试运行的次数
	opTimeout  = 5 * time.Second // 单个操作的超时时间

	// --- 配置常量 ---
	httpURL       = "http://10.206.206.3:9999"
	redisAddr     = "10.206.206.3:6379"
	redisPassword = "kkkzoz" // 如果没有密码，留空 ""
	mongoURI      = "mongodb://10.206.206.4:27018"
	mongoUsername = "admin"
	mongoPassword = "password"
)

var cassandraHosts = []string{"10.206.206.5"} // Cassandra 节点列表

// --- HTTP 测试函数 ---
func testHTTP(url string, iterations int) {
	log.Printf("Testing HTTP GET to %s...", url)
	client := &http.Client{
		Timeout: 10 * time.Second, // 整体请求超时
	}

	for i := 0; i < iterations; i++ {
		start := time.Now()
		resp, err := client.Get(url)
		if err != nil {
			log.Printf("[HTTP] Iteration %d: Error during GET: %v", i+1, err)
			time.Sleep(1 * time.Second)
			continue
		}

		_, err = io.Copy(io.Discard, resp.Body)
		resp.Body.Close() // 确保关闭 Body
		duration := time.Since(start)

		if err != nil {
			log.Printf("[HTTP] Iteration %d: Error reading response body: %v", i+1, err)
		} else {
			log.Printf("[HTTP] Iteration %d: Request took: %v (Status: %s)", i+1, duration, resp.Status)
		}
		time.Sleep(500 * time.Millisecond) // 稍微间隔
	}
}

// --- Redis 测试函数 ---
func testRedis(addr, password string, iterations int) {
	log.Printf("Testing Redis connection to %s...", addr)
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0, // use default DB
	})

	ctxInitial, cancelInitial := context.WithTimeout(context.Background(), opTimeout)
	defer cancelInitial()
	_, err := rdb.Ping(ctxInitial).Result()
	if err != nil {
		log.Printf("[Redis] Failed to connect: %v. Skipping Redis tests.", err)
		return // 连接失败则跳过后续测试
	}
	log.Println("[Redis] Connection successful.")

	for i := 0; i < iterations; i++ {
		ctxOp, cancelOp := context.WithTimeout(context.Background(), opTimeout)

		start := time.Now()
		pong, err := rdb.Ping(ctxOp).Result() // 使用 PING 作为测试命令
		duration := time.Since(start)

		cancelOp() // 及时释放 context

		if err != nil {
			log.Printf("[Redis] Iteration %d: Error during PING: %v", i+1, err)
		} else {
			log.Printf("[Redis] Iteration %d: PING took: %v, Result: %s", i+1, duration, pong)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// --- MongoDB 测试函数 ---
func testMongoDB(uri, username, password string, iterations int) {
	log.Printf("Testing MongoDB connection to %s...", uri)

	credential := options.Credential{
		Username: username,
		Password: password,
		// AuthSource: "admin", // 如果需要指定认证数据库
	}
	clientOptions := options.Client().ApplyURI(uri).SetAuth(credential)

	ctxConnect, cancelConnect := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelConnect()
	client, err := mongo.Connect(ctxConnect, clientOptions)
	if err != nil {
		log.Printf("[MongoDB] Failed to create client: %v. Skipping MongoDB tests.", err)
		return
	}

	// 设置断开连接
	defer func() {
		log.Println("[MongoDB] Disconnecting...")
		ctxDisconnect, cancelDisconnect := context.WithTimeout(context.Background(), opTimeout)
		defer cancelDisconnect()
		if err := client.Disconnect(ctxDisconnect); err != nil {
			log.Printf("[MongoDB] Error disconnecting: %v", err)
		}
	}()

	// 检查初始连接
	ctxPing, cancelPing := context.WithTimeout(context.Background(), opTimeout)
	defer cancelPing()
	err = client.Ping(ctxPing, readpref.Primary())
	if err != nil {
		log.Printf("[MongoDB] Failed to connect (ping failed): %v. Skipping MongoDB tests.", err)
		return
	}
	log.Println("[MongoDB] Connection successful.")

	for i := 0; i < iterations; i++ {
		ctxOp, cancelOp := context.WithTimeout(context.Background(), opTimeout)

		start := time.Now()
		err = client.Ping(ctxOp, readpref.Primary()) // 使用 Ping 作为测试操作
		duration := time.Since(start)

		cancelOp() // 及时释放 context

		if err != nil {
			log.Printf("[MongoDB] Iteration %d: Error during Ping: %v", i+1, err)
		} else {
			log.Printf("[MongoDB] Iteration %d: Ping took: %v", i+1, duration)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// --- Cassandra 测试函数 ---
func testCassandra(hosts []string, iterations int) {
	log.Printf("Testing Cassandra connection to %v...", hosts)

	cluster := gocql.NewCluster(hosts...)
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = opTimeout
	// 如果需要认证或指定 Keyspace 在这里配置:
	// cluster.Authenticator = gocql.PasswordAuthenticator{Username: "...", Password: "..."}
	// cluster.Keyspace = "..."

	session, err := cluster.CreateSession()
	if err != nil {
		log.Printf("[Cassandra] Failed to connect: %v. Skipping Cassandra tests.", err)
		return
	}
	defer session.Close()
	log.Println("[Cassandra] Connection successful.")

	query := "SELECT release_version FROM system.local" // 简单的测试查询

	for i := 0; i < iterations; i++ {
		start := time.Now()
		iter := session.Query(query).Iter()
		var releaseVersion string
		scanSuccessful := iter.Scan(&releaseVersion) // 尝试读取结果
		err := iter.Close()                          // 必须关闭 iterator 并检查错误
		duration := time.Since(start)

		if err != nil {
			log.Printf(
				"[Cassandra] Iteration %d: Error during query or closing iterator: %v",
				i+1,
				err,
			)
		} else if !scanSuccessful && err == nil { // Scan 可能未成功但 Close 没有错误 (例如表空)
			log.Printf("[Cassandra] Iteration %d: Query (%s) took: %v, but no rows returned/scanned.", i+1, query, duration)
		} else {
			log.Printf("[Cassandra] Iteration %d: Query (%s) took: %v, Version: %s", i+1, query, duration, releaseVersion)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// --- Main 函数 ---
func main() {
	log.Println("=============================================")
	log.Println("Starting Combined Network Latency Tests")
	log.Println("=============================================")

	log.Println("\n--- Starting HTTP Test ---")
	testHTTP(httpURL, iterations)
	log.Println("--- Finished HTTP Test ---")

	log.Println("\n--- Starting Redis Test ---")
	testRedis(redisAddr, redisPassword, iterations)
	log.Println("--- Finished Redis Test ---")

	log.Println("\n--- Starting MongoDB Test ---")
	testMongoDB(mongoURI, mongoUsername, mongoPassword, iterations)
	log.Println("--- Finished MongoDB Test ---")

	log.Println("\n--- Starting Cassandra Test ---")
	testCassandra(cassandraHosts, iterations)
	log.Println("--- Finished Cassandra Test ---")

	log.Println("\n=============================================")
	log.Println("All tests completed.")
	log.Println("=============================================")
}
