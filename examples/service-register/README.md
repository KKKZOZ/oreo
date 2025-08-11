# Service Register Example

This example demonstrates how to build a client application that uses Oreo's service discovery feature to connect with and use a distributed transactional system.

The application starts an API server that can perform transactional reads and writes to a Redis database. It discovers the necessary transaction-coordinating nodes (`executor`) via a service discovery mechanism (either `etcd` or HTTP).

## Architecture

1. **API Server (`main.go`)**: A web server built with Fiber that exposes endpoints to read/write data.
2. **Oreo Client (`network.NewClient`)**: The client component responsible for communicating with the Oreo transaction executors.
3. **Service Discovery**: The client is configured to find executor nodes using a discovery service. This example supports:
    * **etcd**: Executors register themselves with an `etcd` cluster. The client queries `etcd` to find them.
    * **HTTP**: Executors register with a central HTTP registry. The client queries this registry.
4. **Oreo Executor Nodes**: Separate processes (e.g., `ft-executor`) that handle the actual transaction logic. This example application acts as a *client* to these nodes.
5. **Time Oracle**: A central service (`ft-timeoracle`) that provides timestamps for transactions.

## Prerequisites

* Go 1.18+
* A running **Time Oracle** instance.
* Running **Executor** instances that have registered themselves using a service discovery method.
* A running **Redis** instance.
* (Optional) A running **etcd** cluster if you are using `etcd` for service discovery.

## Deploying Dependencies (Redis & etcd)

You can easily deploy the required services using Docker.

### Redis

Run the following command to start a Redis container:

```sh
docker run --name oreo-redis -p 6379:6379 -d redis:7.2-alpine
```

### etcd

If you are using `etcd` for service discovery, run this command to start an etcd container:

```sh
docker run -d \
    --name oreo-etcd \
    -p 2379:2379 \
    -p 2380:2380 \
    --env ALLOW_NONE_AUTHENTICATION=yes \
    --env ETCD_ADVERTISE_CLIENT_URLS=http://localhost:2379 \
    --env ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379 \
    gcr.io/etcd-development/etcd:v3.5.13
```

This starts a single-node etcd cluster that is accessible at `localhost:2379` without authentication.

## Configuration (`config.yaml`)

The application is configured using `config.yaml`. Key sections are:

* `time_oracle_url`: The URL of the running time oracle.
* `server_port`: The port on which this example's API server will run.
* `redis_addr`: The address of the Redis instance.
* `service_discovery`: Configures the discovery method.
  * `http`: Configuration for the HTTP discovery.
  * `etcd`: Configuration for the `etcd` discovery.

## How to Run

1. **Build Binaries**: First, compile the `ft-timeoracle` and `ft-executor` services using the provided script. The script will place the compiled binaries in the current directory.

    ```sh
    ./build_services.sh
    ```

2. **Start Oreo Services**: Before running this example, ensure the prerequisite Oreo services are running.

    * **Start Time Oracle**:

        ```sh
        ./ft-timeoracle -role primary -p 8012 -type hybrid -max-skew 50ms
        ```

    * **Start at least one Executor**

        ```sh
        # Using etcd for service discovery
        ./ft-executor -p 8001 --advertise-addr "localhost:8001" -bc "./executor-etcd-config.yaml" -w ycsb -db "Redis" -registry etcd

        # OR

        # Using HTTP for service discovery
        ./ft-executor -p 8001 --advertise-addr "localhost:8001" -bc "./executor-http-config.yaml" -w ycsb -db "Redis" -registry http
        ```

3. **Configure the Example**: Edit `config.yaml` in this directory (`examples/service-register`) to match your environment (e.g., `etcd` endpoints, Redis address).

4. **Run the Example Application**:
    Navigate to this directory and run the main program.

    ```sh
    go run main.go --discovery etcd

    # OR
    go run main.go --discovery http
    ```

    The API server will start on the port specified in `config.yaml` (e.g., `:3001`).

## API Endpoints

You can now interact with the running API server.

* **`GET /health`**: Health check for the API server.
* **`GET /api/v1/service-discovery/status`**: Shows the configured service discovery method.
* **`GET /api/v1/services/redis`**: Attempts to discover the address of a registered "Redis" service via the discovery mechanism.

### Redis Transactions

* **`POST /api/v1/redis-test`**: Writes data to Redis within an Oreo transaction.

  **Request Body**:

  ```json
  {
    "id": "my-key-1",
    "message": "Hello, distributed world!",
    "timestamp": "2024-01-01T00:00:00Z"
  }
  ```

* **`GET /api/v1/redis-test/:id`**: Reads data from Redis within an Oreo transaction.
  * Example: `GET http://localhost:3001/api/v1/redis-test/my-key-1`

### Testing the API

A helper script, `test_api.sh`, is provided to test all endpoints. Make sure `httpie` is installed, then run the script:

```sh
./test_api.sh
```
