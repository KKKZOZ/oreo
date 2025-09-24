package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/discovery"
	"github.com/kkkzoz/oreo/pkg/network"
	"github.com/kkkzoz/oreo/pkg/timesource"
	"github.com/kkkzoz/oreo/pkg/txn"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ServiceDiscovery struct {
		HTTP struct {
			RegistryAddr string `yaml:"registry_addr"`
			RegistryPort string `yaml:"registry_port"`
		} `yaml:"http"`
		Etcd struct {
			Endpoints   []string `yaml:"endpoints"`
			Username    string   `yaml:"username"`
			Password    string   `yaml:"password"`
			DialTimeout string   `yaml:"dial_timeout"`
			KeyPrefix   string   `yaml:"key_prefix"`
		} `yaml:"etcd"`
	} `yaml:"service_discovery"`
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
	TimeOracleURL string `yaml:"time_oracle_url"`
	ServerPort    string `yaml:"server_port"`
}

type TestData struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	client        *network.Client
	oracle        timesource.TimeSourcer
	redisConn     *redis.RedisConnection
	config        Config
	discoveryType = flag.String(
		"discovery",
		"HTTP",
		"discover type (http or etcd)",
	)
	configPath = flag.String("config", "./client-config-9000.yaml", "path to config file")
)

func loadConfig() error {
	data, err := os.ReadFile(*configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	log.Printf("Loaded config: %+v\n", config)
	return nil
}

func initConnections() error {
	var err error

	log.Printf("Initializing service discovery with type: %s", *discoveryType)

	// Create service discovery configuration
	var discoveryConfig *discovery.ServiceDiscoveryConfig

	switch *discoveryType {
	case "http":
		log.Println("Using HTTP service discovery")
		log.Printf(
			"Starting as HTTP registry server on port %s",
			config.ServiceDiscovery.HTTP.RegistryPort,
		)
		discoveryConfig = &discovery.ServiceDiscoveryConfig{
			Type: discovery.HTTPDiscovery,
			HTTP: &discovery.HTTPDiscoveryConfig{
				RegistryPort: config.ServiceDiscovery.HTTP.RegistryPort,
			},
		}
	case "etcd":
		log.Printf(
			"Using etcd service discovery with endpoints: %v",
			config.ServiceDiscovery.Etcd.Endpoints,
		)
		log.Printf("Etcd key prefix: %s", config.ServiceDiscovery.Etcd.KeyPrefix)
		discoveryConfig = &discovery.ServiceDiscoveryConfig{
			Type: discovery.EtcdDiscovery,
			Etcd: &discovery.EtcdDiscoveryConfig{
				Endpoints: config.ServiceDiscovery.Etcd.Endpoints,
				KeyPrefix: config.ServiceDiscovery.Etcd.KeyPrefix,
			},
		}
	default:
		log.Printf(
			"Unknown service discovery type '%s', falling back to HTTP",
			*discoveryType,
		)
		// Fallback to HTTP service discovery
		discoveryConfig = &discovery.ServiceDiscoveryConfig{
			Type: discovery.HTTPDiscovery,
			HTTP: &discovery.HTTPDiscoveryConfig{
				RegistryPort: ":9000",
			},
		}
	}

	// Create network client
	log.Println("Creating network client with service discovery...")
	client, err = network.NewClient(discoveryConfig)
	if err != nil {
		return fmt.Errorf("failed to create network client: %v", err)
	}
	log.Printf(
		"Network client created successfully with %s service discovery",
		discoveryConfig.Type,
	)

	time.Sleep(2 * time.Second) // Wait for service discovery to stabilize

	redisAddr, redisErr := client.GetServerAddr("Redis")
	log.Printf("Redis server address: %v, error: %v", redisAddr, redisErr) // Test service discovery

	// Initialize time source
	oracle = timesource.NewGlobalTimeSource(config.TimeOracleURL)

	// Initialize Redis connection
	log.Printf("Initializing Redis connection to %s", config.RedisAddr)
	redisConn = redis.NewRedisConnection(&redis.ConnectionOptions{
		Address:  config.RedisAddr,
		Password: config.RedisPassword,
	})
	if err := redisConn.Connect(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		redisConn = nil
	} else {
		log.Println("Redis connection established successfully")
	}

	log.Println("Database connectors initialization completed")
	return nil
}

func createDatastoresForTransaction() []txn.Datastorer {
	var datastores []txn.Datastorer

	if redisConn != nil {
		redisDatastore := redis.NewRedisDatastore("Redis", redisConn)
		datastores = append(datastores, redisDatastore)
	}

	return datastores
}

// HTTP Handlers

// healthHandler handles health check requests
func healthHandler(c *fiber.Ctx) error {
	return c.SendString("OK")
}

// redisTestWriteHandler handles Redis write test requests
func redisTestWriteHandler(c *fiber.Ctx) error {
	if redisConn == nil {
		return c.Status(fiber.StatusInternalServerError).
			SendString("Redis connection not initialized")
	}

	var testData TestData
	if err := c.BodyParser(&testData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create distributed transaction
	txn := txn.NewTransactionWithRemote(client, oracle)

	// Add datastores
	datastores := createDatastoresForTransaction()
	txn.AddDatastores(datastores...)

	if err := txn.Start(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to start transaction: %v", err),
		})
	}

	// Write to Redis
	key := fmt.Sprintf("test:data:%s", testData.ID)
	dataJSON, _ := json.Marshal(testData)
	if err := txn.Write("Redis", key, string(dataJSON)); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to write to Redis: %v", err),
		})
	}

	// Commit transaction
	if err := txn.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to commit transaction: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Data written to Redis successfully",
		"data":    testData,
	})
}

// redisTestReadHandler handles Redis read test requests
func redisTestReadHandler(c *fiber.Ctx) error {
	if redisConn == nil {
		return c.Status(fiber.StatusInternalServerError).
			SendString("Redis connection not initialized")
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID is required",
		})
	}

	// Create read-only transaction
	txn := txn.NewTransactionWithRemote(client, oracle)

	// Add datastores
	datastores := createDatastoresForTransaction()
	txn.AddDatastores(datastores...)

	if err := txn.Start(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to start transaction: %v", err),
		})
	}

	// Read from Redis
	key := fmt.Sprintf("test:data:%s", id)
	var dataJSON string
	if err := txn.Read("Redis", key, &dataJSON); err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Data not found",
		})
	}

	// Commit transaction
	if err := txn.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to commit transaction: %v", err),
		})
	}

	var testData TestData
	if err := json.Unmarshal([]byte(dataJSON), &testData); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to parse data: %v", err),
		})
	}

	return c.JSON(testData)
}

// serviceDiscoveryStatusHandler handles service discovery status requests
func serviceDiscoveryStatusHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"type":   *discoveryType,
		"status": "active",
		"config": map[string]interface{}{
			"type":           *discoveryType,
			"etcd_endpoints": config.ServiceDiscovery.Etcd.Endpoints,
			"http_registry":  config.ServiceDiscovery.HTTP.RegistryPort,
		},
	})
}

// redisServiceHandler handles Redis service address requests
func redisServiceHandler(c *fiber.Ctx) error {
	if client == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Service discovery client not initialized",
		})
	}

	redisAddr, err := client.GetServerAddr("Redis")
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":    fmt.Sprintf("Redis service not found: %v", err),
			"fallback": config.RedisAddr,
		})
	}

	return c.JSON(fiber.Map{
		"service":        "Redis",
		"address":        redisAddr,
		"discovery_type": *discoveryType,
	})
}

func main() {
	flag.Parse()

	if err := loadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if err := initConnections(); err != nil {
		log.Fatalf("Error initializing connections: %v", err)
	}

	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New())

	// Routes
	app.Get("/health", healthHandler)

	api := app.Group("/api/v1")
	api.Post("/redis-test", redisTestWriteHandler)
	api.Get("/redis-test/:id", redisTestReadHandler)
	api.Get("/service-discovery/status", serviceDiscoveryStatusHandler)
	api.Get("/services/redis", redisServiceHandler)

	log.Printf("Starting server on %s", config.ServerPort)
	log.Fatal(app.Listen(config.ServerPort))
}
