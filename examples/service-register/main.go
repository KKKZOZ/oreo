package main

import (
	"encoding/json"
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
		Type string `yaml:"type"` // "http" or "etcd"
		HTTP struct {
			RegistryAddr      string `yaml:"registry_addr"`
			RegistryPort      string `yaml:"registry_port"`
			RegistryServerURL string `yaml:"registry_server_url"`
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
	client    *network.Client
	oracle    timesource.TimeSourcer
	redisConn *redis.RedisConnection
	config    Config
)

func loadConfig() error {
	data, err := os.ReadFile("./config.yaml")
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

	log.Printf("Initializing service discovery with type: %s", config.ServiceDiscovery.Type)

	// Create service discovery configuration
	var discoveryConfig *discovery.ServiceDiscoveryConfig

	switch config.ServiceDiscovery.Type {
	case "http":
		log.Println("Using HTTP service discovery")
		if config.ServiceDiscovery.HTTP.RegistryServerURL == "" {
			// Act as registry server
			log.Printf(
				"Starting as HTTP registry server on port %s",
				config.ServiceDiscovery.HTTP.RegistryPort,
			)
			discoveryConfig = &discovery.ServiceDiscoveryConfig{
				Type: discovery.HTTPDiscovery,
				HTTP: &discovery.HTTPDiscoveryConfig{
					RegistryPort: config.ServiceDiscovery.HTTP.RegistryPort,
					// No RegistryServerURL means we ARE the registry server
				},
			}
		} else {
			// Connect to external registry server
			log.Printf("Connecting to external HTTP registry server: %s", config.ServiceDiscovery.HTTP.RegistryServerURL)
			discoveryConfig = &discovery.ServiceDiscoveryConfig{
				Type: discovery.HTTPDiscovery,
				HTTP: &discovery.HTTPDiscoveryConfig{
					RegistryServerURL: config.ServiceDiscovery.HTTP.RegistryServerURL,
				},
			}
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
			config.ServiceDiscovery.Type,
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

func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if err := initConnections(); err != nil {
		log.Fatalf("Error initializing connections: %v", err)
	}

	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	api := app.Group("/api/v1")

	// Redis test endpoint
	api.Post("/redis-test", func(c *fiber.Ctx) error {
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

		oracle := timesource.NewGlobalTimeSource("http://localhost:8012")
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
	})

	// Redis read endpoint
	api.Get("/redis-test/:id", func(c *fiber.Ctx) error {
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
	})

	// In main function's API routes section
	api.Get("/service-discovery/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"type":   config.ServiceDiscovery.Type,
			"status": "active",
			"config": map[string]interface{}{
				"type":           config.ServiceDiscovery.Type,
				"etcd_endpoints": config.ServiceDiscovery.Etcd.Endpoints,
				"http_registry":  config.ServiceDiscovery.HTTP.RegistryAddr,
			},
		})
	})

	api.Get("/services/redis", func(c *fiber.Ctx) error {
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
			"discovery_type": config.ServiceDiscovery.Type,
		})
	})

	log.Printf("Starting server on %s", config.ServerPort)
	log.Fatal(app.Listen(config.ServerPort))
}
