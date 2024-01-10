package main

import (
	"fmt"
	"time"

	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	red "github.com/kkkzoz/oreo/pkg/datastore/redis"
	. "github.com/kkkzoz/oreo/pkg/logger"
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
	expectedItem := red.RedisItem{
		Key:       key,
		Value:     util.ToJSONString(expectedValue),
		TxnId:     "1",
		TxnState:  config.COMMITTED,
		TValid:    time.Now().Add(-3 * time.Second),
		TLease:    time.Now().Add(-2 * time.Second),
		Prev:      "",
		IsDeleted: false,
		Version:   2,
	}
	err := conn.PutItem(key, expectedItem)
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
	Log.Infow("Get success", "item.Value", item.Value)
	// testutil.Log.Infow("failed to fetch URL",
	// 	// Structured context as loosely typed key-value pairs.
	// 	"url", 12313123,
	// 	"attempt", 3,
	// 	"backoff", time.Second,
	// )
	// log.Infow("Get item", "item", item)
}
