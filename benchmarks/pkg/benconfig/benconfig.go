package benconfig

import (
	"strings"
	"time"

	"github.com/kkkzoz/oreo/pkg/network"
)

var (
	ExecutorAddressMap                  = map[string][]string{"ALL": {"localhost:8000"}}
	TimeOracleUrl                       = "http://localhost:8010"
	ZipfianConstant                     = 0.9
	GlobalLatency                       = 3 * time.Millisecond
	MaxLoadBatchSize                    = 100
	RegistryAddr                        = "localhost:9000"
	GlobalIsFaultTolerance              = false
	GlobalFaultToleranceRequestInterval = 0 * time.Millisecond
)

var GlobalClient *network.Client

type BenchmarkConfig struct {
	RegistryAddr       string              `yaml:"registry_addr"`
	RegistryAddrs      []string            `yaml:"registry_addrs"`
	ExecutorAddressMap map[string][]string `yaml:"executor_address_map"`
	TimeOracleUrl      string              `yaml:"time_oracle_url"`
	ZipfianConstant    float64             `yaml:"zipfian_constant"`
	Latency            time.Duration       `yaml:"latency"`
	LatencyValue       int                 `yaml:"latency_value"`
	MaxLoadBatchSize   int                 `yaml:"max_load_batch_size"`

	FaultToleranceRequestInterval time.Duration `yaml:"fault_tolerance_request_interval"`

	FaultToleranceRequestIntervalValue int `yaml:"fault_tolerance_request_interval_value"`

	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`

	MongoDBAddr1    string `yaml:"mongodb_addr1"`
	MongoDBAddr2    string `yaml:"mongodb_addr2"`
	MongoDBUsername string `yaml:"mongodb_username"`
	MongoDBPassword string `yaml:"mongodb_password"`

	KVRocksAddr     string `yaml:"kvrocks_addr"`
	KVRocksPassword string `yaml:"kvrocks_password"`

	CouchDBAddr     string `yaml:"couchdb_addr"`
	CouchDBUsername string `yaml:"couchdb_username"`
	CouchDBPassword string `yaml:"couchdb_password"`

	CassandraAddr     []string `yaml:"cassandra_addr"`
	CassandraUsername string   `yaml:"cassandra_username"`
	CassandraPassword string   `yaml:"cassandra_password"`
	DynamoDBAddr      string   `yaml:"dynamodb_addr"`
	TiKVAddr          []string `yaml:"tikv_addr"`

	// DBCombination []string `yaml:"db_combination"`
}

// ResolveHTTPRegistryAddrs returns the list of configured HTTP registry
// addresses, normalizing legacy single-address fields and trimming blanks.
// It prefers the explicit slice when present to keep backwards compatibility
// with existing configs using registry_addr.
func (cfg *BenchmarkConfig) ResolveHTTPRegistryAddrs() []string {
	var out []string

	appendIfNotEmpty := func(addr string) {
		addr = strings.TrimSpace(addr)
		if addr != "" {
			out = append(out, addr)
		}
	}

	for _, addr := range cfg.RegistryAddrs {
		appendIfNotEmpty(addr)
	}

	if len(out) > 0 {
		return out
	}

	if cfg.RegistryAddr != "" {
		for _, addr := range strings.Split(cfg.RegistryAddr, ",") {
			appendIfNotEmpty(addr)
		}
	}

	return out
}
