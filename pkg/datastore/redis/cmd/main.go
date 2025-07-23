package main

import (
	"fmt"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/testutil"
	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	red "github.com/oreo-dtx-lab/oreo/pkg/datastore/redis"
	"github.com/oreo-dtx-lab/oreo/pkg/logger"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

type User struct {
	ID   int
	Name string
}

func main() {
	conn := red.NewRedisConnection(&red.ConnectionOptions{
		Address:  "localhost:6666",
		Password: "",
	})
	key := "test_key"
	expectedValue := testutil.NewDefaultPerson()
	expectedItem := red.NewRedisItem(txn.ItemOptions{
		Key:          key,
		Value:        util.ToJSONString(expectedValue),
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       time.Now().Add(-3 * time.Second).UnixMicro(),
		TLease:       time.Now().Add(-2 * time.Second),
		Prev:         "",
		IsDeleted:    false,
		Version:      "2",
	})
	_, err := conn.PutItem(key, expectedItem)
	if err != nil {
		fmt.Println(err)
	}
	// testutil.Log.Info("Put success")
	testutil.Debug(testutil.DTest, "Put success of Debug")
	// testutil.Log.Debugw("Put success of Debugf", "Topic", testutil.DTest)

	item, err := conn.GetItem("test_key")
	if err != nil {
		fmt.Println(err)
		return
	}
	logger.Infow("Get success", "item.Value", item.Value())
	// testutil.Log.Infow("failed to fetch URL",
	// 	// Structured context as loosely typed key-value pairs.
	// 	"url", 12313123,
	// 	"attempt", 3,
	// 	"backoff", time.Second,
	// )
	// log.Infow("Get item", "item", item)
}
