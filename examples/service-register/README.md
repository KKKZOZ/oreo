# Service Register Example

This is a sample project demonstrating how to use the Oreo distributed transaction framework for service registration and discovery.

## Startup Instructions

### 1. Start the Executor

When starting the executor, use the `-register` flag to specify the service discovery type:

```bash
# Use HTTP-based service discovery
./ft-executor -register http

# Use etcd-based service discovery
./ft-executor -register etcd
```

### 2. Configuration File

The project requires a `config.yaml` configuration file, which must include the `service_discovery` section:

```yaml
service_discovery:
  type: "etcd"  # "http" or "etcd"
  
  # HTTP service discovery configuration
  http:
    registry_port: ":9000"
  
  # etcd service discovery configuration
  etcd:
    endpoints:
      - "localhost:2379"
    username: ""
    password: ""
    dial_timeout: "5s"
    key_prefix: "/oreo/services"
```

### 3. Full Configuration Example

```yaml
# Time oracle configuration
time_oracle_url: "http://localhost:8012/timestamp/common"
server_port: ":3001"

# Redis configuration
redis_addr: "localhost:6379"
redis_password: "kkkzoz"

# MongoDB configuration
mongodb_addr: "mongodb://localhost:27017"
mongodb_username: "admin"
mongodb_password: "password"
mongodb_db_name: "test_db"
mongodb_collection_name: "test_data"

# Service discovery configuration
service_discovery:
  type: "etcd"  # "http" or "etcd"
  
  http:
    registry_port: ":9000"
  
  etcd:
    endpoints:
      - "localhost:2379"
    username: ""
    password: ""
    dial_timeout: "5s"
    key_prefix: "/oreo/services"
```

## API Endpoints

### Health Check

- **GET** `/health` – Service health check

### Redis Test Endpoints

#### Write Data to Redis

- **POST** `/api/v1/redis-test` – Write test data to Redis

**Request body:**
```json
{
  "id": "test-id",
  "message": "test message",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

**Response:**
```json
{
  "message": "Data written to Redis successfully",
  "data": {
    "id": "test-id",
    "message": "test message",
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

#### Read Data from Redis

- **GET** `/api/v1/redis-test/:id` – Read test data from Redis

**Response:**
```json
{
  "id": "test-id",
  "message": "test message",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### MongoDB Test Endpoints

#### Write Data to MongoDB

- **POST** `/api/v1/mongo-test` – Write test data to MongoDB

**Request body:**
```json
{
  "id": "test-id",
  "message": "test message"
}
```

**Response:**
```json
{
  "message": "Data written to MongoDB successfully",
  "data": {
    "id": "test-id",
    "message": "test message",
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

### Service Discovery Status Endpoints

#### Get Service Discovery Status

- **GET** `/api/v1/service-discovery/status` – View current service discovery status

**Response:**
```json
{
  "type": "etcd",
  "status": "active",
  "config": {
    "type": "etcd",
    "etcd_endpoints": ["localhost:2379"],
    "http_registry": ""
  }
}
```

#### Get Redis Service Address

- **GET** `/api/v1/services/redis` – View Redis service address

**Response:**
```json
{
  "service": "Redis",
  "address": "localhost:6379",
  "discovery_type": "etcd"
}
```

## Startup Steps

1. Start the time oracle service
2. Start the executor: `./ft-executor -register etcd`
3. Start the service-register service: `go run main.go`
4. Use the API endpoints to test

## Notes

- The service discovery type must match the `-register` parameter used when starting the executor
- etcd-based service discovery requires a running etcd cluster
- HTTP-based service discovery will start an internal HTTP registry
- If service discovery fails, the system will fall back to static configuration