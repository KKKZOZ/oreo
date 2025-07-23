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
	"strings"
	"syscall"
	"time"

	"benchmark/pkg/benconfig" //nolint

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	jsoniter "github.com/json-iterator/go"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/datastore/cassandra"
	"github.com/kkkzoz/oreo/pkg/datastore/couchdb"
	"github.com/kkkzoz/oreo/pkg/datastore/dynamodb"
	"github.com/kkkzoz/oreo/pkg/datastore/mongo"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/datastore/tikv"
	"github.com/kkkzoz/oreo/pkg/logger"
	"github.com/kkkzoz/oreo/pkg/network"
	"github.com/kkkzoz/oreo/pkg/serializer"
	"github.com/kkkzoz/oreo/pkg/timesource"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/valyala/fasthttp"
)

var json2 = jsoniter.ConfigCompatibleWithStandardLibrary

var Banner = `
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

func NewServer(
	port int,
	connMap map[string]txn.Connector,
	factory txn.DataItemFactory,
	timeSource timesource.TimeSourcer,
) *Server {
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
		case "/cache":
			s.cacheHandler(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	address := fmt.Sprintf(":%d", s.port)
	// fmt.Println(banner)
	logger.Infow("Server running", "address", address)
	log.Fatalf("Server failed: %v", fasthttp.ListenAndServe(address, router))
}

func (s *Server) pingHandler(ctx *fasthttp.RequestCtx) {
	_, _ = ctx.WriteString("pong")
}

func (s *Server) cacheHandler(ctx *fasthttp.RequestCtx) {
	method := string(ctx.Method())

	if method == "GET" {
		status := s.reader.GetCacheStatistic()
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBodyString(status)
		return
	}

	if method == "POST" {
		s.reader.ClearCache()
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBodyString("Cache cleared successfully")
		return
	}

	// 处理不支持的请求方法
	ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
	ctx.SetBodyString("Method not allowed")
}

func (s *Server) readHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		logger.Debugw("Read request", "latency", time.Since(startTime))
	}()

	var req network.ReadRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		errMsg := fmt.Sprintf("Invalid timestamp parameter: %s", err.Error())
		ctx.Error(errMsg, fasthttp.StatusBadRequest)
		return
	}

	logger.Infow(
		"Read request",
		"dsName",
		req.DsName,
		"key",
		req.Key,
		"startTime",
		req.StartTime,
		"config",
		req.Config,
	)

	item, dataType, gk, err := s.reader.Read(req.DsName, req.Key, req.StartTime, req.Config, true)

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
			GroupKey:     gk,
			ItemType:     network.GetItemType(req.DsName),
		}
		// fmt.Printf("Read response: %v\n", response)
	}
	respBytes, _ := json.Marshal(response)
	_, err = ctx.Write(respBytes)
	logger.CheckAndLogError("Failed to write response", err)
}

func (s *Server) prepareHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		logger.Debugw("Prepare request", "latency", time.Since(startTime), "Topic", "CheckPoint")
	}()

	var req network.PrepareRequest
	// body := ctx.PostBody()
	// logger.Infow("Prepare request", "body", string(body))
	if err := json2.Unmarshal(ctx.PostBody(), &req); err != nil {
		errMsg := fmt.Sprintf(
			"Invalid prepare request body, error: %s\n Body: %v\n",
			err.Error(),
			string(ctx.PostBody()),
		)
		ctx.Error(errMsg, fasthttp.StatusBadRequest)
		return
	}

	logger.Infow(
		"Prepare request",
		"dsName",
		req.DsName,
		"itemList",
		req.ItemList,
		"startTime",
		req.StartTime,
		"config",
		req.Config,
		"validationMap",
		req.ValidationMap,
	)

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
	_, err = ctx.Write(respBytes)
	logger.CheckAndLogError("Failed to write response", err)
}

func (s *Server) commitHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		logger.Debugw("Commit request", "latency", time.Since(startTime))
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
	_, err = ctx.Write(respBytes)
	logger.CheckAndLogError("Failed to write response", err)
}

func (s *Server) abortHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		logger.Debugw("Abort request", "latency", time.Since(startTime))
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
	_, err = ctx.Write(respBytes)
	logger.CheckAndLogError("Failed to write response", err)
}

// const (
// 	RedisPassword = "password"
// 	MongoUsername = "admin"
// 	MongoPassword = "password"
// 	CouchUsername = "admin"
// 	CouchPassword = "password"
// )

var (
	port           = 8000
	poolSize       = 60
	traceFlag      = false
	pprofFlag      = false
	workloadType   = ""
	db_combination = ""
	benConfigPath  = ""
	cg             = false
)

var benConfig = benconfig.BenchmarkConfig{}

func main() {
	parseFlag()
	err := loadConfig()
	if err != nil {
		logger.Fatal(err)
	}

	if pprofFlag {
		cpuFile, err := os.Create("executor_cpu_profile.prof")
		if err != nil {
			fmt.Println("Can not create CPU profile file:", err)
			return
		}
		defer func() {
			if err := cpuFile.Close(); err != nil {
				fmt.Println("Can not close CPU profile file:", err)
			}
		}()
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			fmt.Println("Can not start CPU profile:", err)
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

	config.Debug.DebugMode = false

	connMap := getConnMap()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	oracle := timesource.NewGlobalTimeSource(benConfig.TimeOracleUrl)
	server := NewServer(port, connMap, &redis.RedisItemFactory{}, oracle)
	go server.Run()

	<-sigs

	logger.Info("Shutting down server")
	fmt.Printf("Cache: %v\n", server.reader.GetCacheStatistic())
}

func loadConfig() error {
	bcLoader := aconfig.LoaderFor(&benConfig, aconfig.Config{
		SkipDefaults: true,
		SkipFiles:    false,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{benConfigPath},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
	})

	if err := bcLoader.Load(); err != nil {
		log.Fatalf("Error when loading benchmark configuration: %v\n", err)
	}

	if benConfig.TimeOracleUrl == "" {
		logger.Fatal("Time Oracle URL must be specified")
	}
	return nil
}

func parseFlag() {
	flag.IntVar(&port, "p", 8000, "Server Port")
	flag.IntVar(&poolSize, "s", 60, "Pool Size")
	flag.BoolVar(&traceFlag, "trace", false, "Enable trace")
	flag.BoolVar(&pprofFlag, "pprof", false, "Enable pprof")
	flag.StringVar(&workloadType, "w", "", "Workload Type")
	flag.StringVar(&db_combination, "db", "", "Database Combination")
	flag.BoolVar(&cg, "cg", false, "Enable Cherry Garcia Mode")
	flag.StringVar(&benConfigPath, "bc", "", "Benchmark Configuration Path")
	flag.Parse()

	if benConfigPath == "" {
		logger.Fatal("Benchmark Configuration Path must be specified")
	}

	if workloadType == "ycsb" && db_combination == "" {
		logger.Fatal("Database Combination must be specified for YCSB workload")
	}
}

func getConnMap() map[string]txn.Connector {
	connMap := make(map[string]txn.Connector)
	switch workloadType {
	case "iot":
		// if kvRocksAddr == "" || mongoAddr1 == "" {
		// 	Log.Fatal("IOT Datastore address must be specified")
		// }
		mongoConn1 := getMongoConn(1)
		redisConn := getRedisConn(1)
		connMap["MongoDB"] = mongoConn1
		connMap["Redis"] = redisConn
	case "social":
		// if mongoAddr1 == "" || couchAddr == "" || redisAddr1 == "" {
		// 	Log.Fatal("SOCIAL Datastore address must be specified")
		// }
		mongoConn1 := getMongoConn(1)
		redisConn := getRedisConn(1)
		cassandraConn := getCassandraConn()
		connMap["MongoDB"] = mongoConn1
		connMap["Redis"] = redisConn
		connMap["Cassandra"] = cassandraConn
	case "order":
		// if mongoAddr1 == "" || couchAddr == "" || redisAddr1 == "" || kvRocksAddr == "" {
		// 	Log.Fatal("ORDER Datastore address must be specified")
		// }
		mongoConn1 := getMongoConn(1)
		kvrocksConn := getKVRocksConn()
		redisConn := getRedisConn(1)
		cassandraConn := getCassandraConn()
		connMap["MongoDB"] = mongoConn1
		connMap["KVRocks"] = kvrocksConn
		connMap["Redis"] = redisConn
		connMap["Cassandra"] = cassandraConn

	case "ycsb":
		// if redisAddr1 == "" && mongoAddr1 == "" && mongoAddr2 == "" {
		// 	Log.Fatal("No datastore address specified")
		// }
		dbList := strings.Split(db_combination, ",")
		for _, db := range dbList {
			switch db {
			case "Redis":
				redisConn := getRedisConn(1)
				connMap["Redis"] = redisConn
			case "MongoDB1":
				mongoConn1 := getMongoConn(1)
				connMap["MongoDB1"] = mongoConn1
			case "MongoDB2":
				mongoConn2 := getMongoConn(2)
				connMap["MongoDB2"] = mongoConn2
			case "KVRocks":
				kvConn := getKVRocksConn()
				connMap["KVRocks"] = kvConn
			case "CouchDB":
				couchConn := getCouchConn()
				connMap["CouchDB"] = couchConn
			case "Cassandra":
				cassConn := getCassandraConn()
				connMap["Cassandra"] = cassConn
			case "DynamoDB":
				dynamoConn := getDynamoConn()
				connMap["DynamoDB"] = dynamoConn
			case "TiKV":
				tikvConn := getTiKVConn()
				connMap["TiKV"] = tikvConn
			default:
				logger.Fatal("Invalid database combination")
			}
		}
	}
	return connMap
}

func getKVRocksConn() *redis.RedisConnection {
	kvConn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address:  benConfig.KVRocksAddr,
		Password: benConfig.RedisPassword,
		PoolSize: poolSize,
	})
	err := kvConn.Connect()
	if err != nil {
		logger.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		_, _ = kvConn.Get("ping")
	}
	return kvConn
}

func getCouchConn() *couchdb.CouchDBConnection {
	couchConn := couchdb.NewCouchDBConnection(&couchdb.ConnectionOptions{
		Address: benConfig.CouchDBAddr,
		// Username: CouchUsername,
		// Password: CouchPassword,
		DBName: "oreo",
	})
	err := couchConn.Connect()
	if err != nil {
		logger.Fatal(err)
	}
	return couchConn
}

func getMongoConn(id int) *mongo.MongoConnection {
	address := ""
	switch id {
	case 1:
		address = benConfig.MongoDBAddr1
	case 2:
		address = benConfig.MongoDBAddr2
	default:
		logger.Fatal("Invalid mongo id")
	}
	mongoConn := mongo.NewMongoConnection(&mongo.ConnectionOptions{
		Address:        address,
		DBName:         "oreo",
		CollectionName: "benchmark",
		Username:       benConfig.MongoDBUsername,
		Password:       benConfig.MongoDBPassword,
	})
	err := mongoConn.Connect()
	if err != nil {
		logger.Fatal(err)
	}
	return mongoConn
}

func getRedisConn(id int) *redis.RedisConnection {
	address := ""
	switch id {
	case 1:
		address = benConfig.RedisAddr
	default:
		logger.Fatal("Invalid redis id")
	}

	redisConn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address:  address,
		Password: benConfig.RedisPassword,
		PoolSize: poolSize,
	})
	err := redisConn.Connect()
	if err != nil {
		logger.Fatal(err)
	}
	return redisConn
}

func getCassandraConn() *cassandra.CassandraConnection {
	cassConn := cassandra.NewCassandraConnection(&cassandra.ConnectionOptions{
		Hosts:    benConfig.CassandraAddr,
		Keyspace: "oreo",
	})
	err := cassConn.Connect()
	if err != nil {
		logger.Fatal(err)
	}
	return cassConn
}

func getDynamoConn() *dynamodb.DynamoDBConnection {
	dynamoConn := dynamodb.NewDynamoDBConnection(&dynamodb.ConnectionOptions{
		TableName: "oreo",
		Endpoint:  benConfig.DynamoDBAddr,
	})
	err := dynamoConn.Connect()
	if err != nil {
		logger.Fatal(err)
	}
	return dynamoConn
}

func getTiKVConn() *tikv.TiKVConnection {
	tikvConn := tikv.NewTiKVConnection(&tikv.ConnectionOptions{
		PDAddrs: benConfig.TiKVAddr,
	})
	err := tikvConn.Connect()
	if err != nil {
		logger.Fatal(err)
	}
	return tikvConn
}
