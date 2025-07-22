package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/oreo-dtx-lab/oreo/pkg/timesource"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	port       int
	oracleType string
	Log        *zap.SugaredLogger
)

type TimeOracleServer struct {
	oracle timesource.TimeSourcer
	port   int
}

// 处理 HTTP 请求，返回时间戳
func (t TimeOracleServer) handleTimestamp(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer func() {
		Log.Debugw(
			"handleTimestamp",
			"LatencyInFunction",
			time.Since(startTime).Microseconds(),
			"Topic",
			"CheckPoint",
		)
	}()
	timestamp, _ := t.oracle.GetTime("pattern")
	w.Write([]byte(fmt.Sprintf("%d", timestamp)))
}

func main() {
	flag.IntVar(&port, "p", 8010, "HTTP server port number")
	flag.StringVar(&oracleType, "type", "hybrid", "Time Oracle Implementaion Type")
	flag.Parse()
	newLogger()

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

func newLogger() {
	conf := zap.NewDevelopmentConfig()

	logLevel := os.Getenv("LOG")

	switch logLevel {
	case "DEBUG":
		conf.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "INFO":
		conf.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "WARN":
		conf.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "ERROR":
		conf.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "FATAL":
		conf.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		conf.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}

	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	conf.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	conf.EncoderConfig.MessageKey = "msg"
	logger, _ := conf.Build()
	Log = logger.Sugar()
}
