package main

import (
	"benchmark/pkg/benconfig"
	"benchmark/pkg/client"
	"benchmark/pkg/measurement"
	"benchmark/pkg/workload"
	"benchmark/ycsb"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"runtime/trace"
	"time"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	cfg "github.com/oreo-dtx-lab/oreo/pkg/config"
)

var (
	benConfig = benconfig.BenchmarkConfig{}
)

var help = flag.Bool("help", false, "Show help")
var dbType = ""
var mode = "load"
var workloadType = ""
var workloadConfigPath = ""
var benConfigPath = ""
var threadNum = 1
var traceFlag = false
var pprofFlag = false
var isRemote = false
var preset = ""
var readStrategy = ""
var ablationLevel = 4

func main() {
	parseAndValidateFlag()

	if pprofFlag {
		cpuFile, err := os.Create("ben_profile.prof")
		if err != nil {
			log.Fatalln("Can not create CPU profile:", err)
		}
		defer cpuFile.Close()
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			log.Fatalln("Can not start CPU profile:", err)
		}
		defer pprof.StopCPUProfile()
	}

	if traceFlag {
		f, err := os.Create("ben_trace.out")
		if err != nil {
			panic(err)
		}
		err = trace.Start(f)
		if err != nil {
			panic(err)
		}
		defer trace.Stop()
	}

	wp := loadConfig()

	switch preset {
	case "cg":
		fmt.Printf("Running under Cherry Garcia Mode\n")
		cfg.Config.ReadStrategy = cfg.Pessimistic
		cfg.Debug.CherryGarciaMode = true
		cfg.Debug.DebugMode = true
		cfg.Debug.ConnAdditionalLatency = benConfig.Latency
		cfg.Config.ConcurrentOptimizationLevel = 0
		cfg.Config.AsyncLevel = 2
	case "native":
		fmt.Printf("Running under Native Mode\n")
		cfg.Debug.NativeMode = true
		cfg.Debug.DebugMode = true
		cfg.Debug.ConnAdditionalLatency = benConfig.Latency
	case "oreo":
		fmt.Printf("Running under Oreo Mode\n")
		isRemote = true
		cfg.Config.ReadStrategy = cfg.Pessimistic
		cfg.Debug.DebugMode = true
		cfg.Debug.HTTPAdditionalLatency = benConfig.Latency
		cfg.Debug.ConnAdditionalLatency = 0
		cfg.Config.ConcurrentOptimizationLevel = 2
		cfg.Config.AsyncLevel = 2
	}
	cfg.Config.LeaseTime = 1000 * time.Millisecond
	cfg.Config.MaxRecordLength = 4

	switch readStrategy {
	case "p":
		cfg.Config.ReadStrategy = cfg.Pessimistic
	case "ac":
		cfg.Config.ReadStrategy = cfg.AssumeCommit
	case "aa":
		cfg.Config.ReadStrategy = cfg.AssumeAbort
	}

	cfg.Config.AblationLevel = ablationLevel

	measurement.InitMeasure()
	measurement.EnableWarmUp(true)

	wp.ThreadCount = threadNum
	// setupDistribution(wp, dbType)
	wl := createWorkload(wp)
	client := generateClient(&wl, wp, dbType)

	// if isRemote {
	// 	warmUpHttpClient()
	// }
	// warmupTimeSourceClient()
	displayBenchmarkInfo()

	switch mode {
	case "load":
		// cfg.Config.ConcurrentOptimizationLevel = 1
		if benconfig.MaxLoadBatchSize == 0 {
			log.Fatalf("MaxLoadBatchSize should be specified")
		}
		cfg.Debug.DebugMode = false
		cfg.Debug.HTTPAdditionalLatency = 0
		cfg.Debug.ConnAdditionalLatency = 0
		wp.DoBenchmark = false
		if workloadType == "multi-ycsb" {
			fmt.Printf("No support load mode for multi-ycsb\n")
			return
		}
		fmt.Println("Start to load data")
		client.RunLoad()
		fmt.Println("Load finished")
	case "run":
		wp.DoBenchmark = true
		fmt.Println("Start to run benchmark")
		measurement.EnableWarmUp(false)
		client.RunBenchmark()
	default:
		panic("Invalid mode")
	}

	if mode == "load" {
		time.Sleep(2 * time.Second)
	}

}

// func warmUpHttpClient() {
// 	for _, addr := range config.RemoteAddressList {
// 		url := fmt.Sprintf("http://%s/ping", addr)
// 		num := min(800, threadNum+200)
// 		var wg sync.WaitGroup
// 		wg.Add(num)

// 		for i := 0; i < num; i++ {
// 			go func() {
// 				defer wg.Done()
// 				resp, err := network.HttpClient.Get(url)
// 				if err != nil {
// 					fmt.Printf("Error when warming up http client: %v\n", err)
// 				}
// 				defer func() {
// 					_, _ = io.CopyN(io.Discard, resp.Body, 1024*4)
// 					_ = resp.Body.Close()
// 				}()
// 			}()
// 		}
// 		wg.Wait()
// 	}
// }

// func warmupTimeSourceClient() {
// 	timeUrl := fmt.Sprintf("%s/timestamp/common", config.TimeOracleUrl)
// 	num := 300
// 	var wg sync.WaitGroup
// 	wg.Add(num)

// 	for i := 0; i < num; i++ {
// 		go func() {
// 			defer wg.Done()
// 			resp, err := timesource.HttpClient.Get(timeUrl)
// 			if err != nil {
// 				fmt.Printf("Error when warming up http client: %v\n", err)
// 			}
// 			defer func() {
// 				_, _ = io.CopyN(io.Discard, resp.Body, 1024*4)
// 				_ = resp.Body.Close()
// 			}()
// 		}()
// 	}
// 	wg.Wait()
// }

func createWorkload(wp *workload.WorkloadParameter) workload.Workload {
	if dbType == "oreo-ycsb" {
		fmt.Println("This is a Oreo YCSB benchmark")
		return workload.NewOreoYCSBWorkload(wp)
	}

	if workloadType != "" {
		switch workloadType {
		case "ycsb":
			fmt.Println("This is a YCSB benchmark")
			wp.WorkloadName = "ycsb"
			return workload.NewYCSBWorkload(wp)
		case "multi-ycsb":
			fmt.Println("This is a multi-ycsb benchmark")
			wp.WorkloadName = "multi-ycsb"
			return workload.NewMultiYCSBWorkload(wp)
		case "dc":
			fmt.Println("This is a data consistency test")
			return workload.NewDataConsistencyWorkload(wp)
		case "tp":
			fmt.Println("This is a transaction performance test")
			return workload.NewTxnPerformanceWorkload(wp)
		case "ad":
			fmt.Println("This is a across datastore test")
			return workload.NewAcrossDatastoreWorkload(wp)
		case "iot":
			fmt.Println("This is a IoT workload")
			return workload.NewIotWorkload(wp)
		case "social":
			fmt.Println("This is a social network workload")
			return workload.NewSocialWorkload(wp)
		case "order":
			fmt.Println("This is a order workload")
			return workload.NewOrderWorkload(wp)
		default:
			panic("Invalid workload type")
		}
	} else {
		panic("WorkloadType should be specified")
	}
}

func generateClient(wl *workload.Workload, wp *workload.WorkloadParameter, dbName string) *client.Client {
	if dbType == "" {
		panic("DBType should be specified")
	}

	var c *client.Client
	switch dbName {
	case "oreo-ycsb":
		wp.DBName = "oreo-ycsb"
		creator, err := OreoYCSBCreator(workloadType, preset)
		if err != nil {
			fmt.Printf("Error when creating oreo-ycsb client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)

	case "oreo":
		wp.DBName = "oreo"
		creator, err := OreoRealisticCreator(workloadType, isRemote, preset)
		if err != nil {
			fmt.Printf("Error when creating oreo client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)

	case "native":
		wp.DBName = "native"
		creator, err := NativeRealisticCreator(workloadType)
		if err != nil {
			fmt.Printf("Error when creating native client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)

	case "redis":
		wp.DBName = "redis"
		creator, err := RedisCreator(benConfig.RedisAddr)
		if err != nil {
			fmt.Printf("Error when creating redis client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	case "mongo":
		wp.DBName = "mongo"
		creator, err := MongoCreator(benConfig.MongoDBAddr1)
		if err != nil {
			fmt.Printf("Error when creating mongo client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	case "native-mm":
		wp.DBName = "native-mm"
		dbSetCreator, err := NativeCreator("mm")
		if err != nil {
			fmt.Printf("Error when creating native-mm client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			"native-mm": dbSetCreator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	case "native-rm":
		wp.DBName = "native-rm"
		dbSetCreator, err := NativeCreator("rm")
		if err != nil {
			fmt.Printf("Error when creating native-rm client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			"native-rm": dbSetCreator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	case "oreo-redis":
		wp.DBName = "oreo-redis"
		creator, err := OreoRedisCreator(isRemote)
		if err != nil {
			fmt.Printf("Error when creating oreo-redis client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	case "oreo-mongo":
		wp.DBName = "oreo-mongo"
		creator, err := OreoMongoCreator(isRemote)
		if err != nil {
			fmt.Printf("Error when creating oreo-mongo client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	case "oreo-couch":
		wp.DBName = "oreo-couch"
		creator, err := OreoCouchCreator(isRemote)
		if err != nil {
			fmt.Printf("Error when creating oreo-couch client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	case "oreo-mm":
		wp.DBName = "oreo-mm"
		creator, err := OreoCreator("mm", isRemote)
		if err != nil {
			fmt.Printf("Error when creating oreo-mm client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	case "oreo-rm":
		wp.DBName = "oreo-rm"
		creator, err := OreoCreator("rm", isRemote)
		if err != nil {
			fmt.Printf("Error when creating oreo-rm client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	default:
		panic("Unsupport db type")
	}
	return c
}

func parseAndValidateFlag() {

	flag.StringVar(&dbType, "d", "", "DB type")
	flag.StringVar(&mode, "m", "load", "Mode: load or run")
	flag.StringVar(&workloadType, "wl", "", "Workload type")
	flag.StringVar(&workloadConfigPath, "wc", "", "Workload configuration path")
	flag.StringVar(&benConfigPath, "bc", "", "Benchmark configuration path")
	flag.IntVar(&threadNum, "t", 1, "Thread number")
	flag.BoolVar(&traceFlag, "trace", false, "Enable trace")
	flag.BoolVar(&pprofFlag, "pprof", false, "Enable pprof")
	flag.BoolVar(&isRemote, "remote", false, "Run in remote mode (for Oreo series)")
	flag.StringVar(&preset, "ps", "", "Preset configuration for evaluation")
	flag.StringVar(&readStrategy, "read", "p", "Read Strategy")
	flag.IntVar(&ablationLevel, "ab", 4, "Ablation level")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if workloadConfigPath == "" {
		panic("Workload configuration path should be specified")
	}

	if threadNum <= 0 {
		panic("ThreadNum should be a positive integer")
	}
}

func displayBenchmarkInfo() {
	fmt.Printf("-----------------\n")
	fmt.Printf("DBType: %s\n", dbType)
	fmt.Printf("Mode: %s\n", mode)
	fmt.Printf("WorkloadType: %s\n", workloadType)
	fmt.Printf("ThreadNum: %d\n", threadNum)
	fmt.Printf("Remote Mode: %v\n", isRemote)
	fmt.Printf("Read Strategy: %v\n", readStrategy)
	fmt.Printf("ConcurrentOptimizationLevel: %d\nAsyncLevel: %d\nMaxOutstandingRequest: %d\nMaxRecordLength: %d\n",
		cfg.Config.ConcurrentOptimizationLevel, cfg.Config.AsyncLevel,
		cfg.Config.MaxOutstandingRequest, cfg.Config.MaxRecordLength)
	fmt.Printf("HTTPAdditionalLatency: %v ConnAdditionalLatency: %v\n",
		cfg.Debug.HTTPAdditionalLatency, cfg.Debug.ConnAdditionalLatency)
	fmt.Printf("LeaseTime: %v\n", cfg.Config.LeaseTime)
	fmt.Printf("ZipfianConstant: %v\n", benConfig.ZipfianConstant)
	fmt.Printf("-----------------\n")
}

func loadConfig() *workload.WorkloadParameter {

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
		return nil
	}

	benConfig.Latency = time.Duration(benConfig.LatencyValue) * time.Millisecond
	benconfig.ExecutorAddressMap = benConfig.ExecutorAddressMap
	benconfig.TimeOracleUrl = benConfig.TimeOracleUrl
	benconfig.ZipfianConstant = benConfig.ZipfianConstant
	benconfig.MaxLoadBatchSize = benConfig.MaxLoadBatchSize

	wp := &workload.WorkloadParameter{}
	wpLoader := aconfig.LoaderFor(wp, aconfig.Config{
		SkipDefaults: true,
		SkipFiles:    false,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{workloadConfigPath},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
	})
	if err := wpLoader.Load(); err != nil {
		log.Fatalf("Error when loading workload configuration: %v\n", err)
		return nil
	}
	benconfig.MaxLoadBatchSize = wp.MaxLoadBatchSize

	return wp
}
