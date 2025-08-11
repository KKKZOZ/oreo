package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/kkkzoz/oreo/pkg/datastore/cassandra"
	"github.com/kkkzoz/oreo/pkg/datastore/mongo"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/discovery"
	"github.com/kkkzoz/oreo/pkg/network"
	"github.com/kkkzoz/oreo/pkg/timesource"
	"github.com/kkkzoz/oreo/pkg/txn"
	"gopkg.in/yaml.v2"
)

// Config configuration structure
type Config struct {
	RegistryAddr          string   `yaml:"registry_addr"`
	TimeOracleURL         string   `yaml:"time_oracle_url"`
	ServerPort            string   `yaml:"server_port"`
	RedisAddr             string   `yaml:"redis_addr"`
	RedisPassword         string   `yaml:"redis_password"`
	MongoDBAddr           string   `yaml:"mongodb_addr"`
	MongoDBUsername       string   `yaml:"mongodb_username"`
	MongoDBPassword       string   `yaml:"mongodb_password"`
	MongoDBDBName         string   `yaml:"mongodb_db_name"`
	MongoDBCollectionName string   `yaml:"mongodb_collection_name"`
	CassandraHosts        []string `yaml:"cassandra_hosts"`
	CassandraKeyspace     string   `yaml:"cassandra_keyspace"`
	CassandraUsername     string   `yaml:"cassandra_username"`
	CassandraPassword     string   `yaml:"cassandra_password"`
}

// Device IoT device structure
type Device struct {
	DeviceID     string    `json:"device_id"`
	DeviceName   string    `json:"device_name"`
	DeviceType   string    `json:"device_type"`
	Location     string    `json:"location"`
	Status       string    `json:"status"`
	RegisteredAt time.Time `json:"registered_at"`
}

// SensorData IoT sensor data structure
type SensorData struct {
	DeviceID   string    `json:"device_id"`
	SensorType string    `json:"sensor_type"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
	Timestamp  time.Time `json:"timestamp"`
	Location   string    `json:"location"`
}

// DeviceStats device statistics information
type DeviceStats struct {
	DeviceID      string    `json:"device_id"`
	TotalReadings int       `json:"total_readings"`
	LastReading   time.Time `json:"last_reading"`
	AvgValue      float64   `json:"avg_value"`
	MinValue      float64   `json:"min_value"`
	MaxValue      float64   `json:"max_value"`
}

// Global variables
var (
	client        *network.Client
	oracle        timesource.TimeSourcer
	redisConn     *redis.RedisConnection
	mongoConn     *mongo.MongoConnection
	cassandraConn *cassandra.CassandraConnection
	config        Config
)

// loadConfig loads configuration file
func loadConfig(configPath string) error {
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	// Set default values
	if config.ServerPort == "" {
		config.ServerPort = ":8081"
	}
	if config.MongoDBDBName == "" {
		config.MongoDBDBName = "iot_platform"
	}
	if config.MongoDBCollectionName == "" {
		config.MongoDBCollectionName = "devices"
	}
	if config.CassandraKeyspace == "" {
		config.CassandraKeyspace = "iot_data"
	}

	return nil
}

// initConnections initializes database connections
func initConnections() error {
	var err error

	// Extract port from registry_addr
	registryPort := ":9000" // Default port
	if strings.Contains(config.RegistryAddr, ":") {
		// If it's a complete URL format, extract the port part
		if strings.HasPrefix(config.RegistryAddr, "http://") {
			// Extract ":9000" from "http://localhost:9000"
			parts := strings.Split(config.RegistryAddr, ":")
			if len(parts) >= 3 {
				registryPort = ":" + parts[2]
			}
		} else if strings.HasPrefix(config.RegistryAddr, ":") {
			// If it's already in port format, use it directly
			registryPort = config.RegistryAddr
		}
	}

	// Initialize network client
	networkConfig := &discovery.ServiceDiscoveryConfig{
		Type: discovery.HTTPDiscovery,
		HTTP: &discovery.HTTPDiscoveryConfig{
			RegistryPort: registryPort,
		},
	}
	client, err = network.NewClient(networkConfig)
	if err != nil {
		return fmt.Errorf("failed to create network client: %v", err)
	}

	// Initialize time source
	oracle = timesource.NewGlobalTimeSource(config.TimeOracleURL)

	// Initialize Redis connector
	if len(config.RedisAddr) > 0 {
		redisConn = redis.NewRedisConnection(&redis.ConnectionOptions{
			Address:  config.RedisAddr,
			Password: config.RedisPassword,
		})
		log.Println("Redis connector initialized")
	} else {
		log.Println("Warning: Redis not configured")
		redisConn = nil
	}

	// Initialize MongoDB connector
	if len(config.MongoDBAddr) > 0 {
		mongoConn = mongo.NewMongoConnection(&mongo.ConnectionOptions{
			Address:        config.MongoDBAddr,
			Username:       config.MongoDBUsername,
			Password:       config.MongoDBPassword,
			DBName:         config.MongoDBDBName,
			CollectionName: config.MongoDBCollectionName,
		})
		log.Println("MongoDB connector initialized")
	} else {
		log.Println("Warning: MongoDB not configured")
		mongoConn = nil
	}

	// Initialize Cassandra connector
	if len(config.CassandraHosts) > 0 {
		cassandraConn = cassandra.NewCassandraConnection(&cassandra.ConnectionOptions{
			Hosts:    config.CassandraHosts,
			Keyspace: config.CassandraKeyspace,
			Username: config.CassandraUsername,
			Password: config.CassandraPassword,
		})
		log.Println("Cassandra connector initialized")
	} else {
		log.Println("Warning: Cassandra not configured")
		cassandraConn = nil
	}

	log.Println("Database connectors initialization completed")

	return nil
}

// createDatastoresForTransaction creates datastores for transaction
func createDatastoresForTransaction() []txn.Datastorer {
	var datastores []txn.Datastorer

	if redisConn != nil {
		redisDatastore := redis.NewRedisDatastore("Redis", redisConn)
		datastores = append(datastores, redisDatastore)
	}

	if mongoConn != nil {
		mongoDatastore := mongo.NewMongoDatastore("MongoDB1", mongoConn)
		datastores = append(datastores, mongoDatastore)
	}

	if cassandraConn != nil {
		cassandraDatastore := cassandra.NewCassandraDatastore("Cassandra", cassandraConn)
		datastores = append(datastores, cassandraDatastore)
	}

	return datastores
}

// API handler functions

// registerDevice device registration API
func registerDevice(c *fiber.Ctx) error {
	var device Device
	if err := c.BodyParser(&device); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Set registration time
	device.RegisteredAt = time.Now()
	if device.Status == "" {
		device.Status = "active"
	}

	// Create distributed transaction
	txn := txn.NewTransactionWithRemote(client, oracle)

	// Create datastores for transaction
	datastores := createDatastoresForTransaction()
	txn.AddDatastores(datastores...)

	if err := txn.Start(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to start transaction: %v", err),
		})
	}

	// Cache device status in Redis
	if redisConn != nil {
		deviceStatusKey := fmt.Sprintf("device:status:%s", device.DeviceID)
		if err := txn.Write("Redis", deviceStatusKey, device.Status); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to cache device status: %v", err),
			})
		}
	}

	// Store device information in MongoDB
	if mongoConn != nil {
		deviceInfoKey := fmt.Sprintf("device:info:%s", device.DeviceID)
		deviceJSON, _ := json.Marshal(device)
		if err := txn.Write("MongoDB1", deviceInfoKey, string(deviceJSON)); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to store device info: %v", err),
			})
		}
	}

	// Initialize device statistics in Cassandra
	if cassandraConn != nil {
		stats := DeviceStats{
			DeviceID:      device.DeviceID,
			TotalReadings: 0,
			LastReading:   time.Now(),
			AvgValue:      0,
			MinValue:      0,
			MaxValue:      0,
		}
		statsKey := fmt.Sprintf("device:stats:%s", device.DeviceID)
		statsJSON, _ := json.Marshal(stats)
		if err := txn.Write("Cassandra", statsKey, string(statsJSON)); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to initialize device stats: %v", err),
			})
		}
	}

	// Commit transaction
	if err := txn.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to commit transaction: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Device registered successfully",
		"device":  device,
	})
}

// reportSensorData data reporting API
func reportSensorData(c *fiber.Ctx) error {
	var sensorData SensorData
	if err := c.BodyParser(&sensorData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Set timestamp
	sensorData.Timestamp = time.Now()

	// Create distributed transaction
	txn := txn.NewTransactionWithRemote(client, oracle)

	// Create datastores for transaction
	datastores := createDatastoresForTransaction()
	txn.AddDatastores(datastores...)

	if err := txn.Start(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to start transaction: %v", err),
		})
	}

	// Update latest data in Redis
	sensorJSON, _ := json.Marshal(sensorData)
	if redisConn != nil {
		latestDataKey := fmt.Sprintf(
			"device:latest:%s:%s",
			sensorData.DeviceID,
			sensorData.SensorType,
		)
		if err := txn.Write("Redis", latestDataKey, string(sensorJSON)); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to update latest data: %v", err),
			})
		}
	}

	// Store historical data in MongoDB
	if mongoConn != nil {
		// Store historical data in MongoDB
		historyKey := fmt.Sprintf(
			"sensor:history:%s:%d",
			sensorData.DeviceID,
			sensorData.Timestamp.Unix(),
		)
		if err := txn.Write("MongoDB1", historyKey, string(sensorJSON)); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to store history data: %v", err),
			})
		}
	}

	// Read and update statistics in Cassandra
	if cassandraConn != nil {
		statsKey := fmt.Sprintf("device:stats:%s", sensorData.DeviceID)
		var statsJSON string
		if err := txn.Read("Cassandra", statsKey, &statsJSON); err != nil {
			// If statistics don't exist, create new ones
			stats := DeviceStats{
				DeviceID:      sensorData.DeviceID,
				TotalReadings: 1,
				LastReading:   sensorData.Timestamp,
				AvgValue:      sensorData.Value,
				MinValue:      sensorData.Value,
				MaxValue:      sensorData.Value,
			}
			newStatsJSON, _ := json.Marshal(stats)
			if err := txn.Write("Cassandra", statsKey, string(newStatsJSON)); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to create device stats: %v", err),
				})
			}
		} else {
			// Update existing statistics
			var stats DeviceStats
			if err := json.Unmarshal([]byte(statsJSON), &stats); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to parse device stats: %v", err),
				})
			}

			// Update statistical data
			stats.TotalReadings++
			stats.LastReading = sensorData.Timestamp
			stats.AvgValue = (stats.AvgValue*float64(stats.TotalReadings-1) + sensorData.Value) / float64(stats.TotalReadings)
			if sensorData.Value < stats.MinValue {
				stats.MinValue = sensorData.Value
			}
			if sensorData.Value > stats.MaxValue {
				stats.MaxValue = sensorData.Value
			}

			updatedStatsJSON, _ := json.Marshal(stats)
			if err := txn.Write("Cassandra", statsKey, string(updatedStatsJSON)); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to update device stats: %v", err),
				})
			}
		}
	}

	// Commit transaction
	if err := txn.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to commit transaction: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Sensor data reported successfully",
		"data":    sensorData,
	})
}

// getDeviceInfo get device information API
func getDeviceInfo(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	if deviceID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Device ID is required",
		})
	}

	// Create read-only transaction
	txn := txn.NewTransactionWithRemote(client, oracle)

	// Create datastores for transaction
	datastores := createDatastoresForTransaction()
	txn.AddDatastores(datastores...)

	if err := txn.Start(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to start transaction: %v", err),
		})
	}

	// Read device status from Redis
	deviceStatusKey := fmt.Sprintf("device:status:%s", deviceID)
	var status string
	if err := txn.Read("Redis", deviceStatusKey, &status); err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Device not found",
		})
	}

	// Read device information from MongoDB
	deviceInfoKey := fmt.Sprintf("device:info:%s", deviceID)
	var deviceJSON string
	if err := txn.Read("MongoDB1", deviceInfoKey, &deviceJSON); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to read device info: %v", err),
		})
	}

	// Read statistics from Cassandra
	var statsJSON string
	var hasStats bool
	if cassandraConn != nil {
		statsKey := fmt.Sprintf("device:stats:%s", deviceID)
		if err := txn.Read("Cassandra", statsKey, &statsJSON); err != nil {
			// Statistics don't exist or read failed, continue processing
			hasStats = false
		} else {
			hasStats = true
		}
	}

	// Commit transaction
	if err := txn.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to commit transaction: %v", err),
		})
	}

	// Parse data
	var device Device
	if err := json.Unmarshal([]byte(deviceJSON), &device); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to parse device info: %v", err),
		})
	}

	response := fiber.Map{
		"device": device,
	}

	// If statistics exist, add to response
	if hasStats {
		var stats DeviceStats
		if err := json.Unmarshal([]byte(statsJSON), &stats); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to parse device stats: %v", err),
			})
		}
		response["stats"] = stats
	}

	return c.JSON(response)
}

// getLatestSensorData get device latest data API
func getLatestSensorData(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	sensorType := c.Query("sensor_type")

	if deviceID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Device ID is required",
		})
	}

	// Create read-only transaction
	txn := txn.NewTransactionWithRemote(client, oracle)

	// Only need Redis datastore
	if redisConn != nil {
		redisDatastore := redis.NewRedisDatastore("Redis", redisConn)
		txn.AddDatastores(redisDatastore)
	}

	if err := txn.Start(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to start transaction: %v", err),
		})
	}

	if sensorType != "" {
		// Get latest data for specific sensor type
		latestDataKey := fmt.Sprintf(
			"device:latest:%s:%s",
			deviceID,
			sensorType,
		)
		var sensorJSON string
		if err := txn.Read("Redis", latestDataKey, &sensorJSON); err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Sensor data not found",
			})
		}

		if err := txn.Commit(); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to commit transaction: %v", err),
			})
		}

		var sensorData SensorData
		if err := json.Unmarshal([]byte(sensorJSON), &sensorData); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to parse sensor data: %v", err),
			})
		}
		return c.JSON(sensorData)
	} else {
		return c.Status(400).JSON(fiber.Map{
			"error": "sensor_type parameter is required",
		})
	}
}

// batchProcessData batch data processing API
func batchProcessData(c *fiber.Ctx) error {
	var batchData []SensorData
	if err := c.BodyParser(&batchData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(batchData) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "No data provided",
		})
	}

	// Create distributed transaction
	txn := txn.NewTransactionWithRemote(client, oracle)

	// Create datastores for transaction
	datastores := createDatastoresForTransaction()
	txn.AddDatastores(datastores...)

	if err := txn.Start(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to start transaction: %v", err),
		})
	}

	processedCount := 0
	for _, sensorData := range batchData {
		sensorData.Timestamp = time.Now()

		// Update latest data in Redis
		latestDataKey := fmt.Sprintf(
			"device:latest:%s:%s",
			sensorData.DeviceID,
			sensorData.SensorType)
		sensorJSON, _ := json.Marshal(sensorData)
		if err := txn.Write("Redis", latestDataKey, string(sensorJSON)); err != nil {
			continue
		}

		// Store historical data in MongoDB
		historyKey := fmt.Sprintf(
			"sensor:history:%s:%d",
			sensorData.DeviceID,
			sensorData.Timestamp.Unix())
		if err := txn.Write("MongoDB1", historyKey, string(sensorJSON)); err != nil {
			continue
		}

		processedCount++
	}

	// Commit transaction
	if err := txn.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to commit transaction: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message":         "Batch data processed successfully",
		"processed_count": processedCount,
		"total_count":     len(batchData),
	})
}

func main() {
	// Load configuration
	if err := loadConfig("config.yaml"); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize connections
	if err := initConnections(); err != nil {
		log.Fatalf("Failed to initialize connections: %v", err)
	}

	// Wait for executor connections
	fmt.Println("Waiting for executor connections...")
	time.Sleep(3 * time.Second)

	// Create Fiber application
	app := fiber.New(fiber.Config{
		AppName: "IoT Platform API",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now(),
		})
	})

	// API routes
	api := app.Group("/api/v1")

	// Device management
	api.Post("/devices", registerDevice)
	api.Get("/devices/:deviceId", getDeviceInfo)

	// Data management
	api.Post("/data", reportSensorData)
	api.Get("/devices/:deviceId/latest", getLatestSensorData)
	api.Post("/data/batch", batchProcessData)

	// Start server
	fmt.Printf("IoT Platform API server starting on port %s...\n", config.ServerPort)
	log.Fatal(app.Listen(config.ServerPort))
}
