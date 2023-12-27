package redis

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
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
	redisClient, _ := redismock.NewClientMock()
	connection := &RedisConnection{rdb: redisClient}

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
	redisClient, mock := redismock.NewClientMock()
	connection := &RedisConnection{rdb: redisClient}

	key := "test_key"
	tValidStr := time.Now().Format(time.RFC3339Nano)
	tLeaseStr := time.Now().Format(time.RFC3339Nano)
	tValid, _ := time.Parse(time.RFC3339Nano, tValidStr)
	tLease, _ := time.Parse(time.RFC3339Nano, tLeaseStr)

	expectedValue := testutil.NewDefaultPerson()
	expectedItem := RedisItem{
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
	redisClient, mock := redismock.NewClientMock()
	connection := &RedisConnection{rdb: redisClient}

	key := "test_key"
	mock.ExpectHGetAll(key).SetVal(nil)

	_, err := connection.GetItem(key)

	assert.EqualError(t, err, fmt.Sprintf("key not found: %s", key))
}

func TestRedisConnectionPutAndGet(t *testing.T) {
	conn := NewRedisConnection(nil)

	key := "test_key"
	expectedValue := testutil.NewDefaultPerson()
	expectedItem := RedisItem{
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

func TestRedisConnectionReplaceAndGet(t *testing.T) {
	conn := NewRedisConnection(nil)

	key := "test_key"
	olderPerson := testutil.NewDefaultPerson()
	olderItem := RedisItem{
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
	newerItem := RedisItem{
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
	olderItem := RedisItem{
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
	newerItem := RedisItem{
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

	err = conn.ConditionalUpdate(key, newerItem)
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
	olderItem := RedisItem{
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
	newerItem := RedisItem{
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

	err = conn.ConditionalUpdate(key, newerItem)
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
	newerItem := RedisItem{
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

	err := conn.ConditionalUpdate(key, newerItem)
	assert.NoError(t, err)

	item, err := conn.GetItem(key)
	assert.NoError(t, err)

	newerItem.Version++
	if !item.Equal(newerItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", newerItem, item)
	}
}

func TestRedisConnectionConditionalUpdateConcurrently(t *testing.T) {
	conn := NewRedisConnection(nil)

	key := "test_key"
	olderPerson := testutil.NewDefaultPerson()
	olderItem := RedisItem{
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
			newerItem := RedisItem{
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

			err = conn.ConditionalUpdate(key, newerItem)
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
	if item.TxnId != strconv.Itoa(globalId) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", globalId, item.TxnId)
	}
}
