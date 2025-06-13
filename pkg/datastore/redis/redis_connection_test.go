package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/testutil"
	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/serializer"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testRedisURI string

func TestMain(m *testing.M) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
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
	connectionOptions := &ConnectionOptions{
		Address: testRedisURI,
	}
	connection := NewRedisConnection(connectionOptions)
	err := connection.Connect()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %s", err)
	}
	return connection
}

func TestNewRedisConnection_DefaultNilArgument(t *testing.T) {
	fmt.Println("testRedisURI:", testRedisURI)
	connection := NewRedisConnection(nil)
	assert.NotNil(t, connection)
	assert.Equal(t, "localhost:6379", connection.Address)
}

func TestNewRedisConnection_DefaultAddress(t *testing.T) {
	connectionOptions := &ConnectionOptions{}
	connection := NewRedisConnection(connectionOptions)
	assert.NotNil(t, connection)
	assert.Equal(t, "localhost:6379", connection.Address)
}

func TestNewRedisConnection_WithAddress(t *testing.T) {
	expectedAddress := testRedisURI
	connectionOptions := &ConnectionOptions{Address: expectedAddress}
	connection := NewRedisConnection(connectionOptions)
	assert.NotNil(t, connection)
	assert.Equal(t, expectedAddress, connection.Address)
}

func TestRedisConnection_Connect(t *testing.T) {
	conn := newTestRedisConnection()

	err := conn.Connect()
	assert.Nil(t, err)
}

func TestTimestamp(t *testing.T) {
	tValid := time.Now().Add(-3 * time.Second)
	// tValidStr :=
	t1, _ := time.Parse(time.RFC3339Nano, tValid.Format(time.RFC3339Nano))
	// assert.Equal(t, tValid, t1)
	if !t1.Equal(tValid) {
		t.Error("Not Equal")
	}
	// if t1 != tValid {
	// 	t.Error("Not Equal")
	// }
}

func TestRedisConnection_GetItemNotFound(t *testing.T) {
	conn := newTestRedisConnection()

	key := "test_key"

	_, err := conn.GetItem(key)

	assert.EqualError(t, err, txn.KeyNotFound.Error())
}

func TestRedisConnectionPutItemAndGetItem(t *testing.T) {
	conn := newTestRedisConnection()
	conn.Delete("test_key")

	key := "test_key"
	expectedValue := testutil.NewDefaultPerson()
	expectedItem := &RedisItem{
		RKey:          key,
		RValue:        util.ToJSONString(expectedValue),
		RGroupKeyList: "1",
		RTxnState:     config.COMMITTED,
		RTValid:       -1,
		RTLease:       time.Now().Add(-2 * time.Second),
		RPrev:         "",
		RIsDeleted:    false,
		RVersion:      "2",
	}

	_, err := conn.PutItem(key, expectedItem)
	assert.NoError(t, err)

	item, err := conn.GetItem(key)

	assert.NoError(t, err)
	if !item.Equal(expectedItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", expectedItem, item)
	}
}

func TestRedisConnectionReplaceAndGetItem(t *testing.T) {
	conn := newTestRedisConnection()

	key := "test_key"
	olderPerson := testutil.NewDefaultPerson()
	olderItem := &RedisItem{
		RKey:          key,
		RValue:        util.ToJSONString(olderPerson),
		RGroupKeyList: "1",
		RTxnState:     config.COMMITTED,
		RTValid:       -3,
		RTLease:       time.Now().Add(-2 * time.Second),
		RPrev:         "",
		RIsDeleted:    false,
		RVersion:      "2",
	}

	_, err := conn.PutItem(key, olderItem)
	assert.NoError(t, err)

	newerPerson := testutil.NewDefaultPerson()
	newerPerson.Name = "newer"
	newerItem := &RedisItem{
		RKey:          key,
		RValue:        util.ToJSONString(newerPerson),
		RGroupKeyList: "2",
		RTxnState:     config.COMMITTED,
		RTValid:       -1,
		RTLease:       time.Now().Add(1 * time.Second),
		RPrev:         util.ToJSONString(olderItem),
		RIsDeleted:    false,
		RVersion:      "3",
	}

	_, err = conn.PutItem(key, newerItem)
	assert.NoError(t, err)

	item, err := conn.GetItem(key)
	assert.NoError(t, err)
	if !item.Equal(newerItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", newerItem, item)
	}
}

func TestRedisConnectionConditionalUpdateSuccess(t *testing.T) {
	conn := newTestRedisConnection()

	key := "test_key"
	olderPerson := testutil.NewDefaultPerson()
	olderItem := &RedisItem{
		RKey:          key,
		RValue:        util.ToJSONString(olderPerson),
		RGroupKeyList: "1",
		RTxnState:     config.COMMITTED,
		RTValid:       -3,
		RTLease:       time.Now().Add(-2 * time.Second),
		RPrev:         "",
		RIsDeleted:    false,
		RVersion:      "2",
	}
	_, err := conn.PutItem(key, olderItem)
	assert.NoError(t, err)

	newerPerson := testutil.NewDefaultPerson()
	newerPerson.Name = "newer"
	newerItem := &RedisItem{
		RKey:          key,
		RValue:        util.ToJSONString(newerPerson),
		RGroupKeyList: "2",
		RTxnState:     config.COMMITTED,
		RTValid:       -2,
		RTLease:       time.Now().Add(-1 * time.Second),
		RPrev:         "",
		RIsDeleted:    false,
		RVersion:      "2",
	}

	_, err = conn.ConditionalUpdate(key, newerItem, false)
	assert.NoError(t, err)

	item, err := conn.GetItem(key)
	assert.NoError(t, err)

	newerItem.RVersion = util.AddToString(newerItem.RVersion, 1)

	if !item.Equal(newerItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", newerItem, item)
	}

}

func TestRedisConnectionConditionalUpdateFail(t *testing.T) {
	conn := newTestRedisConnection()
	key := "test_key"
	olderPerson := testutil.NewDefaultPerson()
	olderItem := &RedisItem{
		RKey:          key,
		RValue:        util.ToJSONString(olderPerson),
		RGroupKeyList: "1",
		RTxnState:     config.COMMITTED,
		RTValid:       -3,
		RTLease:       time.Now().Add(-2 * time.Second),
		RPrev:         "",
		RIsDeleted:    false,
		RVersion:      "2",
	}
	_, err := conn.PutItem(key, olderItem)
	assert.NoError(t, err)

	newerPerson := testutil.NewDefaultPerson()
	newerPerson.Name = "newer"
	newerItem := &RedisItem{
		RKey:          key,
		RValue:        util.ToJSONString(olderPerson),
		RGroupKeyList: "2",
		RTxnState:     config.COMMITTED,
		RTValid:       -2,
		RTLease:       time.Now().Add(-1 * time.Second),
		RPrev:         "",
		RIsDeleted:    false,
		RVersion:      "3",
	}

	_, err = conn.ConditionalUpdate(key, newerItem, false)
	assert.EqualError(t, err, "version mismatch")

	item, err := conn.GetItem(key)
	assert.NoError(t, err)
	if !item.Equal(olderItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", olderItem, item)
	}
}

func TestRedisConnectionConditionalUpdateNonExist(t *testing.T) {
	conn := newTestRedisConnection()

	key := "test_key"
	conn.Delete(key)
	newerPerson := testutil.NewDefaultPerson()
	newerPerson.Name = "newer"
	newerItem := &RedisItem{
		RKey:          key,
		RValue:        util.ToJSONString(newerPerson),
		RGroupKeyList: "2",
		RTxnState:     config.COMMITTED,
		RTValid:       -2,
		RTLease:       time.Now().Add(-1 * time.Second),
		RPrev:         "",
		RIsDeleted:    false,
		RVersion:      "1",
	}

	_, err := conn.ConditionalUpdate(key, newerItem, true)
	assert.NoError(t, err)

	item, err := conn.GetItem(key)
	assert.NoError(t, err)

	newerItem.RVersion = util.AddToString(newerItem.RVersion, 1)
	if !item.Equal(newerItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", newerItem, item)
	}
}

func TestRedisConnectionConditionalUpdateConcurrently(t *testing.T) {

	t.Run("this is a update", func(t *testing.T) {
		conn := newTestRedisConnection()
		conn.Connect()

		key := "test_key"
		olderPerson := testutil.NewDefaultPerson()
		olderItem := &RedisItem{
			RKey:          key,
			RValue:        util.ToJSONString(olderPerson),
			RGroupKeyList: "1",
			RTxnState:     config.COMMITTED,
			RTValid:       -3,
			RTLease:       time.Now().Add(-2 * time.Second),
			RPrev:         "",
			RIsDeleted:    false,
			RVersion:      "2",
		}
		_, err := conn.PutItem(key, olderItem)
		assert.NoError(t, err)

		resChan := make(chan bool)
		currentNum := 100
		globalId := 0
		for i := 1; i <= currentNum; i++ {
			go func(id int) {
				newerPerson := testutil.NewDefaultPerson()
				newerPerson.Name = "newer"
				newerItem := &RedisItem{
					RKey:          key,
					RValue:        util.ToJSONString(newerPerson),
					RGroupKeyList: strconv.Itoa(id),
					RTxnState:     config.COMMITTED,
					RTValid:       -2,
					RTLease:       time.Now().Add(-1 * time.Second),
					RPrev:         "",
					RIsDeleted:    false,
					RVersion:      "2",
				}

				_, err = conn.ConditionalUpdate(key, newerItem, false)
				if err == nil {
					globalId = id
					resChan <- true
				} else {
					fmt.Printf("error: %v\n", err)
					resChan <- false
				}
			}(i)
		}
		successCnt := 0
		for i := 1; i <= currentNum; i++ {
			res := <-resChan
			if res {
				successCnt++
			}
		}
		assert.Equal(t, 1, successCnt)

		dataItem, err := conn.GetItem(key)
		assert.NoError(t, err)
		item := dataItem.(*RedisItem)
		if item.GroupKeyList() != strconv.Itoa(globalId) {
			t.Errorf("\nexpect: \n%v, \nactual: \n%v", globalId, item.GroupKeyList())
		}
	})

	t.Run("this is a create", func(t *testing.T) {
		conn := newTestRedisConnection()
		conn.Connect()
		key := "test_key"
		conn.Delete(key)

		resChan := make(chan bool)
		currentNum := 100
		globalId := 0
		for i := 1; i <= currentNum; i++ {
			go func(id int) {
				newerPerson := testutil.NewDefaultPerson()
				newerPerson.Name = "newer"
				newerItem := &RedisItem{
					RKey:          key,
					RValue:        util.ToJSONString(newerPerson),
					RGroupKeyList: strconv.Itoa(id),
					RTxnState:     config.COMMITTED,
					RTValid:       -2,
					RTLease:       time.Now().Add(-1 * time.Second),
					RPrev:         "",
					RIsDeleted:    false,
					RVersion:      "2",
				}

				_, err := conn.ConditionalUpdate(key, newerItem, true)
				if err == nil {
					globalId = id
					resChan <- true
				} else {
					resChan <- false
				}
			}(i)
		}
		successCnt := 0
		for i := 1; i <= currentNum; i++ {
			res := <-resChan
			if res {
				successCnt++
			}
		}
		assert.Equal(t, 1, successCnt)

		item, err := conn.GetItem(key)
		assert.NoError(t, err)
		if item.(*RedisItem).GroupKeyList() != strconv.Itoa(globalId) {
			t.Errorf("\nexpect: \n%v, \nactual: \n%v", globalId, item.(*RedisItem).GroupKeyList())
		}
	})
}

func TestRedisConnectionPutAndGet(t *testing.T) {
	conn := newTestRedisConnection()
	se := serializer.NewJSONSerializer()

	key := "test_key"
	person := testutil.NewDefaultPerson()
	item := &RedisItem{
		RKey:          key,
		RValue:        util.ToJSONString(person),
		RGroupKeyList: "1",
		RTxnState:     config.COMMITTED,
		RTValid:       -3,
		RTLease:       time.Now().Add(-2 * time.Second),
		RPrev:         "",
		RIsDeleted:    false,
		RVersion:      "2",
	}
	bs, err := se.Serialize(item)
	assert.NoError(t, err)
	err = conn.Put(key, bs)
	assert.NoError(t, err)

	str, err := conn.Get(key)
	assert.NoError(t, err)
	var actualItem RedisItem
	err = se.Deserialize([]byte(str), &actualItem)
	assert.NoError(t, err)
	if !actualItem.Equal(item) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", item, actualItem)
	}
}

func TestRedisConnectionReplaceAndGet(t *testing.T) {
	conn := newTestRedisConnection()
	se := serializer.NewJSONSerializer()

	key := "test_key"
	person := testutil.NewDefaultPerson()
	item := &RedisItem{
		RKey:          key,
		RValue:        util.ToJSONString(person),
		RGroupKeyList: "1",
		RTxnState:     config.COMMITTED,
		RTValid:       -3,
		RTLease:       time.Now().Add(-2 * time.Second),
		RPrev:         "",
		RIsDeleted:    false,
		RVersion:      "2",
	}
	bs, err := se.Serialize(item)
	assert.NoError(t, err)
	err = conn.Put(key, bs)
	assert.NoError(t, err)

	item.RVersion = util.AddToString(item.RVersion, 1)
	bs, _ = se.Serialize(item)
	err = conn.Put(key, bs)
	assert.NoError(t, err)

	str, err := conn.Get(key)
	assert.NoError(t, err)
	var actualItem RedisItem
	err = se.Deserialize([]byte(str), &actualItem)
	assert.NoError(t, err)
	if !actualItem.Equal(item) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", item, actualItem)
	}
}

func TestRedisConnectionGetNoExist(t *testing.T) {
	conn := newTestRedisConnection()

	key := "test_key"
	conn.Delete(key)

	_, err := conn.Get(key)
	assert.EqualError(t, err, "key not found")
}

func TestRedisConnectionPutDirectItem(t *testing.T) {
	conn := newTestRedisConnection()

	key := "test_key"
	conn.Delete(key)

	person := testutil.NewDefaultPerson()
	item := &RedisItem{
		RKey:          key,
		RValue:        util.ToJSONString(person),
		RGroupKeyList: "1",
		RTxnState:     config.COMMITTED,
		RTValid:       -3,
		RTLease:       time.Now().Add(-2 * time.Second),
		RPrev:         "",
		RIsDeleted:    false,
		RVersion:      "2",
	}

	err := conn.Put(key, item)
	assert.NoError(t, err)

	// post check
	str, err := conn.Get(key)
	assert.NoError(t, err)
	var actualItem RedisItem
	err = json.Unmarshal([]byte(str), &actualItem)
	assert.NoError(t, err)
	if !actualItem.Equal(item) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", item, actualItem)
	}
}

func TestRedisConnectionDeleteTwice(t *testing.T) {

	conn := newTestRedisConnection()
	conn.Put("test_key", "test_value")
	err := conn.Delete("test_key")
	assert.NoError(t, err)
	err = conn.Delete("test_key")
	assert.NoError(t, err)
}

func TestRedisConnectionConditionalUpdateDoCreate(t *testing.T) {

	dbItem := &RedisItem{
		RKey:          "item1",
		RValue:        util.ToJSONString(testutil.NewTestItem("item1-db")),
		RGroupKeyList: "1",
		RTxnState:     config.COMMITTED,
		RTValid:       -3,
		RTLease:       time.Now().Add(-2 * time.Second),
		RPrev:         "",
		RIsDeleted:    false,
		RLinkedLen:    1,
		RVersion:      "1",
	}

	cacheItem := &RedisItem{
		RKey:          "item1",
		RValue:        util.ToJSONString(testutil.NewTestItem("item1-cache")),
		RGroupKeyList: "2",
		RTxnState:     config.COMMITTED,
		RTValid:       -2,
		RTLease:       time.Now().Add(-1 * time.Second),
		RPrev:         util.ToJSONString(dbItem),
		RLinkedLen:    2,
		RVersion:      "1",
	}

	t.Run("there is no item and doCreate is true ", func(t *testing.T) {
		conn := newTestRedisConnection()
		conn.Delete(cacheItem.Key())

		_, err := conn.ConditionalUpdate(cacheItem.Key(), cacheItem, true)
		assert.NoError(t, err)
	})

	t.Run("there is an item and doCreate is true ", func(t *testing.T) {
		conn := newTestRedisConnection()
		conn.PutItem(dbItem.Key(), dbItem)

		_, err := conn.ConditionalUpdate(cacheItem.Key(), cacheItem, true)
		assert.EqualError(t, err, txn.VersionMismatch.Error())
	})

	t.Run("there is no item and doCreate is false ", func(t *testing.T) {
		conn := newTestRedisConnection()
		conn.Delete(cacheItem.Key())

		_, err := conn.ConditionalUpdate(cacheItem.Key(), cacheItem, false)
		assert.EqualError(t, err, txn.VersionMismatch.Error())
	})

	t.Run("there is an item and doCreate is false ", func(t *testing.T) {
		conn := newTestRedisConnection()
		conn.PutItem(dbItem.Key(), dbItem)

		_, err := conn.ConditionalUpdate(cacheItem.Key(), cacheItem, false)
		assert.NoError(t, err)
	})

}

func TestRedisConnectionConditionalCommit(t *testing.T) {

	dbItem := &RedisItem{
		RKey:          "item1",
		RValue:        util.ToJSONString(testutil.NewTestItem("item1-db")),
		RGroupKeyList: "1",
		RTxnState:     config.COMMITTED,
		RTValid:       -3,
		RTLease:       time.Now().Add(-2 * time.Second),
		RPrev:         "",
		RIsDeleted:    false,
		RLinkedLen:    1,
		RVersion:      "1",
	}

	conn := newTestRedisConnection()
	conn.Connect()
	conn.PutItem(dbItem.Key(), dbItem)

	_, err := conn.ConditionalCommit(dbItem.Key(), dbItem.Version(), 100)
	assert.NoError(t, err)

	item, err := conn.GetItem(dbItem.Key())
	assert.NoError(t, err)

	dbItem.RVersion = util.AddToString(dbItem.RVersion, 1)
	dbItem.RTxnState = config.COMMITTED
	dbItem.RTValid = 100

	if !dbItem.Equal(item) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", dbItem, item)
		t.Fail()
	}

}
