package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/kkkzoz/oreo/pkg/datastore/mongo"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/discovery"
	"github.com/kkkzoz/oreo/pkg/network"
	"github.com/kkkzoz/oreo/pkg/timesource"
	"github.com/kkkzoz/oreo/pkg/txn"
	redisd "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	mongod "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// YCSB-style record with 10 fields, each ~100 bytes
type TestRecord struct {
	ID     string `bson:"_id"    json:"id"`
	Field0 string `bson:"field0" json:"field0"`
	Field1 string `bson:"field1" json:"field1"`
	Field2 string `bson:"field2" json:"field2"`
	Field3 string `bson:"field3" json:"field3"`
	Field4 string `bson:"field4" json:"field4"`
	Field5 string `bson:"field5" json:"field5"`
	Field6 string `bson:"field6" json:"field6"`
	Field7 string `bson:"field7" json:"field7"`
	Field8 string `bson:"field8" json:"field8"`
	Field9 string `bson:"field9" json:"field9"`
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// Generate a YCSB-style test record with 10 fields of 100 bytes each
func generateTestRecord(id string) TestRecord {
	return TestRecord{
		ID:     id,
		Field0: generateRandomString(100),
		Field1: generateRandomString(100),
		Field2: generateRandomString(100),
		Field3: generateRandomString(100),
		Field4: generateRandomString(100),
		Field5: generateRandomString(100),
		Field6: generateRandomString(100),
		Field7: generateRandomString(100),
		Field8: generateRandomString(100),
		Field9: generateRandomString(100),
	}
}

func insertViaOreo(
	client *network.Client,
	oracle timesource.TimeSourcer,
	numRecords int,
	batchSize int,
) error {
	redis_conn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address:  "localhost:6379",
		Password: "kkkzoz",
	})
	mongo1_conn := mongo.NewMongoConnection(&mongo.ConnectionOptions{
		Address:        "mongodb://localhost:27017",
		Username:       "admin",
		Password:       "password",
		DBName:         "oreo",
		CollectionName: "records",
	})

	fmt.Printf(
		"Starting Oreo insertion of %d records (YCSB-style, ~1KB each) with batch size %d...\n",
		numRecords,
		batchSize,
	)
	startTime := time.Now()

	totalTxns := (numRecords + batchSize - 1) / batchSize

	for txnIdx := 0; txnIdx < totalTxns; txnIdx++ {
		write_txn := txn.NewTransactionWithRemote(client, oracle)
		redis_datastore := redis.NewRedisDatastore("Redis", redis_conn)
		mongo1_datastore := mongo.NewMongoDatastore("MongoDB1", mongo1_conn)
		write_txn.AddDatastores(redis_datastore, mongo1_datastore)

		err := write_txn.Start()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %v", err)
		}

		startIdx := txnIdx * batchSize
		endIdx := startIdx + batchSize
		if endIdx > numRecords {
			endIdx = numRecords
		}

		for i := startIdx; i < endIdx; i++ {
			record := generateTestRecord(fmt.Sprintf("oreo_record_%d", i))
			// fmt.Printf("Generated record: %v\n", record.ID)

			err = write_txn.Write("Redis", record.ID, record)
			if err != nil {
				return fmt.Errorf("failed to write to Redis: %v", err)
			}

			err = write_txn.Write("MongoDB1", record.ID, record)
			if err != nil {
				return fmt.Errorf("failed to write to MongoDB: %v", err)
			}
		}

		err = write_txn.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %v", err)
		}

		recordsInserted := endIdx
		if (txnIdx+1)%100 == 0 || txnIdx == totalTxns-1 {
			fmt.Printf(
				"Completed %d/%d transactions, inserted %d records\n",
				txnIdx+1,
				totalTxns,
				recordsInserted,
			)
		}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("Oreo insertion completed in %v\n", elapsed)
	fmt.Printf("Average throughput: %.2f records/sec\n", float64(numRecords)/elapsed.Seconds())
	return nil
}

func insertViaNative(numRecords int) error {
	// Native MongoDB connection
	ctx := context.Background()
	mongoClient, err := mongod.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017").
		SetAuth(options.Credential{
			Username: "admin",
			Password: "password",
		}))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database("oreo_native").Collection("records")

	// Native Redis connection
	redisClient := redisd.NewClient(&redisd.Options{
		Addr:     "localhost:6379",
		Password: "kkkzoz",
		DB:       1, // Use different DB to separate from Oreo
	})
	defer redisClient.Close()

	fmt.Printf("Starting native insertion of %d records (YCSB-style, ~1KB each)...\n", numRecords)
	startTime := time.Now()

	for i := 0; i < numRecords; i++ {
		record := generateTestRecord(fmt.Sprintf("native_record_%d", i))

		// Insert to MongoDB
		_, err := collection.InsertOne(ctx, record)
		if err != nil {
			return fmt.Errorf("failed to insert to MongoDB: %v", err)
		}

		// Insert to Redis
		recordBytes, err := bson.Marshal(record)
		if err != nil {
			return fmt.Errorf("failed to marshal record: %v", err)
		}
		err = redisClient.Set(ctx, record.ID, recordBytes, 0).Err()
		if err != nil {
			return fmt.Errorf("failed to set Redis key: %v", err)
		}

		if (i+1)%1000 == 0 {
			fmt.Printf("Inserted %d records via native\n", i+1)
		}
	}

	elapsed := time.Since(startTime)
	fmt.Printf("Native insertion completed in %v\n", elapsed)
	fmt.Printf("Average throughput: %.2f records/sec\n", float64(numRecords)/elapsed.Seconds())
	return nil
}

func getMongoStats(dbName string) {
	ctx := context.Background()
	mongoClient, err := mongod.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017").
		SetAuth(options.Credential{
			Username: "admin",
			Password: "password",
		}))
	if err != nil {
		fmt.Printf("Failed to connect to MongoDB: %v\n", err)
		return
	}
	defer mongoClient.Disconnect(ctx)

	var result bson.M
	err = mongoClient.Database(dbName).
		RunCommand(ctx, bson.D{{Key: "dbStats", Value: 1}}).
		Decode(&result)
	if err != nil {
		fmt.Printf("Failed to get DB stats: %v\n", err)
		return
	}

	fmt.Printf("\nMongoDB (%s) Storage Stats:\n", dbName)
	fmt.Printf("  Data Size: %.2f MB\n", result["dataSize"].(float64)/1024/1024)
	fmt.Printf("  Storage Size: %.2f MB\n", result["storageSize"].(float64)/1024/1024)
	fmt.Printf("  Index Size: %.2f MB\n", result["indexSize"].(float64)/1024/1024)
	fmt.Printf(
		"  Total Size: %.2f MB\n",
		(result["dataSize"].(float64)+result["indexSize"].(float64))/1024/1024,
	)
}

func getRedisStats(dbNum int, prefix string) {
	ctx := context.Background()
	redisClient := redisd.NewClient(&redisd.Options{
		Addr:     "localhost:6379",
		Password: "kkkzoz",
		DB:       dbNum,
	})
	defer redisClient.Close()

	// Get memory info
	memInfo, err := redisClient.Info(ctx, "memory").Result()
	if err != nil {
		fmt.Printf("Failed to get Redis memory info: %v\n", err)
		return
	}

	// Get DB size (number of keys)
	dbSize, err := redisClient.DBSize(ctx).Result()
	if err != nil {
		fmt.Printf("Failed to get Redis DB size: %v\n", err)
		return
	}

	// Parse memory info
	var usedMemory, usedMemoryDataset float64
	fmt.Sscanf(memInfo, "# Memory\nused_memory:%f", &usedMemory)

	// Try to get more detailed memory info
	lines := parseRedisInfo(memInfo)
	if val, ok := lines["used_memory"]; ok {
		fmt.Sscanf(val, "%f", &usedMemory)
	}
	if val, ok := lines["used_memory_dataset"]; ok {
		fmt.Sscanf(val, "%f", &usedMemoryDataset)
	}

	fmt.Printf("\nRedis (DB %d - %s) Storage Stats:\n", dbNum, prefix)
	fmt.Printf("  Number of Keys: %d\n", dbSize)
	fmt.Printf("  Used Memory: %.2f MB\n", usedMemory/1024/1024)
	if usedMemoryDataset > 0 {
		fmt.Printf("  Dataset Memory: %.2f MB\n", usedMemoryDataset/1024/1024)
	}
}

func parseRedisInfo(info string) map[string]string {
	result := make(map[string]string)
	lines := []rune(info)
	var key, value string
	parsingKey := true

	for i := 0; i < len(lines); i++ {
		if lines[i] == '\n' || lines[i] == '\r' {
			if key != "" && value != "" {
				result[key] = value
			}
			key = ""
			value = ""
			parsingKey = true
			continue
		}

		if lines[i] == ':' {
			parsingKey = false
			continue
		}

		if lines[i] == '#' {
			// Skip comment lines
			for i < len(lines) && lines[i] != '\n' && lines[i] != '\r' {
				i++
			}
			continue
		}

		if parsingKey {
			key += string(lines[i])
		} else {
			value += string(lines[i])
		}
	}

	return result
}

func main() {
	mode := flag.String("mode", "oreo", "Mode: oreo or native")
	numRecords := flag.Int("records", 1000, "Number of records to insert")
	batchSize := flag.Int("batch", 10, "Number of records per transaction (oreo mode only)")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	if *mode == "oreo" {
		config := &discovery.ServiceDiscoveryConfig{
			Type: discovery.HTTPDiscovery,
			HTTP: &discovery.HTTPDiscoveryConfig{
				RegistryPort: ":9000",
			},
		}
		client, err := network.NewClient(config)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Waiting for the connection of executors...\n")
		time.Sleep(3 * time.Second)

		oracle := timesource.NewGlobalTimeSource("http://localhost:8010")

		err = insertViaOreo(client, oracle, *numRecords, *batchSize)
		if err != nil {
			panic(err)
		}

		time.Sleep(2 * time.Second)
		getMongoStats("oreo")
		getRedisStats(0, "Oreo")

	} else if *mode == "native" {
		err := insertViaNative(*numRecords)
		if err != nil {
			panic(err)
		}

		time.Sleep(2 * time.Second)
		getMongoStats("oreo_native")
		getRedisStats(1, "Native")

	} else {
		fmt.Printf("Invalid mode. Use 'oreo' or 'native'\n")
	}
}
