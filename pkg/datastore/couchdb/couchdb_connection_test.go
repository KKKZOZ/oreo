package couchdb

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	"github.com/oreo-dtx-lab/oreo/pkg/txn/testsuite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testCouchDBURI string

type couchDBTestSuiteHelper struct{}

func (h *couchDBTestSuiteHelper) MakeItem(ops txn.ItemOptions) txn.DataItem {
	return &CouchDBItem{
		CKey:          ops.Key,
		CValue:        util.ToJSONString(ops.Value),
		CGroupKeyList: ops.GroupKeyList,
		CTxnState:     ops.TxnState,
		CTValid:       ops.TValid,
		CTLease:       ops.TLease,
		CPrev:         ops.Prev,
		CIsDeleted:    ops.IsDeleted,
		CVersion:      ops.Version,
	}
}

func (h *couchDBTestSuiteHelper) NewInstance() txn.DataItem {
	return &CouchDBItem{}
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	const (
		user = "admin"
		pass = "password"
	)

	req := testcontainers.ContainerRequest{
		Image:        "couchdb:3.3",        // 官方镜像中体积较小的 3.x 版 (~150 MB)
		ExposedPorts: []string{"5984/tcp"}, // CouchDB 默认端口
		Env: map[string]string{ // ✨ 关键：必须显式设置
			"COUCHDB_USER":     user,
			"COUCHDB_PASSWORD": pass,
		},
		// CouchDB 在端口开放后才算 ready，可直接用健康检查端点 /_up
		WaitingFor: wait.ForHTTP("/_up").
			WithPort("5984/tcp").
			WithStartupTimeout(30 * time.Second),
	}

	couch, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Could not start CouchDB container: %v", err)
	}
	defer func() {
		if err := couch.Terminate(ctx); err != nil {
			log.Fatalf("Could not stop CouchDB container: %v", err)
		}
	}()

	host, _ := couch.Host(ctx)
	port, _ := couch.MappedPort(ctx, "5984")
	testCouchDBURI = fmt.Sprintf("http://%s:%s@%s:%s/", user, pass, host, port.Port())

	// 运行其他测试
	code := m.Run()
	os.Exit(code)
}

func newTestCouchDBConnection() *CouchDBConnection {
	connectionOptions := &ConnectionOptions{Address: testCouchDBURI, DBName: "oreo"}
	connection := NewCouchDBConnection(connectionOptions)
	if err := connection.Connect(); err != nil {
		log.Fatalf("Could not connect to CouchDB: %s", err)
	}
	return connection
}

func TestCouchDBConnector_InterfaceSuite(t *testing.T) {
	testsuite.TestConnectorSuite(t, newTestCouchDBConnection(), &couchDBTestSuiteHelper{})
}
