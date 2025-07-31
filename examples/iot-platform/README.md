# IoT Platform

An IoT platform example based on the Oreo distributed transaction framework, demonstrating how to handle IoT device data in a multi-database environment.

# Deployment Options

## Docker Compose Deployment

For quick and easy deployment, use Docker Compose:

```bash
cd examples/iot-platform

# Start all services
docker compose up -d
```

## Manual Deployment

## 1. Start Database Services

```powershell
docker compose -f './docker-compose.yml' up -d cassandra redis mongodb
```

## 2. Configure Cassandra

Connect to the Cassandra container and create keyspace:

```powershell
# Connect to Cassandra container
docker exec -it cassandra cqlsh

# Create keyspace
CREATE KEYSPACE oreo WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
```

## 3. Start Time Oracle Service

```powershell
./ft-timeoracle -role primary -p 8012 -type hybrid -max-skew 50ms
```

## 4. Start Executor Instances

Start in different terminals:

```powershell
# Redis and MongoDB executor
./ft-executor -p 8002 -w ycsb --advertise-addr "localhost:8002" -bc "./executor-config.yaml" -db "Redis,MongoDB1"

# Cassandra executor
./ft-executor -p 8003 -w ycsb --advertise-addr "localhost:8003" -bc "./executor-config.yaml" -db "Cassandra"
```

## 5. Start IoT Platform Service

```powershell
go run main.go
```

The service will start at `http://localhost:8081`.

# API Documentation

## 1. Device Registration API

**Endpoint**: `POST /api/v1/devices`

**Function**: Register a new IoT device to the system

**Request Body Example**:
```json
{
  "device_id": "sensor001",
  "device_name": "Temperature Sensor",
  "device_type": "temperature_sensor",
  "location": "Server Room A",
  "status": "active"
}
```

**Database Operation Flow**:
1. **Redis**: Write device status cache
   - Key: `device:status:{device_id}`
   - Value: Device status (e.g., "active")
   - Purpose: Quick query of device online status

2. **MongoDB**: Store complete device information
   - Key: `device:info:{device_id}`
   - Value: Complete device information JSON string
   - Purpose: Persistent storage of detailed device information

3. **Cassandra**: Initialize device statistics
   - Key: `device:stats:{device_id}`
   - Value: Device statistics (reading count, average, min/max values, etc.)
   - Purpose: Store device historical statistics

---

## 2. Sensor Data Reporting API

**Endpoint**: `POST /api/v1/data`

**Function**: Report IoT device sensor data

**Request Body Example**:
```json
{
  "device_id": "sensor001",
  "sensor_type": "temperature",
  "value": 25.6,
  "unit": "°C",
  "location": "Server Room A"
}
```

**Database Operation Flow**:
1. **Redis**: Update latest device data
   - Key: `device:latest:{device_id}:{sensor_type}`
   - Value: Latest sensor data JSON string
   - Purpose: Real-time query of device latest status

2. **MongoDB**: Store historical data
   - Key: `sensor:history:{device_id}:{timestamp}`
   - Value: Sensor data JSON string
   - Purpose: Historical data query and analysis

3. **Cassandra**: Update statistics
   - Read existing statistics: `device:stats:{device_id}`
   - Calculate new statistics (total count, average, max, min)
   - Write back updated statistics
   - Purpose: Real-time statistical analysis

---

## 3. Get Device Information API

**Endpoint**: `GET /api/v1/devices/{deviceId}`

**Function**: Query detailed information and statistics of a specified device

**Database Read Flow**:
1. **Redis**: Read device status
   - Key: `device:status:{device_id}`
   - Purpose: Verify device existence and get current status

2. **MongoDB**: Read detailed device information
   - Key: `device:info:{device_id}`
   - Purpose: Get complete device registration information

3. **Cassandra**: Read statistics (optional)
   - Key: `device:stats:{device_id}`
   - Purpose: Get device historical statistics

**Response Example**:
```json
{
  "device": {
    "device_id": "sensor001",
    "device_name": "Temperature Sensor",
    "device_type": "temperature_sensor",
    "location": "Server Room A",
    "status": "active",
    "registered_at": "2024-01-01T10:00:00Z"
  },
  "stats": {
    "device_id": "sensor001",
    "total_readings": 1250,
    "last_reading": "2024-01-01T15:30:00Z",
    "avg_value": 24.8,
    "min_value": 18.2,
    "max_value": 31.5
  }
}
```

---

## 4. Get Latest Device Data API

**Endpoint**: `GET /api/v1/devices/{deviceId}/latest?sensor_type={sensorType}`

**Function**: Query the latest data of a specific sensor type for a specified device

**Parameters**:
- `deviceId`: Device ID
- `sensor_type`: Sensor type (required)

**Database Operation**:
1. **Redis**: Read latest data
   - Key: `device:latest:{device_id}:{sensor_type}`
   - Purpose: Quick access to real-time data

**Response Example**:
```json
{
  "device_id": "sensor001",
  "sensor_type": "temperature",
  "value": 25.6,
  "unit": "°C",
  "timestamp": "2024-01-01T15:30:00Z",
  "location": "Server Room A"
}
```

---

## 5. Batch Data Processing API

**Endpoint**: `POST /api/v1/data/batch`

**Function**: Process multiple sensor data in batch

**Request Body Example**:
```json
[
  {
    "device_id": "sensor001",
    "sensor_type": "temperature",
    "value": 25.6,
    "unit": "°C",
    "location": "Server Room A"
  },
  {
    "device_id": "sensor002",
    "sensor_type": "humidity",
    "value": 65.2,
    "unit": "%",
    "location": "Server Room B"
  }
]
```

**Database Operation Flow**:
1. **Redis**: Batch update latest data
   - Update corresponding latest value cache for each sensor data

2. **MongoDB**: Batch store historical data
   - Create historical records for each sensor data

**Response Example**:
```json
{
  "message": "Batch data processed successfully",
  "processed_count": 2,
  "total_count": 2
}
```

---

## 6. Health Check API

**Endpoint**: `GET /health`

**Function**: Check service running status

**Response Example**:
```json
{
  "status": "healthy",
  "time": "2024-01-01T15:30:00Z"
}
```
