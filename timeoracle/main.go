package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/oreo-dtx-lab/oreo/pkg/timesource"
)

var port int
var oracleType string

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
	flag.IntVar(&port, "port", 8010, "HTTP server port number")
	flag.StringVar(&oracleType, "type", "hybrid", "Time Oracle Implementaion Type")
	flag.Parse()

	var oracle timesource.TimeSourcer
	switch oracleType {
	case "hybrid":
		oracle = timesource.NewHybridTimeSource(10, 6)
	case "simple":
		oracle = timesource.NewSimpleTimeSource()
	case "counter":
		oracle = timesource.NewCounterTimeSource()
	default:
		panic("Time oracle Implementaion MUST be specified")
	}

	server := TimeOracleServer{
		oracle: oracle,
		port:   port,
	}

	// 设置 HTTP handler，使用 server.handleTimestamp
	http.HandleFunc("/timestamp/", server.handleTimestamp)

	// 启动 HTTP server
	serverAddress := fmt.Sprintf(":%d", server.port)
	fmt.Printf("Server listening on %s\n", serverAddress)
	log.Fatal(http.ListenAndServe(serverAddress, nil))
}
