package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/network"
	"github.com/kkkzoz/oreo/pkg/serializer"
	"github.com/kkkzoz/oreo/pkg/txn"
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
	http.HandleFunc("/ping", s.pingHandler)
	http.HandleFunc("/read", s.readHandler)
	http.HandleFunc("/prepare", s.prepareHandler)
	http.HandleFunc("/commit", s.commitHandler)
	http.HandleFunc("/abort", s.abortHandler)
	address := fmt.Sprintf(":%d", s.port)
	fmt.Println(banner)
	Log.Infow("Server running", "address", address)
	log.Fatalf("Server failed: %v", http.ListenAndServe(address, nil))
}

func (s *Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func (s *Server) readHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer func() {
		Log.Infow("Read request", "latency", time.Since(startTime))
	}()

	requestBody, _ := io.ReadAll(r.Body)
	var req network.ReadRequest
	err := json.Unmarshal(requestBody, &req)
	if err != nil {
		errMsg := fmt.Sprintf("Invalid timestamp parameter: %s", err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// Log.Infow("Read request", "key", req.Key, "start_time", req.StartTime)
	item, err := s.reader.Read(req.Key, req.StartTime, true)

	var response network.ReadResponse
	if err != nil {
		response = network.ReadResponse{
			Status: "Error",
			ErrMsg: err.Error(),
		}
	} else {
		redisItem, ok := item.(*redis.RedisItem)
		if !ok {
			// Handle the case where the type assertion fails
			response = network.ReadResponse{
				Status: "Error",
				ErrMsg: "unexpected data type",
			}
		} else {
			response = network.ReadResponse{
				Status: "OK",
				Data:   redisItem,
			}
		}
	}
	respBytes, _ := json.Marshal(response)
	w.Write(respBytes)
}

func (s *Server) prepareHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer func() {
		Log.Infow("Prepare request", "latency", time.Since(startTime))
	}()

	body, _ := io.ReadAll(r.Body)
	var req network.PrepareRequest
	err := json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
	}

	var itemList []txn.DataItem
	for _, item := range req.ItemList {
		i := item
		itemList = append(itemList, &i)
	}

	// Log.Infow("Prepare request", "item_list", itemList, "start_time",
	// 	req.StartTime, "commit_time", req.CommitTime)

	verMap, err := s.committer.Prepare(itemList, req.StartTime, req.CommitTime)
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
	w.Write(respBytes)
}

func (s *Server) commitHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer func() {
		Log.Infow("Commit request", "latency", time.Since(startTime))
	}()

	body, _ := io.ReadAll(r.Body)
	var req network.CommitRequest
	err := json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
	}

	// Log.Infow("Commit request", "item_list", req.List)
	err = s.committer.Commit(req.List)
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
	w.Write(respBytes)
}

func (s *Server) abortHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer func() {
		Log.Infow("Abort request", "latency", time.Since(startTime))
	}()

	body, _ := io.ReadAll(r.Body)
	var req network.AbortRequest
	err := json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
	}

	// Log.Infow("Abort request", "key_list", req.KeyList, "txn_id", req.TxnId)
	err = s.committer.Abort(req.KeyList, req.TxnId)
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
	w.Write(respBytes)
}

var port = 8000
var poolSize = 60
var maxLen = 2

var Log *zap.SugaredLogger

func main() {

	flag.IntVar(&port, "p", 8000, "Server Port")
	flag.IntVar(&poolSize, "s", 60, "Pool Size")
	flag.IntVar(&maxLen, "m", 2, "Record Max Length")
	flag.Parse()

	newLogger()
	config.Config.MaxRecordLength = maxLen

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
	conf.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	conf.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	conf.EncoderConfig.MessageKey = "msg"
	logger, _ := conf.Build()
	Log = logger.Sugar()
}
