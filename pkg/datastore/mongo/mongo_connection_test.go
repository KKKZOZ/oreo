package mongo

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/serializer"
	"github.com/stretchr/testify/assert"
)

func TestNewMongoConnection_DefaultNilArgument(t *testing.T) {
	connection := NewMongoConnection(nil)
	assert.NotNil(t, connection)
	assert.Equal(t, "mongodb://localhost:27017", connection.Address)
}

func TestNewMongoConnection_DefaultAddress(t *testing.T) {
	connectionOptions := &ConnectionOptions{}
	connection := NewMongoConnection(connectionOptions)
	assert.NotNil(t, connection)
	assert.Equal(t, "mongodb://localhost:27017", connection.Address)
}

func TestNewMongoConnection_WithAddress(t *testing.T) {
	expectedAddress := "127.0.0.1:1234"
	connectionOptions := &ConnectionOptions{Address: expectedAddress}
	connection := NewMongoConnection(connectionOptions)
	assert.NotNil(t, connection)
	assert.Equal(t, expectedAddress, connection.Address)
}

func TestMongoConnection_Connect(t *testing.T) {
	connection := NewMongoConnection(nil)
	err := connection.Connect()
	assert.Nil(t, err)
}

func TestMongoConnection_ConnectWithInvalidAddress(t *testing.T) {
	connectionOptions := &ConnectionOptions{Address: "invalid_address"}
	connection := NewMongoConnection(connectionOptions)
	err := connection.Connect()
	assert.NotNil(t, err)
}

func TestMongoConnection_UseWithoutConnect(t *testing.T) {
	connection := NewMongoConnection(nil)
	err := connection.Delete("test_key")
	assert.NotNil(t, err)
}

func TestTimestamp(t *testing.T) {
	tValid := time.Now().Add(-3 * time.Second)
	// tValidStr :=
	t1, _ := time.Parse(time.RFC3339Nano, tValid.Format(time.RFC3339Nano))
	// assert.Equal(t, tValid, t1)
	if !t1.Equal(tValid) {
		t.Error("Not Equal")
	}
}

func TestMongoConnection_GetItemNotFound(t *testing.T) {
	connection := NewMongoConnection(nil)
	err := connection.Connect()
	assert.Nil(t, err)
	key := "not_found"
	_, err = connection.GetItem(key)
	assert.EqualError(t, err, fmt.Sprintf("key not found: %s", key))
}

func TestMongoConnectionPutItemAndGetItem(t *testing.T) {
	conn := NewMongoConnection(nil)
	err := conn.Connect()
	assert.NoError(t, err)
	conn.Delete("test_key")

	key := "test_key"
	expectedValue := testutil.NewDefaultPerson()
	expectedItem := MongoItem{
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

	err = conn.PutItem(key, expectedItem)
	assert.NoError(t, err)

	item, err := conn.GetItem(key)

	assert.NoError(t, err)
	if !item.Equal(expectedItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", expectedItem, item)
	}
}

func TestMongoConnectionReplaceAndGetItem(t *testing.T) {
	conn := NewMongoConnection(nil)
	conn.Connect()

	key := "test_key"
	olderPerson := testutil.NewDefaultPerson()
	olderItem := MongoItem{
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
	newerItem := MongoItem{
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

func TestMongoConnection_DeleteItem(t *testing.T) {
	conn := NewMongoConnection(nil)
	conn.Connect()

	key := "test_key_for_delete"
	person := testutil.NewDefaultPerson()
	item := MongoItem{
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
	err := conn.PutItem(key, item)
	assert.NoError(t, err)

	err = conn.Delete(key)
	assert.NoError(t, err)

	_, err = conn.GetItem(key)
	assert.EqualError(t, err, fmt.Sprintf("key not found: %s", key))
}

func TestMongoConnection_DeleteItemNotFound(t *testing.T) {
	conn := NewMongoConnection(nil)
	conn.Connect()

	key := "test_key_for_delete_not_found"
	err := conn.Delete(key)
	assert.NoError(t, err)
}

func TestMongoConnection_DeleteTSR(t *testing.T) {
	conn := NewMongoConnection(nil)
	conn.Connect()

	key := "test_key_for_delete_tsr"

	err := conn.Put(key, util.ToString(config.COMMITTED))
	assert.NoError(t, err)

	err = conn.Delete(key)
	assert.NoError(t, err)
}

func TestMongoConnection_ConditionalUpdateSuccess(t *testing.T) {
	conn := NewMongoConnection(nil)
	conn.Connect()
	conn.Delete("test_key")

	key := "test_key"
	olderPerson := testutil.NewDefaultPerson()
	olderItem := MongoItem{
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
	newerItem := MongoItem{
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

func TestMongoConnection_ConditionalUpdateFail(t *testing.T) {
	conn := NewMongoConnection(nil)
	conn.Connect()
	conn.Delete("test_key")

	key := "test_key"
	olderPerson := testutil.NewDefaultPerson()
	olderItem := MongoItem{
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
	newerItem := MongoItem{
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
	assert.EqualError(t, err, fmt.Sprintf("version mismatch while updating key: %s", key))

	item, err := conn.GetItem(key)
	assert.NoError(t, err)
	if !item.Equal(olderItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", olderItem, item)
	}
}

func TestMongoConnection_ConditionalUpdateNonExist(t *testing.T) {
	conn := NewMongoConnection(nil)
	conn.Connect()

	key := "test_key"
	conn.Delete(key)
	newerPerson := testutil.NewDefaultPerson()
	newerPerson.Name = "newer"
	newerItem := MongoItem{
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
	assert.EqualError(t, err, fmt.Sprintf("version mismatch while updating key: %s", key))

}

func TestMongoConnection_ConditionalUpdateConcurrently(t *testing.T) {
	conn := NewMongoConnection(nil)
	conn.Connect()
	conn.Delete("test_key")

	key := "test_key"
	olderPerson := testutil.NewDefaultPerson()
	olderItem := MongoItem{
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
	currentNum := 50
	globalId := 0
	for i := 1; i <= currentNum; i++ {
		go func(id int) {
			newerPerson := testutil.NewDefaultPerson()
			newerPerson.Name = "newer"
			newerItem := MongoItem{
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

func TestMongoConnection_SimplePutAndGet(t *testing.T) {
	conn := NewMongoConnection(nil)
	conn.Connect()

	err := conn.Put("test_key", "test_value")
	assert.NoError(t, err)

	value, err := conn.Get("test_key")
	assert.NoError(t, err)

	assert.Equal(t, "test_value", value)
}

func TestMongoConnection_PutAndGet(t *testing.T) {
	conn := NewMongoConnection(nil)
	conn.Connect()
	se := serializer.NewJSONSerializer()

	key := "test_key"
	person := testutil.NewDefaultPerson()
	item := MongoItem{
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
	var actualItem MongoItem
	err = se.Deserialize([]byte(str), &actualItem)
	assert.NoError(t, err)
	if !actualItem.Equal(item) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", item, actualItem)
	}
}

func TestMongoConnection_ReplaceAndGet(t *testing.T) {
	conn := NewMongoConnection(nil)
	conn.Connect()
	se := serializer.NewJSONSerializer()

	key := "test_key"
	person := testutil.NewDefaultPerson()
	item := MongoItem{
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
	err = conn.Put(key, string(bs))
	assert.NoError(t, err)

	item.Version++
	bs, _ = se.Serialize(item)
	err = conn.Put(key, string(bs))
	assert.NoError(t, err)

	str, err := conn.Get(key)
	assert.NoError(t, err)
	var actualItem MongoItem
	err = se.Deserialize([]byte(str), &actualItem)
	assert.NoError(t, err)
	if !actualItem.Equal(item) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", item, actualItem)
	}
}

func TestMongoConnection_GetNoExist(t *testing.T) {
	conn := NewMongoConnection(nil)
	conn.Connect()

	key := "test_key"
	conn.Delete(key)

	_, err := conn.Get(key)
	assert.EqualError(t, err, fmt.Sprintf("key not found: %s", key))
}

// func TestMongoConnection_PutDirectItem(t *testing.T) {
// 	conn := NewMongoConnection(nil)
// 	conn.Connect()

// 	key := "test_key"
// 	conn.Delete(key)

// 	person := testutil.NewDefaultPerson()
// 	item := MongoItem{
// 		Key:       key,
// 		Value:     util.ToJSONString(person),
// 		TxnId:     "1",
// 		TxnState:  config.COMMITTED,
// 		TValid:    time.Now().Add(-3 * time.Second),
// 		TLease:    time.Now().Add(-2 * time.Second),
// 		Prev:      "",
// 		IsDeleted: false,
// 		Version:   2,
// 	}

// 	err := conn.Put(key, item)
// 	assert.NoError(t, err)

// 	// post check
// 	str, err := conn.Get(key)
// 	assert.NoError(t, err)
// 	var actualItem MongoItem
// 	err = json.Unmarshal([]byte(str), &actualItem)
// 	assert.NoError(t, err)
// 	if !actualItem.Equal(item) {
// 		t.Errorf("\nexpect: \n%v, \nactual: \n%v", item, actualItem)
// 	}
// }

func TestMongoConnection_DeleteTwice(t *testing.T) {
	conn := NewMongoConnection(nil)
	conn.Connect()
	conn.Put("test_key", "test_value")
	err := conn.Delete("test_key")
	assert.NoError(t, err)
	err = conn.Delete("test_key")
	assert.NoError(t, err)
}
