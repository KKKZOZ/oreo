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
	"github.com/kkkzoz/oreo/pkg/network"
	"github.com/kkkzoz/oreo/pkg/timesource"
	"github.com/kkkzoz/oreo/pkg/txn"
	"gopkg.in/yaml.v2"
)

// Config 配置结构体
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

// IoT 设备结构体
type Device struct {
	DeviceID     string    `json:"device_id"`
	DeviceName   string    `json:"device_name"`
	DeviceType   string    `json:"device_type"`
	Location     string    `json:"location"`
	Status       string    `json:"status"`
	RegisteredAt time.Time `json:"registered_at"`
}

// IoT 数据结构体
type SensorData struct {
	DeviceID   string    `json:"device_id"`
	SensorType string    `json:"sensor_type"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
	Timestamp  time.Time `json:"timestamp"`
	Location   string    `json:"location"`
}

// 设备统计信息
type DeviceStats struct {
	DeviceID      string    `json:"device_id"`
	TotalReadings int       `json:"total_readings"`
	LastReading   time.Time `json:"last_reading"`
	AvgValue      float64   `json:"avg_value"`
	MinValue      float64   `json:"min_value"`
	MaxValue      float64   `json:"max_value"`
}

// 全局变量
var (
	client             *network.Client
	oracle             timesource.TimeSourcer
	redisDatastore     txn.Datastorer
	mongoDatastore     txn.Datastorer
	cassandraDatastore txn.Datastorer
	config             Config
)

// 加载配置文件
func loadConfig(configPath string) error {
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	// 设置默认值
	if config.ServerPort == "" {
		config.ServerPort = ":8080"
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

// 初始化数据库连接
func initConnections() error {
	var err error

	// 从registry_addr中提取端口
	registryPort := ":9000" // 默认端口
	if strings.Contains(config.RegistryAddr, ":") {
		// 如果是完整URL格式，提取端口部分
		if strings.HasPrefix(config.RegistryAddr, "http://") {
			// 从 "http://localhost:9000" 提取 ":9000"
			parts := strings.Split(config.RegistryAddr, ":")
			if len(parts) >= 3 {
				registryPort = ":" + parts[2]
			}
		} else if strings.HasPrefix(config.RegistryAddr, ":") {
			// 如果已经是端口格式，直接使用
			registryPort = config.RegistryAddr
		}
	}

	// 初始化网络客户端
	client, err = network.NewClient(registryPort)
	if err != nil {
		return fmt.Errorf("failed to create network client: %v", err)
	}

	// 初始化时间源
	oracle = timesource.NewGlobalTimeSource(config.TimeOracleURL)

	// 初始化Redis连接 - 用于缓存设备状态和实时数据
	redisConn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address:  config.RedisAddr,
		Password: config.RedisPassword,
	})

	// 初始化MongoDB连接 - 用于存储设备信息和历史数据
	mongoConn := mongo.NewMongoConnection(&mongo.ConnectionOptions{
		Address:        config.MongoDBAddr,
		Username:       config.MongoDBUsername,
		Password:       config.MongoDBPassword,
		DBName:         config.MongoDBDBName,
		CollectionName: config.MongoDBCollectionName,
	})

	// 初始化Cassandra连接
	cassandraConn := cassandra.NewCassandraConnection(&cassandra.ConnectionOptions{
		Hosts:    config.CassandraHosts,
		Keyspace: config.CassandraKeyspace,
		Username: config.CassandraUsername,
		Password: config.CassandraPassword,
	})

	// 连接到数据库
	log.Println("Attempting to connect to databases...")

	// 尝试连接 Redis
	if len(config.RedisAddr) > 0 {
		if err := redisConn.Connect(); err != nil {
			log.Printf("Warning: Failed to connect to Redis: %v", err)
			redisDatastore = nil
		} else {
			log.Println("Redis connected successfully")
			redisDatastore = redis.NewRedisDatastore("Redis", redisConn)
		}
	} else {
		log.Println("Warning: Failed to connect to Redis: no address provided")
		redisDatastore = nil
	}

	// 尝试连接 MongoDB
	if len(config.MongoDBAddr) > 0 {
		if err := mongoConn.Connect(); err != nil {
			log.Printf("Warning: Failed to connect to MongoDB: %v", err)
			mongoDatastore = nil
		} else {
			log.Println("MongoDB connected successfully")
			mongoDatastore = mongo.NewMongoDatastore("MongoDB1", mongoConn)
		}
	} else {
		log.Println("Warning: Failed to connect to MongoDB: no URI provided")
		mongoDatastore = nil
	}

	// 尝试连接Cassandra
	if len(config.CassandraHosts) > 0 {
		if err := cassandraConn.Connect(); err != nil {
			log.Printf("Warning: Failed to connect to Cassandra: %v", err)
			cassandraDatastore = nil
		} else {
			log.Println("Cassandra connected successfully")
			cassandraDatastore = cassandra.NewCassandraDatastore("Cassandra", cassandraConn)
		}
	} else {
		log.Println("Warning: Failed to connect to Cassandra: no hosts provided")
		cassandraDatastore = nil
	}

	// 检查至少有一个数据库连接成功
	if redisDatastore == nil && mongoDatastore == nil && cassandraDatastore == nil {
		return fmt.Errorf("failed to connect to any database")
	}

	log.Println("Database initialization completed")

	return nil
}

// 设备注册 API
func registerDevice(c *fiber.Ctx) error {
	var device Device
	if err := c.BodyParser(&device); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// 设置注册时间
	device.RegisteredAt = time.Now()
	device.Status = "active"

	// 创建分布式事务
	txn := txn.NewTransactionWithRemote(client, oracle)

	// 添加所有可用的数据存储
	if redisDatastore != nil && mongoDatastore != nil && cassandraDatastore != nil {
		txn.AddDatastores(redisDatastore, mongoDatastore, cassandraDatastore)
	} else if redisDatastore != nil && mongoDatastore != nil {
		txn.AddDatastores(redisDatastore, mongoDatastore)
	} else if redisDatastore != nil && cassandraDatastore != nil {
		txn.AddDatastores(redisDatastore, cassandraDatastore)
	} else if mongoDatastore != nil && cassandraDatastore != nil {
		txn.AddDatastores(mongoDatastore, cassandraDatastore)
	} else if redisDatastore != nil {
		txn.AddDatastores(redisDatastore)
	} else if mongoDatastore != nil {
		txn.AddDatastores(mongoDatastore)
	} else if cassandraDatastore != nil {
		txn.AddDatastores(cassandraDatastore)
	}

	if err := txn.Start(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to start transaction: %v", err),
		})
	}

	// 在Redis中缓存设备状态
	if redisDatastore != nil {
		deviceStatusKey := fmt.Sprintf("device:status:%s", device.DeviceID)
		if err := txn.Write("Redis", deviceStatusKey, device.Status); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to cache device status: %v", err),
			})
		}
	}

	// 在MongoDB中存储设备信息
	if mongoDatastore != nil {
		deviceInfoKey := fmt.Sprintf("device:info:%s", device.DeviceID)
		deviceJSON, _ := json.Marshal(device)
		if err := txn.Write("MongoDB1", deviceInfoKey, string(deviceJSON)); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to store device info: %v", err),
			})
		}
	}

	// 在Cassandra中初始化设备统计
	if cassandraDatastore != nil {
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

	// 提交事务
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

// 数据上报 API
func reportSensorData(c *fiber.Ctx) error {
	var sensorData SensorData
	if err := c.BodyParser(&sensorData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// 设置时间戳
	sensorData.Timestamp = time.Now()

	// 创建分布式事务
	txn := txn.NewTransactionWithRemote(client, oracle)

	// 添加所有可用的数据存储
	if redisDatastore != nil && mongoDatastore != nil && cassandraDatastore != nil {
		txn.AddDatastores(redisDatastore, mongoDatastore, cassandraDatastore)
	} else if redisDatastore != nil && mongoDatastore != nil {
		txn.AddDatastores(redisDatastore, mongoDatastore)
	} else if redisDatastore != nil && cassandraDatastore != nil {
		txn.AddDatastores(redisDatastore, cassandraDatastore)
	} else if mongoDatastore != nil && cassandraDatastore != nil {
		txn.AddDatastores(mongoDatastore, cassandraDatastore)
	} else if redisDatastore != nil {
		txn.AddDatastores(redisDatastore)
	} else if mongoDatastore != nil {
		txn.AddDatastores(mongoDatastore)
	} else if cassandraDatastore != nil {
		txn.AddDatastores(cassandraDatastore)
	}

	if err := txn.Start(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to start transaction: %v", err),
		})
	}

	// 在Redis中更新最新数据
	sensorJSON, _ := json.Marshal(sensorData)
	if redisDatastore != nil {
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

	// 在MongoDB中存储历史数据
	if mongoDatastore != nil {
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

	// 读取并更新Cassandra中的统计信息
	if cassandraDatastore != nil {
		statsKey := fmt.Sprintf("device:stats:%s", sensorData.DeviceID)
		var statsJSON string
		if err := txn.Read("Cassandra", statsKey, &statsJSON); err != nil {
			// 如果统计信息不存在，创建新的
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
			// 更新现有统计信息
			var stats DeviceStats
			if err := json.Unmarshal([]byte(statsJSON), &stats); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to unmarshal stats: %v", err),
				})
			}

			// 更新统计数据
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

	// 提交事务
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

// 获取设备信息 API
func getDeviceInfo(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	if deviceID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Device ID is required",
		})
	}

	// 创建只读事务
	txn := txn.NewTransactionWithRemote(client, oracle)

	// 添加所有可用的数据存储
	if redisDatastore != nil && mongoDatastore != nil && cassandraDatastore != nil {
		txn.AddDatastores(redisDatastore, mongoDatastore, cassandraDatastore)
	} else if redisDatastore != nil && mongoDatastore != nil {
		txn.AddDatastores(redisDatastore, mongoDatastore)
	} else if redisDatastore != nil && cassandraDatastore != nil {
		txn.AddDatastores(redisDatastore, cassandraDatastore)
	} else if mongoDatastore != nil && cassandraDatastore != nil {
		txn.AddDatastores(mongoDatastore, cassandraDatastore)
	} else if redisDatastore != nil {
		txn.AddDatastores(redisDatastore)
	} else if mongoDatastore != nil {
		txn.AddDatastores(mongoDatastore)
	} else if cassandraDatastore != nil {
		txn.AddDatastores(cassandraDatastore)
	}

	if err := txn.Start(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to start transaction: %v", err),
		})
	}

	// 从Redis读取设备状态
	deviceStatusKey := fmt.Sprintf("device:status:%s", deviceID)
	var status string
	if err := txn.Read("Redis", deviceStatusKey, &status); err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Device not found",
		})
	}

	// 从MongoDB读取设备信息
	deviceInfoKey := fmt.Sprintf("device:info:%s", deviceID)
	var deviceJSON string
	if err := txn.Read("MongoDB1", deviceInfoKey, &deviceJSON); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to read device info: %v", err),
		})
	}

	// 从Cassandra读取统计信息
	var statsJSON string
	var hasStats bool
	if cassandraDatastore != nil {
		statsKey := fmt.Sprintf("device:stats:%s", deviceID)
		if err := txn.Read("Cassandra", statsKey, &statsJSON); err != nil {
			// 统计信息不存在或读取失败，继续处理
			hasStats = false
		} else {
			hasStats = true
		}
	}

	// 提交事务
	if err := txn.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to commit transaction: %v", err),
		})
	}

	// 解析数据
	var device Device
	if err := json.Unmarshal([]byte(deviceJSON), &device); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to unmarshal device data: %v", err),
		})
	}

	response := fiber.Map{
		"device": device,
	}

	// 如果有统计信息，添加到响应中
	if hasStats {
		var stats DeviceStats
		if err := json.Unmarshal([]byte(statsJSON), &stats); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to unmarshal stats data: %v", err),
			})
		}
		response["stats"] = stats
	}

	return c.JSON(response)
}

// 获取设备最新数据 API
func getLatestSensorData(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	sensorType := c.Query("sensor_type")

	if deviceID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Device ID is required",
		})
	}

	// 创建只读事务
	txn := txn.NewTransactionWithRemote(client, oracle)

	// 添加所有可用的数据存储
	if redisDatastore != nil {
		txn.AddDatastores(redisDatastore)
	}

	if err := txn.Start(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to start transaction: %v", err),
		})
	}

	if sensorType != "" {
		// 获取特定传感器类型的最新数据
		latestDataKey := fmt.Sprintf("device:latest:%s:%s", deviceID, sensorType)
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
				"error": fmt.Sprintf("Failed to unmarshal sensor data: %v", err),
			})
		}
		return c.JSON(sensorData)
	} else {
		return c.Status(400).JSON(fiber.Map{
			"error": "sensor_type parameter is required",
		})
	}
}

// 批量数据处理 API
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

	// 创建分布式事务
	txn := txn.NewTransactionWithRemote(client, oracle)

	// 添加所有可用的数据存储
	if redisDatastore != nil && mongoDatastore != nil && cassandraDatastore != nil {
		txn.AddDatastores(redisDatastore, mongoDatastore, cassandraDatastore)
	} else if redisDatastore != nil && mongoDatastore != nil {
		txn.AddDatastores(redisDatastore, mongoDatastore)
	} else if redisDatastore != nil && cassandraDatastore != nil {
		txn.AddDatastores(redisDatastore, cassandraDatastore)
	} else if mongoDatastore != nil && cassandraDatastore != nil {
		txn.AddDatastores(mongoDatastore, cassandraDatastore)
	} else if redisDatastore != nil {
		txn.AddDatastores(redisDatastore)
	} else if mongoDatastore != nil {
		txn.AddDatastores(mongoDatastore)
	} else if cassandraDatastore != nil {
		txn.AddDatastores(cassandraDatastore)
	}

	if err := txn.Start(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to start transaction: %v", err),
		})
	}

	processedCount := 0
	for _, sensorData := range batchData {
		sensorData.Timestamp = time.Now()

		// 在Redis中更新最新数据
		latestDataKey := fmt.Sprintf(
			"device:latest:%s:%s",
			sensorData.DeviceID,
			sensorData.SensorType,
		)
		sensorJSON, _ := json.Marshal(sensorData)
		if err := txn.Write("Redis", latestDataKey, string(sensorJSON)); err != nil {
			continue
		}

		// 在MongoDB中存储历史数据
		historyKey := fmt.Sprintf(
			"sensor:history:%s:%d",
			sensorData.DeviceID,
			sensorData.Timestamp.Unix(),
		)
		if err := txn.Write("MongoDB1", historyKey, string(sensorJSON)); err != nil {
			continue
		}

		processedCount++
	}

	// 提交事务
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
	// 加载配置
	if err := loadConfig("config.yaml"); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 初始化连接
	if err := initConnections(); err != nil {
		log.Fatalf("Failed to initialize connections: %v", err)
	}

	// 等待执行器连接
	fmt.Println("Waiting for executor connections...")
	time.Sleep(3 * time.Second)

	// 创建Fiber应用
	app := fiber.New(fiber.Config{
		AppName: "IoT Platform API",
	})

	// 中间件
	app.Use(logger.New())
	app.Use(cors.New())

	// 健康检查
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now(),
		})
	})

	// API 路由
	api := app.Group("/api/v1")

	// 设备管理
	api.Post("/devices", registerDevice)
	api.Get("/devices/:deviceId", getDeviceInfo)

	// 数据管理
	api.Post("/data", reportSensorData)
	api.Get("/devices/:deviceId/latest", getLatestSensorData)
	api.Post("/data/batch", batchProcessData)

	// 启动服务器
	fmt.Printf("IoT Platform API server starting on port %s...\n", config.ServerPort)
	log.Fatal(app.Listen(config.ServerPort))
}
