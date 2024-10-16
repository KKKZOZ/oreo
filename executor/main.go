package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"
	"runtime/trace"
	"syscall"
	"time"

	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/couchdb"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/mongo"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/redis"
	"github.com/oreo-dtx-lab/oreo/pkg/network"
	"github.com/oreo-dtx-lab/oreo/pkg/serializer"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
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

func NewServer(port int, connMap map[string]txn.Connector, factory txn.DataItemFactory) *Server {
	return &Server{
		port:      port,
		reader:    *network.NewReader(connMap, factory, serializer.NewJSONSerializer()),
		committer: *network.NewCommitter(connMap, serializer.NewJSONSerializer(), factory),
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

func (s *Server) getItemType(dsName string) txn.ItemType {
	switch dsName {
	case "redis1", "Redis":
		return txn.RedisItem
	case "mongo1", "mongo2", "MongoDB":
		return txn.MongoItem
	case "CouchDB":
		return txn.CouchItem
	case "KVRocks":
		return txn.RedisItem
	default:
		return ""
	}
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

	item, dataType, err := s.reader.Read(req.DsName, req.Key, req.StartTime, req.Config, true)

	var response network.ReadResponse
	if err != nil {
		response = network.ReadResponse{
			Status: "Error",
			ErrMsg: err.Error(),
		}
	} else {
		// redisItem, ok := item.(*redis.RedisItem)
		// if !ok {
		// 	response = network.ReadResponse{
		// 		Status: "Error",
		// 		ErrMsg: "unexpected data type",
		// 	}
		// } else {
		// 	response = network.ReadResponse{
		// 		Status:   "OK",
		// 		DataType: dataType,
		// 		Data:     redisItem,
		// 	}
		// }

		response = network.ReadResponse{
			Status:       "OK",
			DataStrategy: dataType,
			Data:         item,
			ItemType:     s.getItemType(req.DsName),
		}
		// fmt.Printf("Read response: %v\n", response)
	}
	respBytes, _ := json.Marshal(response)
	ctx.Write(respBytes)
}

func (s *Server) prepareHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		Log.Infow("Prepare request", "latency", time.Since(startTime), "Topic", "CheckPoint")
	}()

	var req network.PrepareRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		errMsg := fmt.Sprintf("Invalid prepare request body, error: %s\n Body: %v\n", err.Error(), string(ctx.PostBody()))
		ctx.Error(errMsg, fasthttp.StatusBadRequest)
		return
	}

	verMap, err := s.committer.Prepare(req.DsName, req.ItemList,
		req.StartTime, req.CommitTime,
		req.Config, req.ValidationMap)
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
		ctx.Error("Invalid commit request body.", fasthttp.StatusBadRequest)
		return
	}

	err := s.committer.Commit(req.DsName, req.List)
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
		ctx.Error("Invalid abort request body.", fasthttp.StatusBadRequest)
		return
	}

	err := s.committer.Abort(req.DsName, req.KeyList, req.TxnId)
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

const (
	RedisPassword = "password"
	MongoUsername = "admin"
	MongoPassword = "admin"
	CouchUsername = "admin"
	CouchPassword = "password"
)

var port = 8000
var poolSize = 60
var traceFlag = false
var pprofFlag = false
var redisAddr1 = ""
var mongoAddr1 = ""
var mongoAddr2 = ""
var kvRocksAddr = ""
var couchAddr = ""
var workloadType = ""
var cg = false

var Log *zap.SugaredLogger

func main() {
	flag.IntVar(&port, "p", 8000, "Server Port")
	flag.IntVar(&poolSize, "s", 60, "Pool Size")
	flag.BoolVar(&traceFlag, "trace", false, "Enable trace")
	flag.BoolVar(&pprofFlag, "pprof", false, "Enable pprof")
	flag.StringVar(&workloadType, "w", "ycsb", "Workload Type")
	flag.StringVar(&redisAddr1, "redis1", "", "Redis Address")
	flag.StringVar(&mongoAddr1, "mongo1", "", "Mongo Address")
	flag.StringVar(&mongoAddr2, "mongo2", "", "Mongo Address")
	flag.StringVar(&kvRocksAddr, "kvrocks", "", "KVRocks Address")
	flag.StringVar(&couchAddr, "couch", "", "Couch Address")
	flag.BoolVar(&cg, "cg", false, "Enable Cherry Garcia Mode")
	flag.Parse()
	newLogger()

	if pprofFlag {
		// runtime.SetCPUProfileRate(1000)

		cpuFile, err := os.Create("executor_profile.prof")
		if err != nil {
			fmt.Println("无法创建 CPU profile 文件:", err)
			return
		}
		defer cpuFile.Close()

		// 开始 CPU profile
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			fmt.Println("无法启动 CPU profile:", err)
			return
		}
		defer pprof.StopCPUProfile()
	}

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
	if cg {
		fmt.Printf("Running under Cherry Garcia Mode")
		config.Debug.CherryGarciaMode = true
	}

	connMap := make(map[string]txn.Connector)
	switch workloadType {
	case "iot":
		if kvRocksAddr == "" || mongoAddr1 == "" {
			Log.Fatal("IOT Datastore address must be specified")
		}
		kvConn := getKVRocksConn()
		mongoConn1 := getMongoConn(1)
		connMap["KVRocks"] = kvConn
		connMap["MongoDB"] = mongoConn1
	case "social":
		if mongoAddr1 == "" || couchAddr == "" || redisAddr1 == "" {
			Log.Fatal("SOCIAL Datastore address must be specified")
		}
		mongoConn1 := getMongoConn(1)
		couchConn := getCouchConn()
		redisConn := getRedisConn(1)
		connMap["MongoDB"] = mongoConn1
		connMap["CouchDB"] = couchConn
		connMap["Redis"] = redisConn
	case "order":
		if mongoAddr1 == "" || couchAddr == "" || redisAddr1 == "" || kvRocksAddr == "" {
			Log.Fatal("ORDER Datastore address must be specified")
		}
		mongoConn1 := getMongoConn(1)
		couchConn := getCouchConn()
		redisConn := getRedisConn(1)
		kvConn := getKVRocksConn()
		connMap["MongoDB"] = mongoConn1
		connMap["CouchDB"] = couchConn
		connMap["Redis"] = redisConn
		connMap["KVRocks"] = kvConn

	case "ycsb":
		if redisAddr1 == "" && mongoAddr1 == "" && mongoAddr2 == "" {
			Log.Fatal("No datastore address specified")
		}

		if redisAddr1 != "" {
			redisConn := getRedisConn(1)
			connMap["redis1"] = redisConn
		}

		if mongoAddr1 != "" {
			mongoConn1 := getMongoConn(1)
			connMap["mongo1"] = mongoConn1
		}

		if mongoAddr2 != "" {
			mongoConn2 := getMongoConn(2)
			connMap["mongo2"] = mongoConn2
		}
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	server := NewServer(port, connMap, &redis.RedisItemFactory{})
	go server.Run()

	<-sigs

	Log.Info("Shutting down server")

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

func getKVRocksConn() *redis.RedisConnection {
	kvConn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address:  kvRocksAddr,
		Password: RedisPassword,
		PoolSize: poolSize,
	})
	err := kvConn.Connect()
	if err != nil {
		Log.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		_, _ = kvConn.Get("ping")
	}
	return kvConn
}

func getCouchConn() *couchdb.CouchDBConnection {
	couchConn := couchdb.NewCouchDBConnection(&couchdb.ConnectionOptions{
		Address:  couchAddr,
		Username: CouchUsername,
		Password: CouchPassword,
	})
	err := couchConn.Connect()
	if err != nil {
		Log.Fatal(err)
	}
	return couchConn
}

func getMongoConn(id int) *mongo.MongoConnection {
	address := ""
	switch id {
	case 1:
		address = mongoAddr1
	case 2:
		address = mongoAddr2
	default:
		Log.Fatal("Invalid mongo id")
	}
	mongoConn1 := mongo.NewMongoConnection(&mongo.ConnectionOptions{
		Address:        address,
		DBName:         "oreo",
		CollectionName: "benchmark",
		Username:       MongoUsername,
		Password:       MongoPassword,
	})
	err := mongoConn1.Connect()
	if err != nil {
		Log.Fatal(err)
	}
	return mongoConn1
}

func getRedisConn(id int) *redis.RedisConnection {

	address := ""
	switch id {
	case 1:
		address = redisAddr1
	default:
		Log.Fatal("Invalid redis id")
	}

	redisConn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address:  address,
		Password: RedisPassword,
		PoolSize: poolSize,
	})
	err := redisConn.Connect()
	if err != nil {
		Log.Fatal(err)
	}
	return redisConn
}
