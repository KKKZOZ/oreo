package redis

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/serializer"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisConnection_DefaultNilArgument(t *testing.T) {
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
	expectedAddress := "127.0.0.1:1234"
	connectionOptions := &ConnectionOptions{Address: expectedAddress}
	connection := NewRedisConnection(connectionOptions)
	assert.NotNil(t, connection)
	assert.Equal(t, expectedAddress, connection.Address)
}

func TestRedisConnection_Connect(t *testing.T) {
	RedisClient, _ := redismock.NewClientMock()
	connection := &RedisConnection{rdb: RedisClient}

	err := connection.Connect()
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

func TestRedisConnection_GetItem(t *testing.T) {
	RedisClient, mock := redismock.NewClientMock()
	connection := &RedisConnection{rdb: RedisClient}

	key := "test_key"
	tValidStr := time.Now().Format(time.RFC3339Nano)
	tLeaseStr := time.Now().Format(time.RFC3339Nano)
	tValid, _ := time.Parse(time.RFC3339Nano, tValidStr)
	tLease, _ := time.Parse(time.RFC3339Nano, tLeaseStr)

	expectedValue := testutil.NewDefaultPerson()
	expectedItem := txn.DataItem{
		Key:       key,
		Value:     util.ToJSONString(expectedValue),
		TxnId:     "1",
		TxnState:  config.COMMITTED,
		TValid:    tValid,
		TLease:    tLease,
		Prev:      "",
		IsDeleted: false,
		Version:   2,
	}
	itemMap := map[string]string{
		"Key":       key,
		"Value":     util.ToJSONString(expectedValue),
		"TxnId":     "1",
		"TxnState":  fmt.Sprint(config.COMMITTED),
		"TValid":    tValidStr,
		"TLease":    tLeaseStr,
		"Prev":      "",
		"isDeleted": "false",
		"Version":   "2",
	}

	mock.ExpectHGetAll(key).SetVal(itemMap)

	actualItem, err := connection.GetItem(key)

	assert.Nil(t, err)
	if actualItem != expectedItem {
		t.Error("Not Equal")
	}
	assert.Equal(t, expectedItem, actualItem)
}

func TestRedisConnection_GetItemNotFound(t *testing.T) {
	RedisClient, mock := redismock.NewClientMock()
	connection := &RedisConnection{rdb: RedisClient}

	key := "test_key"
	mock.ExpectHGetAll(key).SetVal(nil)

	_, err := connection.GetItem(key)

	assert.EqualError(t, err, txn.KeyNotFound.Error())
}

func TestRedisConnectionPutItemAndGetItem(t *testing.T) {
	conn := NewRedisConnection(nil)
	conn.Delete("test_key")

	key := "test_key"
	expectedValue := testutil.NewDefaultPerson()
	expectedItem := txn.DataItem{
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
	assert.NoError(t, err)

	item, err := conn.GetItem(key)

	assert.NoError(t, err)
	if !item.Equal(expectedItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", expectedItem, item)
	}
}

func TestRedisConnectionReplaceAndGetItem(t *testing.T) {
	conn := NewRedisConnection(nil)

	key := "test_key"
	olderPerson := testutil.NewDefaultPerson()
	olderItem := txn.DataItem{
		Key:       key,
		Value:     util.ToJSONString(olderPerson),
		TxnId:     "1",
		TxnState:  config.COMMITTED,
		TValid:    time.Now().Add(-3 * time.Second),
		TLease:    time.Now().Add(-2 * time.Second),
		Prev:      "",
		IsDeleted: false,
		Version:   2,
	}

	err := conn.PutItem(key, olderItem)
	assert.NoError(t, err)

	newerPerson := testutil.NewDefaultPerson()
	newerPerson.Name = "newer"
	newerItem := txn.DataItem{
		Key:       key,
		Value:     util.ToJSONString(newerPerson),
		TxnId:     "2",
		TxnState:  config.COMMITTED,
		TValid:    time.Now().Add(-1 * time.Second),
		TLease:    time.Now().Add(1 * time.Second),
		Prev:      util.ToJSONString(olderItem),
		IsDeleted: false,
		Version:   3,
	}

	err = conn.PutItem(key, newerItem)
	assert.NoError(t, err)

	item, err := conn.GetItem(key)
	assert.NoError(t, err)
	if !item.Equal(newerItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", newerItem, item)
	}
}

func TestRedisConnectionConditionalUpdateSuccess(t *testing.T) {
	conn := NewRedisConnection(nil)

	key := "test_key"
	olderPerson := testutil.NewDefaultPerson()
	olderItem := txn.DataItem{
		Key:       key,
		Value:     util.ToJSONString(olderPerson),
		TxnId:     "1",
		TxnState:  config.COMMITTED,
		TValid:    time.Now().Add(-3 * time.Second),
		TLease:    time.Now().Add(-2 * time.Second),
		Prev:      "",
		IsDeleted: false,
		Version:   2,
	}
	err := conn.PutItem(key, olderItem)
	assert.NoError(t, err)

	newerPerson := testutil.NewDefaultPerson()
	newerPerson.Name = "newer"
	newerItem := txn.DataItem{
		Key:       key,
		Value:     util.ToJSONString(newerPerson),
		TxnId:     "2",
		TxnState:  config.COMMITTED,
		TValid:    time.Now().Add(-2 * time.Second),
		TLease:    time.Now().Add(-1 * time.Second),
		Prev:      "",
		IsDeleted: false,
		Version:   2,
	}

	err = conn.ConditionalUpdate(key, newerItem, false)
	assert.NoError(t, err)

	item, err := conn.GetItem(key)
	assert.NoError(t, err)

	newerItem.Version++
	if !item.Equal(newerItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", newerItem, item)
	}

}

func TestRedisConnectionConditionalUpdateFail(t *testing.T) {
	conn := NewRedisConnection(nil)
	key := "test_key"
	olderPerson := testutil.NewDefaultPerson()
	olderItem := txn.DataItem{
		Key:       key,
		Value:     util.ToJSONString(olderPerson),
		TxnId:     "1",
		TxnState:  config.COMMITTED,
		TValid:    time.Now().Add(-3 * time.Second),
		TLease:    time.Now().Add(-2 * time.Second),
		Prev:      "",
		IsDeleted: false,
		Version:   2,
	}
	err := conn.PutItem(key, olderItem)
	assert.NoError(t, err)

	newerPerson := testutil.NewDefaultPerson()
	newerPerson.Name = "newer"
	newerItem := txn.DataItem{
		Key:       key,
		Value:     util.ToJSONString(olderPerson),
		TxnId:     "2",
		TxnState:  config.COMMITTED,
		TValid:    time.Now().Add(-2 * time.Second),
		TLease:    time.Now().Add(-1 * time.Second),
		Prev:      "",
		IsDeleted: false,
		Version:   3,
	}

	err = conn.ConditionalUpdate(key, newerItem, false)
	assert.EqualError(t, err, "version mismatch")

	item, err := conn.GetItem(key)
	assert.NoError(t, err)
	if !item.Equal(olderItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", olderItem, item)
	}
}

func TestRedisConnectionConditionalUpdateNonExist(t *testing.T) {
	conn := NewRedisConnection(nil)

	key := "test_key"
	conn.Delete(key)
	newerPerson := testutil.NewDefaultPerson()
	newerPerson.Name = "newer"
	newerItem := txn.DataItem{
		Key:       key,
		Value:     util.ToJSONString(newerPerson),
		TxnId:     "2",
		TxnState:  config.COMMITTED,
		TValid:    time.Now().Add(-2 * time.Second),
		TLease:    time.Now().Add(-1 * time.Second),
		Prev:      "",
		IsDeleted: false,
		Version:   1,
	}

	err := conn.ConditionalUpdate(key, newerItem, true)
	assert.NoError(t, err)

	item, err := conn.GetItem(key)
	assert.NoError(t, err)

	newerItem.Version++
	if !item.Equal(newerItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", newerItem, item)
	}
}

func TestRedisConnectionConditionalUpdateConcurrently(t *testing.T) {

	t.Run("this is a update", func(t *testing.T) {
		conn := NewRedisConnection(nil)

		key := "test_key"
		olderPerson := testutil.NewDefaultPerson()
		olderItem := txn.DataItem{
			Key:       key,
			Value:     util.ToJSONString(olderPerson),
			TxnId:     "1",
			TxnState:  config.COMMITTED,
			TValid:    time.Now().Add(-3 * time.Second),
			TLease:    time.Now().Add(-2 * time.Second),
			Prev:      "",
			IsDeleted: false,
			Version:   2,
		}
		err := conn.PutItem(key, olderItem)
		assert.NoError(t, err)

		resChan := make(chan bool)
		currentNum := 100
		globalId := 0
		for i := 1; i <= currentNum; i++ {
			go func(id int) {
				newerPerson := testutil.NewDefaultPerson()
				newerPerson.Name = "newer"
				newerItem := txn.DataItem{
					Key:       key,
					Value:     util.ToJSONString(newerPerson),
					TxnId:     strconv.Itoa(id),
					TxnState:  config.COMMITTED,
					TValid:    time.Now().Add(-2 * time.Second),
					TLease:    time.Now().Add(-1 * time.Second),
					Prev:      "",
					IsDeleted: false,
					Version:   2,
				}

				err = conn.ConditionalUpdate(key, newerItem, false)
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
		if item.TxnId != strconv.Itoa(globalId) {
			t.Errorf("\nexpect: \n%v, \nactual: \n%v", globalId, item.TxnId)
		}
	})

	t.Run("this is a create", func(t *testing.T) {
		conn := NewRedisConnection(nil)
		key := "test_key"
		conn.Delete(key)

		resChan := make(chan bool)
		currentNum := 100
		globalId := 0
		for i := 1; i <= currentNum; i++ {
			go func(id int) {
				newerPerson := testutil.NewDefaultPerson()
				newerPerson.Name = "newer"
				newerItem := txn.DataItem{
					Key:       key,
					Value:     util.ToJSONString(newerPerson),
					TxnId:     strconv.Itoa(id),
					TxnState:  config.COMMITTED,
					TValid:    time.Now().Add(-2 * time.Second),
					TLease:    time.Now().Add(-1 * time.Second),
					Prev:      "",
					IsDeleted: false,
					Version:   2,
				}

				err := conn.ConditionalUpdate(key, newerItem, true)
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
		if item.TxnId != strconv.Itoa(globalId) {
			t.Errorf("\nexpect: \n%v, \nactual: \n%v", globalId, item.TxnId)
		}
	})
}

func TestRedisConnectionPutAndGet(t *testing.T) {
	conn := NewRedisConnection(nil)
	se := serializer.NewJSONSerializer()

	key := "test_key"
	person := testutil.NewDefaultPerson()
	item := txn.DataItem{
		Key:       key,
		Value:     util.ToJSONString(person),
		TxnId:     "1",
		TxnState:  config.COMMITTED,
		TValid:    time.Now().Add(-3 * time.Second),
		TLease:    time.Now().Add(-2 * time.Second),
		Prev:      "",
		IsDeleted: false,
		Version:   2,
	}
	bs, err := se.Serialize(item)
	assert.NoError(t, err)
	err = conn.Put(key, bs)
	assert.NoError(t, err)

	str, err := conn.Get(key)
	assert.NoError(t, err)
	var actualItem txn.DataItem
	err = se.Deserialize([]byte(str), &actualItem)
	assert.NoError(t, err)
	if !actualItem.Equal(item) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", item, actualItem)
	}
}

func TestRedisConnectionReplaceAndGet(t *testing.T) {
	conn := NewRedisConnection(nil)
	se := serializer.NewJSONSerializer()

	key := "test_key"
	person := testutil.NewDefaultPerson()
	item := txn.DataItem{
		Key:       key,
		Value:     util.ToJSONString(person),
		TxnId:     "1",
		TxnState:  config.COMMITTED,
		TValid:    time.Now().Add(-3 * time.Second),
		TLease:    time.Now().Add(-2 * time.Second),
		Prev:      "",
		IsDeleted: false,
		Version:   2,
	}
	bs, err := se.Serialize(item)
	assert.NoError(t, err)
	err = conn.Put(key, bs)
	assert.NoError(t, err)

	item.Version++
	bs, _ = se.Serialize(item)
	err = conn.Put(key, bs)
	assert.NoError(t, err)

	str, err := conn.Get(key)
	assert.NoError(t, err)
	var actualItem txn.DataItem
	err = se.Deserialize([]byte(str), &actualItem)
	assert.NoError(t, err)
	if !actualItem.Equal(item) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", item, actualItem)
	}
}

func TestRedisConnectionGetNoExist(t *testing.T) {
	conn := NewRedisConnection(nil)

	key := "test_key"
	conn.Delete(key)

	_, err := conn.Get(key)
	assert.EqualError(t, err, fmt.Sprintf("key not found: %s", key))
}

func TestRedisConnectionPutDirectItem(t *testing.T) {
	conn := NewRedisConnection(nil)

	key := "test_key"
	conn.Delete(key)

	person := testutil.NewDefaultPerson()
	item := txn.DataItem{
		Key:       key,
		Value:     util.ToJSONString(person),
		TxnId:     "1",
		TxnState:  config.COMMITTED,
		TValid:    time.Now().Add(-3 * time.Second),
		TLease:    time.Now().Add(-2 * time.Second),
		Prev:      "",
		IsDeleted: false,
		Version:   2,
	}

	err := conn.Put(key, item)
	assert.NoError(t, err)

	// post check
	str, err := conn.Get(key)
	assert.NoError(t, err)
	var actualItem txn.DataItem
	err = json.Unmarshal([]byte(str), &actualItem)
	assert.NoError(t, err)
	if !actualItem.Equal(item) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", item, actualItem)
	}
}

func TestRedisConnectionDeleteTwice(t *testing.T) {

	conn := NewRedisConnection(nil)
	conn.Put("test_key", "test_value")
	err := conn.Delete("test_key")
	assert.NoError(t, err)
	err = conn.Delete("test_key")
	assert.NoError(t, err)
}

func TestRedisConnectionConditionalUpdateDoCreate(t *testing.T) {

	dbItem := txn.DataItem{
		Key:       "item1",
		Value:     util.ToJSONString(testutil.NewTestItem("item1-db")),
		TxnId:     "1",
		TxnState:  config.COMMITTED,
		TValid:    time.Now().Add(-3 * time.Second),
		TLease:    time.Now().Add(-2 * time.Second),
		Prev:      "",
		IsDeleted: false,
		LinkedLen: 1,
		Version:   1,
	}

	cacheItem := txn.DataItem{
		Key:       "item1",
		Value:     util.ToJSONString(testutil.NewTestItem("item1-cache")),
		TxnId:     "2",
		TxnState:  config.COMMITTED,
		TValid:    time.Now().Add(-2 * time.Second),
		TLease:    time.Now().Add(-1 * time.Second),
		Prev:      util.ToJSONString(dbItem),
		LinkedLen: 2,
		Version:   1,
	}

	t.Run("there is no item and doCreate is true ", func(t *testing.T) {
		conn := NewRedisConnection(nil)
		conn.Delete(cacheItem.Key)

		err := conn.ConditionalUpdate(cacheItem.Key, cacheItem, true)
		assert.NoError(t, err)
	})

	t.Run("there is an item and doCreate is true ", func(t *testing.T) {
		conn := NewRedisConnection(nil)
		conn.PutItem(dbItem.Key, dbItem)

		err := conn.ConditionalUpdate(cacheItem.Key, cacheItem, true)
		assert.EqualError(t, err, txn.VersionMismatch.Error())
	})

	t.Run("there is no item and doCreate is false ", func(t *testing.T) {
		conn := NewRedisConnection(nil)
		conn.Delete(cacheItem.Key)

		err := conn.ConditionalUpdate(cacheItem.Key, cacheItem, false)
		assert.EqualError(t, err, txn.VersionMismatch.Error())
	})

	t.Run("there is an item and doCreate is false ", func(t *testing.T) {
		conn := NewRedisConnection(nil)
		conn.PutItem(dbItem.Key, dbItem)

		err := conn.ConditionalUpdate(cacheItem.Key, cacheItem, false)
		assert.NoError(t, err)
	})

}
