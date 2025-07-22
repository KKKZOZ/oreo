package mongo

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	"github.com/oreo-dtx-lab/oreo/pkg/txn/testsuite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testMongoDBURI string

type mongoTestSuiteHelper struct{}

func (h *mongoTestSuiteHelper) MakeItem(ops txn.ItemOptions) txn.DataItem {
	return &MongoItem{
		MKey:          ops.Key,
		MValue:        util.ToJSONString(ops.Value),
		MGroupKeyList: ops.GroupKeyList,
		MTxnState:     ops.TxnState,
		MTValid:       ops.TValid,
		MTLease:       ops.TLease,
		MPrev:         ops.Prev,
		MIsDeleted:    ops.IsDeleted,
		MVersion:      ops.Version,
	}
}

func (h *mongoTestSuiteHelper) NewInstance() txn.DataItem {
	return &MongoItem{}
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mongo:7.0",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForListeningPort("27017/tcp"),
	}

	mongoContainer, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	if err != nil {
		log.Fatalf("Could not start MongoDB container: %s", err)
	}

	defer func() {
		if err := mongoContainer.Terminate(ctx); err != nil {
			log.Fatalf("Could not stop MongoDB container: %s", err)
		}
	}()

	mappedPort, _ := mongoContainer.MappedPort(ctx, "27017")
	host, _ := mongoContainer.Host(ctx)
	testMongoDBURI = fmt.Sprintf("mongodb://%s:%s", host, mappedPort.Port())

	exitCode := m.Run()
	os.Exit(exitCode)
}

func newTestMongoDBConnection() *MongoConnection {
	connectionOptions := &ConnectionOptions{
		Address: testMongoDBURI, DBName: "oreo",
		CollectionName: "records",
	}
	connection := NewMongoConnection(connectionOptions)
	if err := connection.Connect(); err != nil {
		log.Fatalf("Could not connect to MongoDB: %s", err)
	}
	return connection
}

func TestMongoDBConnector_InterfaceSuite(t *testing.T) {
	testsuite.TestConnectorSuite(t, newTestMongoDBConnection(), &mongoTestSuiteHelper{})
}
