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

	jsoniter "github.com/json-iterator/go"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/cassandra"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/couchdb"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/dynamodb"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/mongo"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/redis"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/tikv"
	"github.com/oreo-dtx-lab/oreo/pkg/network"
	"github.com/oreo-dtx-lab/oreo/pkg/serializer"
	"github.com/oreo-dtx-lab/oreo/pkg/timesource"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var json2 = jsoniter.ConfigCompatibleWithStandardLibrary

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

func NewServer(port int, connMap map[string]txn.Connector, factory txn.DataItemFactory, timeSource timesource.TimeSourcer) *Server {
	reader := *network.NewReader(connMap, factory, serializer.NewJSON2Serializer(), network.NewCacher())
	return &Server{
		port:      port,
		reader:    reader,
		committer: *network.NewCommitter(connMap, reader, serializer.NewJSON2Serializer(), factory, timeSource),
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
	// fmt.Println(banner)
	Log.Infow("Server running", "address", address)
	log.Fatalf("Server failed: %v", fasthttp.ListenAndServe(address, router))
}

// func (s *Server) getItemType(dsName string) txn.ItemType {
// 	switch dsName {
// 	case "redis1", "Redis":
// 		return txn.RedisItem
// 	case "mongo1", "mongo2", "MongoDB":
// 		return txn.MongoItem
// 	case "CouchDB":
// 		return txn.CouchItem
// 	case "KVRocks":
// 		return txn.RedisItem
// 	case "Cassandra":
// 		return txn.CassandraItem
// 	default:
// 		return ""
// 	}
// }

func (s *Server) pingHandler(ctx *fasthttp.RequestCtx) {
	ctx.WriteString("pong")
}

func (s *Server) readHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		Log.Debugw("Read request", "latency", time.Since(startTime))
	}()

	var req network.ReadRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		errMsg := fmt.Sprintf("Invalid timestamp parameter: %s", err.Error())
		ctx.Error(errMsg, fasthttp.StatusBadRequest)
		return
	}

	Log.Infow("Read request", "dsName", req.DsName, "key", req.Key, "startTime", req.StartTime, "config", req.Config)

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
			ItemType:     network.GetItemType(req.DsName),
		}
		// fmt.Printf("Read response: %v\n", response)
	}
	respBytes, _ := json.Marshal(response)
	ctx.Write(respBytes)
}

func (s *Server) prepareHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		Log.Debugw("Prepare request", "latency", time.Since(startTime), "Topic", "CheckPoint")
	}()

	var req network.PrepareRequest
	// body := ctx.PostBody()
	// Log.Infow("Prepare request", "body", string(body))
	if err := json2.Unmarshal(ctx.PostBody(), &req); err != nil {
		errMsg := fmt.Sprintf("Invalid prepare request body, error: %s\n Body: %v\n", err.Error(), string(ctx.PostBody()))
		ctx.Error(errMsg, fasthttp.StatusBadRequest)
		return
	}

	Log.Infow("Prepare request", "dsName", req.DsName, "itemList", req.ItemList, "startTime", req.StartTime, "config", req.Config, "validationMap", req.ValidationMap)

	verMap, tCommit, err := s.committer.Prepare(req.DsName, req.ItemList,
		req.StartTime, req.Config, req.ValidationMap)
	var resp network.PrepareResponse
	if err != nil {
		resp = network.PrepareResponse{
			Status: "Error",
			ErrMsg: err.Error(),
		}
	} else {
		resp = network.PrepareResponse{
			Status:  "OK",
			VerMap:  verMap,
			TCommit: tCommit,
		}
	}
	respBytes, _ := json2.Marshal(resp)
	ctx.Write(respBytes)
}

func (s *Server) commitHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		Log.Debugw("Commit request", "latency", time.Since(startTime))
	}()

	var req network.CommitRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.Error("Invalid commit request body.", fasthttp.StatusBadRequest)
		return
	}

	err := s.committer.Commit(req.DsName, req.List, req.TCommit)
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
		Log.Debugw("Abort request", "latency", time.Since(startTime))
	}()

	var req network.AbortRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.Error("Invalid abort request body.", fasthttp.StatusBadRequest)
		return
	}

	err := s.committer.Abort(req.DsName, req.KeyList, req.GroupKeyList)
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
	MongoPassword = "password"
	CouchUsername = "admin"
	CouchPassword = "password"
)

var port = 8000
var poolSize = 60
var traceFlag = false
var pprofFlag = false
var timeOracleUrl = ""
var redisAddr1 = ""
var mongoAddr1 = ""
var mongoAddr2 = ""
var kvRocksAddr = ""
var couchAddr = ""
var cassandraAddr = ""
var dynamodbAddr = ""
var tikvAddr = ""
var workloadType = ""
var cg = false

var Log *zap.SugaredLogger

func main() {
	parseFlag()

	if pprofFlag {
		cpuFile, err := os.Create("executor_cpu_profile.prof")
		if err != nil {
			fmt.Println("无法创建 CPU profile 文件:", err)
			return
		}
		defer cpuFile.Close()
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			fmt.Println("无法启动 CPU profile:", err)
			return
		}
		defer pprof.StopCPUProfile()

		// fMem, err := os.Create("executor_mem_profile.prof")
		// if err != nil {
		// 	panic(err)
		// }
		// defer func() {
		// 	runtime.GC() // 触发 GC，确保内存分配的准确性
		// 	pprof.WriteHeapProfile(fMem)
		// 	fMem.Close()
		// }()
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

	connMap := getConnMap()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	oracle := timesource.NewGlobalTimeSource(timeOracleUrl)
	server := NewServer(port, connMap, &redis.RedisItemFactory{}, oracle)
	go server.Run()

	<-sigs

	Log.Info("Shutting down server")
	fmt.Printf("Cache: %v\n", server.reader.GetCacheStatistic())

}

func parseFlag() {
	flag.IntVar(&port, "p", 8000, "Server Port")
	flag.IntVar(&poolSize, "s", 60, "Pool Size")
	flag.BoolVar(&traceFlag, "trace", false, "Enable trace")
	flag.BoolVar(&pprofFlag, "pprof", false, "Enable pprof")
	flag.StringVar(&workloadType, "w", "ycsb", "Workload Type")
	flag.StringVar(&timeOracleUrl, "timeurl", "", "Time Oracle URL")
	flag.StringVar(&redisAddr1, "redis1", "", "Redis Address")
	flag.StringVar(&mongoAddr1, "mongo1", "", "Mongo Address")
	flag.StringVar(&mongoAddr2, "mongo2", "", "Mongo Address")
	flag.StringVar(&kvRocksAddr, "kvrocks", "", "KVRocks Address")
	flag.StringVar(&couchAddr, "couch", "", "Couch Address")
	flag.StringVar(&cassandraAddr, "cas", "", "Cassandra Address")
	flag.StringVar(&dynamodbAddr, "dynamodb", "", "DynamoDB Address")
	flag.StringVar(&tikvAddr, "tikv", "", "TiKV Address")
	flag.BoolVar(&cg, "cg", false, "Enable Cherry Garcia Mode")
	flag.Parse()

	if timeOracleUrl == "" {
		Log.Fatal("Time Oracle URL must be specified")
	}

	newLogger()
}

func getConnMap() map[string]txn.Connector {
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
		// if redisAddr1 == "" && mongoAddr1 == "" && mongoAddr2 == "" {
		// 	Log.Fatal("No datastore address specified")
		// }

		if redisAddr1 != "" {
			redisConn := getRedisConn(1)
			connMap["Redis"] = redisConn
		}

		if mongoAddr1 != "" {
			mongoConn1 := getMongoConn(1)
			connMap["MongoDB"] = mongoConn1
		}

		if mongoAddr2 != "" {
			mongoConn2 := getMongoConn(2)
			connMap["MongoDB2"] = mongoConn2
		}

		if kvRocksAddr != "" {
			kvConn := getKVRocksConn()
			connMap["KVRocks"] = kvConn
		}

		if couchAddr != "" {
			couchConn := getCouchConn()
			connMap["CouchDB"] = couchConn
		}
		if cassandraAddr != "" {
			cassConn := getCassandraConn()
			connMap["Cassandra"] = cassConn
		}
		if dynamodbAddr != "" {
			dynamoConn := getDynamoConn()
			connMap["DynamoDB"] = dynamoConn
		}
		if tikvAddr != "" {
			tikvConn := getTiKVConn()
			connMap["TiKV"] = tikvConn
		}
	}
	return connMap
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
		Address: couchAddr,
		// Username: CouchUsername,
		// Password: CouchPassword,
		DBName: "oreo",
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

func getCassandraConn() *cassandra.CassandraConnection {
	cassConn := cassandra.NewCassandraConnection(&cassandra.ConnectionOptions{
		Hosts:    []string{cassandraAddr},
		Keyspace: "oreo",
	})
	err := cassConn.Connect()
	if err != nil {
		Log.Fatal(err)
	}
	return cassConn
}

func getDynamoConn() *dynamodb.DynamoDBConnection {
	dynamoConn := dynamodb.NewDynamoDBConnection(&dynamodb.ConnectionOptions{
		TableName: "oreo",
		Endpoint:  dynamodbAddr,
	})
	err := dynamoConn.Connect()
	if err != nil {
		Log.Fatal(err)
	}
	return dynamoConn
}

func getTiKVConn() *tikv.TiKVConnection {
	tikvConn := tikv.NewTiKVConnection(&tikv.ConnectionOptions{
		PDAddrs: []string{tikvAddr},
	})
	err := tikvConn.Connect()
	if err != nil {
		Log.Fatal(err)
	}
	return tikvConn
}
