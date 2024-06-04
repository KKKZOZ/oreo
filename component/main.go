package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/trace"
	"time"

	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/network"
	"github.com/kkkzoz/oreo/pkg/serializer"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var banner = `
 ____  _        _       _               
/ ___|| |_ __ _| |_ ___| | ___  ___ ___ 
\___ \| __/ _| | __/ _ \ |/ _ \/ __/ __|
 ___) | || (_| | ||  __/ |  __/\__ \__ \
|____/ \__\__,_|\__\___|_|\___||___/___/
`

type Server struct {
	port      int
	reader    network.Reader
	committer network.Committer
}

func NewServer(port int, conn txn.Connector, factory txn.DataItemFactory) *Server {
	return &Server{
		port:      port,
		reader:    *network.NewReader(conn, factory, serializer.NewJSONSerializer()),
		committer: *network.NewCommitter(conn, serializer.NewJSONSerializer(), factory),
	}
}

func (s *Server) Run() {
	router := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/ping":
			s.pingHandler(ctx)
		case "/read":
			s.readHandler(ctx)
		case "/prepare":
			s.prepareHandler(ctx)
		case "/commit":
			s.commitHandler(ctx)
		case "/abort":
			s.abortHandler(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	address := fmt.Sprintf(":%d", s.port)
	fmt.Println(banner)
	Log.Infow("Server running", "address", address)
	log.Fatalf("Server failed: %v", fasthttp.ListenAndServe(address, router))
}

func (s *Server) pingHandler(ctx *fasthttp.RequestCtx) {
	ctx.WriteString("pong")
}

func (s *Server) readHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		Log.Infow("Read request", "latency", time.Since(startTime))
	}()

	var req network.ReadRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		errMsg := fmt.Sprintf("Invalid timestamp parameter: %s", err.Error())
		ctx.Error(errMsg, fasthttp.StatusBadRequest)
		return
	}

	item, dataType, err := s.reader.Read(req.Key, req.StartTime, req.Config, true)

	var response network.ReadResponse
	if err != nil {
		response = network.ReadResponse{
			Status: "Error",
			ErrMsg: err.Error(),
		}
	} else {
		redisItem, ok := item.(*redis.RedisItem)
		if !ok {
			response = network.ReadResponse{
				Status: "Error",
				ErrMsg: "unexpected data type",
			}
		} else {
			response = network.ReadResponse{
				Status:   "OK",
				DataType: dataType,
				Data:     redisItem,
			}
		}
	}
	respBytes, _ := json.Marshal(response)
	ctx.Write(respBytes)
}

func (s *Server) prepareHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		Log.Infow("Prepare request", "latency", time.Since(startTime))
	}()

	var req network.PrepareRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.Error("Invalid request body.", fasthttp.StatusBadRequest)
		return
	}

	var itemList []txn.DataItem
	for _, item := range req.ItemList {
		i := item
		itemList = append(itemList, &i)
	}

	verMap, err := s.committer.Prepare(itemList, req.StartTime, req.CommitTime, req.Config)
	var resp network.Response[map[string]string]
	if err != nil {
		resp = network.Response[map[string]string]{
			Status: "Error",
			ErrMsg: err.Error(),
		}
	} else {
		resp = network.Response[map[string]string]{
			Status: "OK",
			Data:   verMap,
		}
	}
	respBytes, _ := json.Marshal(resp)
	ctx.Write(respBytes)
}

func (s *Server) commitHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		Log.Infow("Commit request", "latency", time.Since(startTime))
	}()

	var req network.CommitRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.Error("Invalid request body.", fasthttp.StatusBadRequest)
		return
	}

	err := s.committer.Commit(req.List)
	var resp network.Response[string]
	if err != nil {
		resp = network.Response[string]{
			Status: "Error",
			ErrMsg: err.Error(),
		}
	} else {
		resp = network.Response[string]{
			Status: "OK",
		}
	}
	respBytes, _ := json.Marshal(resp)
	ctx.Write(respBytes)
}

func (s *Server) abortHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		Log.Infow("Abort request", "latency", time.Since(startTime))
	}()

	var req network.AbortRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.Error("Invalid request body.", fasthttp.StatusBadRequest)
		return
	}

	err := s.committer.Abort(req.KeyList, req.TxnId)
	var resp network.Response[string]
	if err != nil {
		resp = network.Response[string]{
			Status: "Error",
			ErrMsg: err.Error(),
		}
	} else {
		resp = network.Response[string]{
			Status: "OK",
		}
	}
	respBytes, _ := json.Marshal(resp)
	ctx.Write(respBytes)
}

var port = 8000
var poolSize = 60
var traceFlag = false

var Log *zap.SugaredLogger

func main() {
	flag.IntVar(&port, "p", 8000, "Server Port")
	flag.IntVar(&poolSize, "s", 60, "Pool Size")
	flag.BoolVar(&traceFlag, "trace", false, "Enable trace")
	flag.Parse()
	newLogger()

	if traceFlag {
		f, err := os.Create("trace.out")
		if err != nil {
			panic(err)
		}
		err = trace.Start(f)
		if err != nil {
			panic(err)
		}
		defer trace.Stop()
	}

	redisConn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address:  "localhost:6380",
		Password: "@ljy123456",
		PoolSize: poolSize,
	})
	server := NewServer(port, redisConn, &redis.RedisItemFactory{})
	server.Run()
}

func newLogger() {
	conf := zap.NewDevelopmentConfig()
	conf.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	conf.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	conf.EncoderConfig.MessageKey = "msg"
	logger, _ := conf.Build()
	Log = logger.Sugar()
}
