package main

import (
	"strings"

	"github.com/kkkzoz/oreo/pkg/datastore/cassandra"
	"github.com/kkkzoz/oreo/pkg/datastore/couchdb"
	"github.com/kkkzoz/oreo/pkg/datastore/dynamodb"
	"github.com/kkkzoz/oreo/pkg/datastore/mongo"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/datastore/tikv"
	"github.com/kkkzoz/oreo/pkg/logger"
	"github.com/kkkzoz/oreo/pkg/txn"
)

func getConnMap(wType string, dbComb string) map[string]txn.Connector {
	connMap := make(map[string]txn.Connector)
	logger.Infow("Setting up database connections", "workload", wType, "dbCombination", dbComb)

	switch wType {
	case "iot":
		if benConfig.MongoDBAddr1 != "" {
			connMap["MongoDB"] = getMongoConn(1)
		} else {
			logger.Warn("MongoDBAddr1 not configured for iot workload")
		}
		if benConfig.RedisAddr != "" {
			connMap["Redis"] = getRedisConn(1)
		} else {
			logger.Warn("RedisAddr not configured for iot workload")
		}
	case "social":
		if benConfig.MongoDBAddr1 != "" {
			connMap["MongoDB"] = getMongoConn(1)
		} else {
			logger.Warn("MongoDBAddr1 not configured for social workload")
		}
		if benConfig.RedisAddr != "" {
			connMap["Redis"] = getRedisConn(1)
		} else {
			logger.Warn("RedisAddr not configured for social workload")
		}
		if len(benConfig.CassandraAddr) > 0 {
			connMap["Cassandra"] = getCassandraConn()
		} else {
			logger.Warn("CassandraAddr not configured for social workload")
		}
	case "order":
		if benConfig.MongoDBAddr1 != "" {
			connMap["MongoDB"] = getMongoConn(1)
		} else {
			logger.Warn("MongoDBAddr1 not configured for order workload")
		}
		if benConfig.KVRocksAddr != "" {
			connMap["KVRocks"] = getKVRocksConn()
		} else {
			logger.Warn("KVRocksAddr not configured for order workload")
		}
		if benConfig.RedisAddr != "" {
			connMap["Redis"] = getRedisConn(1)
		} else {
			logger.Warn("RedisAddr not configured for order workload")
		}
		if len(benConfig.CassandraAddr) > 0 {
			connMap["Cassandra"] = getCassandraConn()
		} else {
			logger.Warn("CassandraAddr not configured for order workload")
		}
	case "ycsb":
		dbList := strings.Split(dbComb, ",")
		for _, db := range dbList {
			db = strings.TrimSpace(db) // Trim whitespace
			if db == "" {
				continue
			}
			logger.Infow("Configuring database for YCSB", "db", db)
			switch db {
			case "Redis":
				if benConfig.RedisAddr != "" {
					connMap["Redis"] = getRedisConn(1)
				} else {
					logger.Warn("RedisAddr not configured despite being requested in --db")
				}
			case "MongoDB1":
				if benConfig.MongoDBAddr1 != "" {
					connMap["MongoDB1"] = getMongoConn(1)
				} else {
					logger.Warn("MongoDBAddr1 not configured despite being requested in --db")
				}
			case "MongoDB2":
				if benConfig.MongoDBAddr2 != "" {
					connMap["MongoDB2"] = getMongoConn(2)
				} else {
					logger.Warn("MongoDBAddr2 not configured despite being requested in --db")
				}
			case "KVRocks":
				if benConfig.KVRocksAddr != "" {
					connMap["KVRocks"] = getKVRocksConn()
				} else {
					logger.Warn("KVRocksAddr not configured despite being requested in --db")
				}
			case "CouchDB":
				if benConfig.CouchDBAddr != "" {
					connMap["CouchDB"] = getCouchConn()
				} else {
					logger.Warn("CouchDBAddr not configured despite being requested in --db")
				}
			case "Cassandra":
				if len(benConfig.CassandraAddr) > 0 {
					connMap["Cassandra"] = getCassandraConn()
				} else {
					logger.Warn("CassandraAddr not configured despite being requested in --db")
				}
			case "DynamoDB":
				if benConfig.DynamoDBAddr != "" {
					connMap["DynamoDB"] = getDynamoConn()
				} else {
					logger.Warn("DynamoDBAddr not configured despite being requested in --db")
				}
			case "TiKV":
				if len(benConfig.TiKVAddr) > 0 {
					connMap["TiKV"] = getTiKVConn()
				} else {
					logger.Warn("TiKVAddr not configured despite being requested in --db")
				}
			default:
				logger.Errorf("Invalid database name '%s' in --db combination", db)
			}
		}
	default:
		logger.Fatalf("Unsupported workload type: %s", wType)
	}

	if len(connMap) == 0 {
		logger.Warnw(
			"No database connections were successfully configured based on workload and config",
			"workload",
			wType,
			"dbCombination",
			dbComb,
		)
	} else {
		dsNames := make([]string, 0, len(connMap))
		for name := range connMap {
			dsNames = append(dsNames, name)
		}
		logger.Infow("Database connections established", "datastores", dsNames)
	}
	return connMap
}

// --- Individual Database Connection Helpers ---

func getKVRocksConn() *redis.RedisConnection {
	logger.Infow("Connecting to KVRocks", "address", benConfig.KVRocksAddr)
	kvConn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address:  benConfig.KVRocksAddr,
		Password: benConfig.RedisPassword, // Assuming same password config as Redis
		PoolSize: *poolSize,
	})
	err := kvConn.Connect()
	if err != nil {
		logger.Fatalw(
			"Failed to connect to KVRocks",
			"address",
			benConfig.KVRocksAddr,
			"error",
			err,
		)
	}
	// Optional: Ping check
	// if _, err := kvConn.Get("ping"); err != nil { // Use GET for KVRocks? Or specific PING?
	// 	logger.Warnw("KVRocks ping failed", "address", benConfig.KVRocksAddr, "error", err)
	// }
	logger.Info("Connected to KVRocks successfully.")
	return kvConn
}

func getCouchConn() *couchdb.CouchDBConnection {
	logger.Infow("Connecting to CouchDB", "address", benConfig.CouchDBAddr, "dbName", "oreo")
	couchConn := couchdb.NewCouchDBConnection(&couchdb.ConnectionOptions{
		Address:  benConfig.CouchDBAddr,
		Username: benConfig.CouchDBUsername, // Use configured username/password
		Password: benConfig.CouchDBPassword,
		DBName:   "oreo", // Hardcoded DB name? Consider making configurable
	})
	err := couchConn.Connect()
	if err != nil {
		logger.Fatalw(
			"Failed to connect to CouchDB",
			"address",
			benConfig.CouchDBAddr,
			"error",
			err,
		)
	}
	logger.Info("Connected to CouchDB successfully.")
	return couchConn
}

func getMongoConn(id int) *mongo.MongoConnection {
	var address string
	switch id {
	case 1:
		address = benConfig.MongoDBAddr1
	case 2:
		address = benConfig.MongoDBAddr2
	default:
		logger.Fatalf("Invalid MongoDB connection ID requested: %d", id)
		return nil // Should not be reached
	}
	logger.Infow(
		"Connecting to MongoDB",
		"id",
		id,
		"address",
		address,
		"dbName",
		"oreo",
		"collection",
		"benchmark",
	)
	mongoConn := mongo.NewMongoConnection(&mongo.ConnectionOptions{
		Address:        address,
		DBName:         "oreo",      // Hardcoded DB name?
		CollectionName: "benchmark", // Hardcoded collection?
		Username:       benConfig.MongoDBUsername,
		Password:       benConfig.MongoDBPassword,
		// PoolSize not directly configurable here, managed by driver
	})
	err := mongoConn.Connect()
	if err != nil {
		logger.Fatalw("Failed to connect to MongoDB", "id", id, "address", address, "error", err)
	}
	logger.Infow("Connected to MongoDB successfully.", "id", id)
	return mongoConn
}

func getRedisConn(id int) *redis.RedisConnection {
	var address string
	switch id {
	case 1:
		address = benConfig.RedisAddr
	// Add cases for RedisAddr2, etc. if needed
	default:
		logger.Fatalf("Invalid Redis connection ID requested: %d", id)
		return nil // Should not be reached
	}
	logger.Infow("Connecting to Redis", "id", id, "address", address)
	redisConn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address:  address,
		Password: benConfig.RedisPassword,
		PoolSize: *poolSize,
	})
	err := redisConn.Connect() // Connect attempts to ping
	if err != nil {
		logger.Fatalw("Failed to connect to Redis", "id", id, "address", address, "error", err)
	}
	logger.Infow("Connected to Redis successfully.", "id", id)
	return redisConn
}

func getCassandraConn() *cassandra.CassandraConnection {
	logger.Infow("Connecting to Cassandra", "hosts", benConfig.CassandraAddr, "keyspace", "oreo")
	cassConn := cassandra.NewCassandraConnection(&cassandra.ConnectionOptions{
		Hosts:    benConfig.CassandraAddr,
		Keyspace: "oreo", // Hardcoded keyspace?
		Username: benConfig.CassandraUsername,
		Password: benConfig.CassandraPassword,
		// PoolSize (NumConns) might be configurable via options if needed
	})
	err := cassConn.Connect()
	if err != nil {
		logger.Fatalw(
			"Failed to connect to Cassandra",
			"hosts",
			benConfig.CassandraAddr,
			"error",
			err,
		)
	}
	logger.Info("Connected to Cassandra successfully.")
	return cassConn
}

func getDynamoConn() *dynamodb.DynamoDBConnection {
	logger.Infow("Connecting to DynamoDB", "endpoint", benConfig.DynamoDBAddr, "tableName", "oreo")
	dynamoConn := dynamodb.NewDynamoDBConnection(&dynamodb.ConnectionOptions{
		TableName: "oreo",                 // Hardcoded table name?
		Endpoint:  benConfig.DynamoDBAddr, // Use Endpoint for local/mock, otherwise relies on AWS SDK defaults
		// Credentials handled by AWS SDK (env vars, instance profile, etc.)
	})
	err := dynamoConn.Connect() // Connect likely initializes the client
	if err != nil {
		logger.Fatalw(
			"Failed to connect to DynamoDB",
			"endpoint",
			benConfig.DynamoDBAddr,
			"error",
			err,
		)
	}
	logger.Info("Connected to DynamoDB successfully.")
	return dynamoConn
}

func getTiKVConn() *tikv.TiKVConnection {
	logger.Infow("Connecting to TiKV", "pdAddrs", benConfig.TiKVAddr)
	tikvConn := tikv.NewTiKVConnection(&tikv.ConnectionOptions{
		PDAddrs: benConfig.TiKVAddr,
		// Security options (TLS) might be needed here via config
	})
	err := tikvConn.Connect() // Connect initializes the client
	if err != nil {
		logger.Fatalw("Failed to connect to TiKV", "pdAddrs", benConfig.TiKVAddr, "error", err)
	}
	logger.Info("Connected to TiKV successfully.")
	return tikvConn
}

// getMapKeys is a small helper for logging map keys without values
func getMapKeys(m map[string]txn.PredicateInfo) []string {
	if m == nil {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
