# README

> [Deprecated]
>
> Use `ft-timeoracle` instead.

Simple implementation of time oracle

## Overview

This program provides timestamps based on a configurable backend implementation from the `oreo/pkg/timesource` library.

The server listens for HTTP GET requests on the `/timestamp/` endpoint and returns a timestamp as a plain text integer.

## Building

To build the server, use the standard Go build command:

```bash
go build -o timeoracle .
```

## Running

You can run the compiled server using:

```bash
./timeoracle [flags]
```

By default, the server runs on port 8010 and uses the "hybrid" time source implementation.

## Configuration

The server can be configured using command-line flags and an environment variable.

### Command-Line Flags

* `-p <port>`: Specifies the HTTP port number the server should listen on.
  * Type: `integer`
  * Default: `8010`
  * Example: `./timeoracle -p 9000`

* `-type <oracle_type>`: Specifies the type of time source implementation to use.
  * Type: `string`
  * Default: `hybrid`
  * Available options: `hybrid`, `simple`, `counter`
  * Example: `./timeoracle -type simple`

### Environment Variables

* `LOG=<log_level>`: Sets the logging level for the server's output.
  * Type: `string`
  * Default: `ERROR`
  * Available options: `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`
  * Example (Bash/Zsh): `LOG=DEBUG ./timeoracle`

## Usage

Once the server is running, you can request a timestamp by sending an HTTP GET request to the `/timestamp/` endpoint.

You can use tools like `curl` to interact with the server:

**Example (using default port 8010):**

```bash
curl http://localhost:8010/timestamp/
```

**Example (if server started with `-p 9000`):**

```bash
curl http://localhost:9000/timestamp/
```

The server will respond with a numerical timestamp string.
