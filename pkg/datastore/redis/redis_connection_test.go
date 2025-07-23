package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/kkkzoz/oreo/pkg/txn/testsuite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testRedisURI string

type redisTestSuiteHelper struct{}

func (h *redisTestSuiteHelper) MakeItem(ops txn.ItemOptions) txn.DataItem {
	return &RedisItem{
		RKey:          ops.Key,
		RValue:        util.ToJSONString(ops.Value),
		RGroupKeyList: ops.GroupKeyList,
		RTxnState:     ops.TxnState,
		RTValid:       ops.TValid,
		RTLease:       ops.TLease,
		RPrev:         ops.Prev,
		RIsDeleted:    ops.IsDeleted,
		RVersion:      ops.Version,
	}
}

func (h *redisTestSuiteHelper) NewInstance() txn.DataItem {
	return &RedisItem{}
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}
	redisContainer, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	if err != nil {
		log.Fatalf("Could not start redis container: %s", err)
	}

	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			log.Fatalf("Could not stop redis container: %s", err)
		}
	}()

	mappedPort, _ := redisContainer.MappedPort(ctx, "6379")
	host, _ := redisContainer.Host(ctx)
	testRedisURI = fmt.Sprintf("%s:%s", host, mappedPort.Port())

	exitCode := m.Run()
	os.Exit(exitCode)
}

func newTestRedisConnection() *RedisConnection {
	connectionOptions := &ConnectionOptions{Address: testRedisURI}
	connection := NewRedisConnection(connectionOptions)
	if err := connection.Connect(); err != nil {
		log.Fatalf("Could not connect to Redis: %s", err)
	}
	return connection
}

func TestRedisConnector_InterfaceSuite(t *testing.T) {
	testsuite.TestConnectorSuite(t, newTestRedisConnection(), &redisTestSuiteHelper{})
}
