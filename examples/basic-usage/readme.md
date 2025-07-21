
# Oreo Basic Usage Example

This example demonstrates how to use the Oreo distributed transaction system with Redis and MongoDB datastores. The example shows how to perform cross-database transactions using Oreo's distributed transaction protocol.

## Overview

This basic usage example includes:

- A simple Go application that performs distributed transactions
- Configuration for Redis and MongoDB connections
- Setup instructions for running the complete Oreo system

The example demonstrates:

1. **Cross-database write transaction**: Writing data to both Redis and MongoDB in a single atomic transaction
2. **Cross-database read transaction**: Reading data from both datastores with consistency guarantees

## Prerequisites

- Go 1.19 or higher
- Docker (for running Redis and MongoDB)
- Oreo system components built and available

## Architecture

The example uses the following components:

- **Time Oracle**: Provides global timestamps for transaction ordering
- **Executors**: Handle transaction execution and coordination
- **Client Application**: Your application that performs transactions
- **Datastores**: Redis and MongoDB instances

## Setup Instructions

### 1. Start Database Services

First, start the required database services using Docker:

```shell
# Start Redis
docker run --name redis -p 6379:6379 --restart=always -d redis

# Start MongoDB
docker run -d \
    --name mongodb \
    -p 27017:27017 \
    -e MONGO_INITDB_ROOT_USERNAME=admin \
    -e MONGO_INITDB_ROOT_PASSWORD=password \
    --restart=always \
    mongo
```

### 2. Build Oreo Components

Build the required Oreo system components:

```shell
# Build the fault-tolerant executor
cd ../../ft-executor
go build .

# Build the time oracle
cd ../timeoracle
go build .

# Copy the built binaries to the example directory
cp ft-executor ../examples/basic-usage/
cp timeoracle ../examples/basic-usage/

# Return to the example directory
cd ../examples/basic-usage
```

### 3. Start Oreo System Components

Start the system components in the following order:

```shell
# 1. Start the time oracle (hybrid mode provides both logical and physical timestamps)
./timeoracle -p 8010 -type hybrid

# 2. Start executor instances (in separate terminals)
# Terminal 2:
./ft-executor -p 8001 -w ycsb --advertise-addr "localhost:8001" -bc "./config.yaml" -db "Redis,MongoDB1"

# Terminal 3:
./ft-executor -p 8002 -w ycsb --advertise-addr "localhost:8002" -bc "./config.yaml" -db "Redis,MongoDB1"
```

### 4. Run the Example Application

Once all components are running, execute the example:

```shell
go run .
```

## Expected Output

The application will output something like:

```text
RegistryClient initialized successfully.
Waiting for the connection of executors...
Starting stale instance cleanup routine (TTL: 6s, Interval: 3s)
Registry server starting on :9000 (TTL: 6s)
Heartbeat received from unknown instance: localhost:8001. Instance should re-register.
Registering new instance: localhost:8001 for DsNames: [Redis MongoDB1]
Heartbeat received from unknown instance: localhost:8002. Instance should re-register.
Registering new instance: localhost:8002 for DsNames: [MongoDB1 Redis]
Write transaction committed successfully.
Read values: value1, value2

```

## Configuration

The `config.yaml` file contains the configuration for:

- **Registry address**: Where executors register themselves
- **Time oracle URL**: The time oracle service endpoint
- **Database connections**: Redis and MongoDB connection details

## Code Explanation

The `main.go` file demonstrates:

1. **Client Setup**: Creating a network client to communicate with executors
2. **Datastore Connections**: Establishing connections to Redis and MongoDB
3. **Transaction Creation**: Creating transactions with remote execution capability
4. **Write Transaction**: Writing data to multiple datastores atomically
5. **Read Transaction**: Reading data with consistency guarantees

## Troubleshooting

### Common Issues

1. **Connection refused errors**: Ensure all components are started in the correct order
2. **Database connection failures**: Verify Redis and MongoDB are running and accessible
3. **Executor registration issues**: Check that the registry address matches in all components

### Component Dependencies

Make sure components are started in this order:

1. Database services (Redis, MongoDB)
2. Time Oracle
3. Executors
4. Client application

## Next Steps

- Explore more complex transaction patterns
- Try different datastore combinations
- Experiment with fault-tolerance scenarios
- Review the benchmarking examples for performance testing
