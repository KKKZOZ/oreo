package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"
)

var (
	// 命令行参数
	port                       int
	physicalTimeUpdateInterval int
	logicalTimeBits            int

	// 逻辑时间的最大值，2^logicalTimeBits - 1
	maxLogicalTime int64

	physicalTime int64 // 物理时间 (精确到毫秒)
	logicalTime  int64 // 逻辑时间
	mu           sync.Mutex
)

// 定时器：每 physicalTimeUpdateInterval ms 更新物理时间，或者在收到通知时立即更新
func updatePhysicalTime() {
	ticker := time.NewTicker(time.Duration(physicalTimeUpdateInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		<-ticker.C
		mu.Lock()
		physicalTime = time.Now().UnixMilli()
		logicalTime = 0
		mu.Unlock()
	}
}

// 获取当前时间戳，包括物理时间和逻辑时间
func getTimestamp() int64 {
	mu.Lock()
	defer mu.Unlock()

	// 如果逻辑时间即将超过上限，更新物理时间并重置逻辑时间
	if logicalTime >= maxLogicalTime-10 {
		// fmt.Printf("Triggered physical time update\n")
		physicalTime = time.Now().UnixMilli() // 更新物理时间
		logicalTime = 0                       // 重置逻辑时间
	}
	logicalTime++

	// 将物理时间和逻辑时间打包成一个 int64
	timestamp := physicalTime*int64(math.Pow10(logicalTimeBits)) + logicalTime
	return timestamp
	// return (physicalTime << logicalTimeBits) | logicalTime

}

// 处理 HTTP 请求，返回时间戳
func handleTimestamp(w http.ResponseWriter, r *http.Request) {
	timestamp := getTimestamp()
	w.Write([]byte(fmt.Sprintf("%d", timestamp)))
}

func main() {
	// 使用 flag 包解析命令行参数
	flag.IntVar(&port, "port", 8010, "HTTP server port number")
	flag.IntVar(&physicalTimeUpdateInterval, "interval", 1, "Physical time update interval in milliseconds")
	flag.IntVar(&logicalTimeBits, "bits", 6, "Number of bits for logical time")
	flag.Parse()

	// 计算逻辑时间的最大值
	maxLogicalTime = (1 << logicalTimeBits) - 1

	// 初始化物理时间
	physicalTime = time.Now().UnixMilli()

	// 启动更新物理时间的 goroutine
	go updatePhysicalTime()

	// 设置 HTTP handler
	http.HandleFunc("/timestamp/", handleTimestamp)

	// 启动 HTTP server
	serverAddress := fmt.Sprintf(":%d", port)
	fmt.Printf("Server listening on %s\n", serverAddress)
	log.Fatal(http.ListenAndServe(serverAddress, nil))
}
