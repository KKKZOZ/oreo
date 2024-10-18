package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/oreo-dtx-lab/oreo/pkg/timesource"
)

var port int

type TimeOracleServer struct {
	oracle timesource.TimeSourcer
	port   int
}

// 处理 HTTP 请求，返回时间戳
func (t TimeOracleServer) handleTimestamp(w http.ResponseWriter, r *http.Request) {
	timestamp, _ := t.oracle.GetTime("pattern")
	w.Write([]byte(fmt.Sprintf("%d", timestamp)))
}

func main() {
	// 使用 flag 包解析命令行参数
	flag.IntVar(&port, "port", 8010, "HTTP server port number")
	flag.Parse()

	server := TimeOracleServer{
		oracle: timesource.NewSimpleTimeSource(),
		port:   port,
	}

	// 设置 HTTP handler，使用 server.handleTimestamp
	http.HandleFunc("/timestamp/", server.handleTimestamp)

	// 启动 HTTP server
	serverAddress := fmt.Sprintf(":%d", server.port)
	fmt.Printf("Server listening on %s\n", serverAddress)
	log.Fatal(http.ListenAndServe(serverAddress, nil))
}
