package main

import (
	"fmt"
	"time"

	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	red "github.com/kkkzoz/oreo/pkg/datastore/redis"
)

type User struct {
	ID   int
	Name string
}

func main() {
	conn := red.NewRedisConnection(&red.ConnectionOptions{
		Address:  "redis-17297.c294.ap-northeast-1-2.ec2.cloud.redislabs.com:17297",
		Password: "57xr6sTv8FLL0QCJtPExr3ULWpoSj5Z6",
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
	fmt.Println("Put success")

	item, err := conn.GetItem("123123")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Get success")
	fmt.Println(item)
}
