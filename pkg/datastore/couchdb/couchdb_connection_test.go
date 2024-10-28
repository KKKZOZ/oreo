package couchdb

import (
	"strconv"
	"testing"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/testutil"
	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/serializer"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
)

func TestNewCouchDBConnection_DefaultNilArgument(t *testing.T) {
	connection := NewCouchDBConnection(nil)
	assert.NotNil(t, connection)
	assert.Equal(t, "http://admin:password@localhost:5984", connection.Address)
}

func TestNewCouchDBConnection_DefaultAddress(t *testing.T) {
	connectionOptions := &ConnectionOptions{}
	connection := NewCouchDBConnection(connectionOptions)
	assert.NotNil(t, connection)
	assert.Equal(t, "http://admin:password@localhost:5984", connection.Address)
}

func TestNewCouchDBConnection_WithAddress(t *testing.T) {
	expectedAddress := "http://admin:password@localhost:5984"
	connectionOptions := &ConnectionOptions{Address: expectedAddress}
	connection := NewCouchDBConnection(connectionOptions)
	assert.NotNil(t, connection)
	assert.Equal(t, expectedAddress, connection.Address)
}

func TestCouchDBConnection_Connect(t *testing.T) {
	connection := NewCouchDBConnection(nil)
	err := connection.Connect()
	assert.Nil(t, err)
}

func TestCouchDBConnection_ConnectWithInvalidAddress(t *testing.T) {
	connectionOptions := &ConnectionOptions{Address: "invalid_address"}
	connection := NewCouchDBConnection(connectionOptions)
	err := connection.Connect()
	assert.NotNil(t, err)
}

func TestCouchDBConnection_UseWithoutConnect(t *testing.T) {
	connection := NewCouchDBConnection(nil)
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

func TestCouchDBConnection_GetItemNotFound(t *testing.T) {
	connection := NewCouchDBConnection(nil)
	err := connection.Connect()
	assert.Nil(t, err)
	key := "not_found"
	_, err = connection.GetItem(key)
	assert.EqualError(t, err, txn.KeyNotFound.Error())
}

func TestMongoConnectionPutItemAndGetItem(t *testing.T) {
	conn := NewCouchDBConnection(nil)
	err := conn.Connect()
	assert.NoError(t, err)
	conn.Delete("test_key")

	key := "test_key"
	expectedValue := testutil.NewDefaultPerson()
	expectedItem := &CouchDBItem{
		CKey:          key,
		CValue:        util.ToJSONString(expectedValue),
		CGroupKeyList: "1",
		CTxnState:     config.COMMITTED,
		CTValid:       time.Now().Add(-3 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(-2 * time.Second),
		CPrev:         "",
		CIsDeleted:    false,
		// CVersion:   "1-ba601809ae3e3beb2e05af59242beb43",
	}

	_, err = conn.PutItem(key, expectedItem)
	assert.NoError(t, err)

	item, err := conn.GetItem(key)

	assert.NoError(t, err)
	expectedItem.CVersion = item.Version()
	if !item.Equal(expectedItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", expectedItem, item)
	}
}

func TestMongoConnectionReplaceAndGetItem(t *testing.T) {
	conn := NewCouchDBConnection(nil)
	err := conn.Connect()
	assert.NoError(t, err)
	conn.Delete("test_key")

	key := "test_key"
	olderPerson := testutil.NewDefaultPerson()
	olderItem := &CouchDBItem{
		CKey:          key,
		CValue:        util.ToJSONString(olderPerson),
		CGroupKeyList: "1",
		CTxnState:     config.COMMITTED,
		CTValid:       time.Now().Add(-3 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(-2 * time.Second),
		CPrev:         "",
		CIsDeleted:    false,
		// CVersion:   "2-asd",
	}

	rev, err := conn.PutItem(key, olderItem)
	assert.NoError(t, err)

	newerPerson := testutil.NewDefaultPerson()
	newerPerson.Name = "newer"
	newerItem := &CouchDBItem{
		CKey:          key,
		CValue:        util.ToJSONString(newerPerson),
		CGroupKeyList: "2",
		CTxnState:     config.COMMITTED,
		CTValid:       time.Now().Add(-1 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(1 * time.Second),
		CPrev:         util.ToJSONString(olderItem),
		CIsDeleted:    false,
		CVersion:      rev,
	}

	rev, err = conn.PutItem(key, newerItem)
	assert.NoError(t, err)

	item, err := conn.GetItem(key)
	assert.NoError(t, err)
	newerItem.CVersion = rev
	if !item.Equal(newerItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", newerItem, item)
	}
}

func TestCouchDBConnection_DeleteItem(t *testing.T) {
	conn := NewCouchDBConnection(nil)
	conn.Connect()

	key := "test_key_for_delete"
	person := testutil.NewDefaultPerson()
	item := &CouchDBItem{
		CKey:          key,
		CValue:        util.ToJSONString(person),
		CGroupKeyList: "1",
		CTxnState:     config.COMMITTED,
		CTValid:       time.Now().Add(-3 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(-2 * time.Second),
		CPrev:         "",
		CIsDeleted:    false,
		// CVersion:   "2",
	}
	_, err := conn.PutItem(key, item)
	assert.NoError(t, err)

	err = conn.Delete(key)
	assert.NoError(t, err)

	_, err = conn.GetItem(key)
	assert.EqualError(t, err, txn.KeyNotFound.Error())
}

func TestCouchDBConnection_DeleteItemNotFound(t *testing.T) {
	conn := NewCouchDBConnection(nil)
	conn.Connect()

	key := "test_key_for_delete_not_found"
	err := conn.Delete(key)
	assert.EqualError(t, err, "Not Found: missing")
	// assert.NoError(t, err)
}

func TestCouchDBConnection_DeleteTSR(t *testing.T) {
	conn := NewCouchDBConnection(nil)
	conn.Connect()

	key := "test_key_for_delete_tsr"

	err := conn.Put(key, util.ToString(config.COMMITTED))
	assert.NoError(t, err)

	err = conn.Delete(key)
	assert.NoError(t, err)
}

func TestCouchDBConnection_ConditionalUpdateSuccess(t *testing.T) {
	conn := NewCouchDBConnection(nil)
	conn.Connect()
	conn.Delete("test_key")

	key := "test_key"
	olderPerson := testutil.NewDefaultPerson()
	olderItem := &CouchDBItem{
		CKey:          key,
		CValue:        util.ToJSONString(olderPerson),
		CGroupKeyList: "1",
		CTxnState:     config.COMMITTED,
		CTValid:       time.Now().Add(-3 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(-2 * time.Second),
		CPrev:         "",
		CIsDeleted:    false,
		// CVersion:   "2",
	}
	rev, err := conn.PutItem(key, olderItem)
	assert.NoError(t, err)

	newerPerson := testutil.NewDefaultPerson()
	newerPerson.Name = "newer"
	newerItem := &CouchDBItem{
		CKey:          key,
		CValue:        util.ToJSONString(newerPerson),
		CGroupKeyList: "2",
		CTxnState:     config.COMMITTED,
		CTValid:       time.Now().Add(-2 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(-1 * time.Second),
		CPrev:         "",
		CIsDeleted:    false,
		CVersion:      rev,
	}

	rev, err = conn.ConditionalUpdate(key, newerItem, false)
	assert.NoError(t, err)

	item, err := conn.GetItem(key)
	assert.NoError(t, err)

	newerItem.CVersion = rev
	if !item.Equal(newerItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", newerItem, item)
	}

}

func TestCouchDBConnection_ConditionalUpdateFail(t *testing.T) {
	conn := NewCouchDBConnection(nil)
	conn.Connect()
	conn.Delete("test_key")

	key := "test_key"
	olderPerson := testutil.NewDefaultPerson()
	olderItem := &CouchDBItem{
		CKey:          key,
		CValue:        util.ToJSONString(olderPerson),
		CGroupKeyList: "1",
		CTxnState:     config.COMMITTED,
		CTValid:       time.Now().Add(-3 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(-2 * time.Second),
		CPrev:         "",
		CIsDeleted:    false,
		// CVersion:   "2",
	}
	rev, err := conn.PutItem(key, olderItem)
	olderItem.CVersion = rev
	assert.NoError(t, err)

	newerPerson := testutil.NewDefaultPerson()
	newerPerson.Name = "newer"
	newerItem := &CouchDBItem{
		CKey:          key,
		CValue:        util.ToJSONString(olderPerson),
		CGroupKeyList: "2",
		CTxnState:     config.COMMITTED,
		CTValid:       time.Now().Add(-2 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(-1 * time.Second),
		CPrev:         "",
		CIsDeleted:    false,
		CVersion:      "3",
	}

	_, err = conn.ConditionalUpdate(key, newerItem, false)
	assert.EqualError(t, err, txn.VersionMismatch.Error())

	item, err := conn.GetItem(key)
	assert.NoError(t, err)
	if !item.Equal(olderItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", olderItem, item)
	}
}

func TestCouchDBConnection_ConditionalUpdateNonExist(t *testing.T) {
	conn := NewCouchDBConnection(nil)
	conn.Connect()

	key := "test_key"
	conn.Delete(key)
	newerPerson := testutil.NewDefaultPerson()
	newerPerson.Name = "newer"
	newerItem := &CouchDBItem{
		CKey:          key,
		CValue:        util.ToJSONString(newerPerson),
		CGroupKeyList: "2",
		CTxnState:     config.COMMITTED,
		CTValid:       time.Now().Add(-2 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(-1 * time.Second),
		CPrev:         "",
		CIsDeleted:    false,
		// CVersion:   "1",
	}

	rev, err := conn.ConditionalUpdate(key, newerItem, true)
	newerItem.CVersion = rev
	assert.NoError(t, err)

	item, err := conn.GetItem(key)
	assert.NoError(t, err)

	// newerItem.CVersion = util.AddToString(newerItem.CVersion, 1)
	if !item.Equal(newerItem) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", newerItem, item)
	}
}

func TestCouchDBConnection_ConditionalUpdateConcurrently(t *testing.T) {

	t.Run("this is a update", func(t *testing.T) {
		conn := NewDefaultConnection()
		conn.Delete("test_key")

		key := "test_key"
		olderPerson := testutil.NewDefaultPerson()
		olderItem := &CouchDBItem{
			CKey:          key,
			CValue:        util.ToJSONString(olderPerson),
			CGroupKeyList: "1",
			CTxnState:     config.COMMITTED,
			CTValid:       time.Now().Add(-3 * time.Second).UnixMicro(),
			CTLease:       time.Now().Add(-2 * time.Second),
			CPrev:         "",
			CIsDeleted:    false,
			// CVersion:   "2",
		}
		rev, err := conn.PutItem(key, olderItem)
		assert.NoError(t, err)

		resChan := make(chan bool)
		currentNum := 50
		globalId := 0
		for i := 1; i <= currentNum; i++ {
			go func(id int) {
				newerPerson := testutil.NewDefaultPerson()
				newerPerson.Name = "newer"
				newerItem := &CouchDBItem{
					CKey:          key,
					CValue:        util.ToJSONString(newerPerson),
					CGroupKeyList: strconv.Itoa(id),
					CTxnState:     config.COMMITTED,
					CTValid:       time.Now().Add(-2 * time.Second).UnixMicro(),
					CTLease:       time.Now().Add(-1 * time.Second),
					CPrev:         "",
					CIsDeleted:    false,
					CVersion:      rev,
				}

				_, err = conn.ConditionalUpdate(key, newerItem, false)
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
		if item.TxnId() != strconv.Itoa(globalId) {
			t.Errorf("\nexpect: \n%v, \nactual: \n%v", globalId, item.TxnId())
		}
	})

	t.Run("this is a create", func(t *testing.T) {
		conn := NewDefaultConnection()
		key := "test_key"
		conn.Delete(key)

		resChan := make(chan bool)
		currentNum := 50
		globalId := 0
		for i := 1; i <= currentNum; i++ {
			go func(id int) {
				newerPerson := testutil.NewDefaultPerson()
				newerPerson.Name = "newer"
				newerItem := &CouchDBItem{
					CKey:          key,
					CValue:        util.ToJSONString(newerPerson),
					CGroupKeyList: strconv.Itoa(id),
					CTxnState:     config.COMMITTED,
					CTValid:       time.Now().Add(-2 * time.Second).UnixMicro(),
					CTLease:       time.Now().Add(-1 * time.Second),
					CPrev:         "",
					CIsDeleted:    false,
					// CVersion:   "2",
				}

				_, err := conn.ConditionalUpdate(key, newerItem, true)
				if err == nil {
					globalId = id
					resChan <- true
				} else {
					// assert.EqualError(t, err, txn.VersionMismatch.Error())
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
		if item.TxnId() != strconv.Itoa(globalId) {
			t.Errorf("\nexpect: \n%v, \nactual: \n%v", globalId, item.TxnId())
		}
	})
}

func TestCouchDBConnection_SimplePutAndGet(t *testing.T) {
	conn := NewCouchDBConnection(nil)
	conn.Connect()
	conn.Delete("test_key")

	err := conn.Put("test_key", "test_value")
	assert.NoError(t, err)

	value, err := conn.Get("test_key")
	assert.NoError(t, err)

	assert.Equal(t, "test_value", value)
}

func TestCouchDBConnection_PutAndGet(t *testing.T) {
	conn := NewCouchDBConnection(nil)
	conn.Connect()
	conn.Delete("test_key")
	se := serializer.NewJSONSerializer()

	key := "test_key"
	person := testutil.NewDefaultPerson()
	item := &CouchDBItem{
		CKey:          key,
		CValue:        util.ToJSONString(person),
		CGroupKeyList: "1",
		CTxnState:     config.COMMITTED,
		CTValid:       time.Now().Add(-3 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(-2 * time.Second),
		CPrev:         "",
		CIsDeleted:    false,
		// CVersion:   "2",
	}
	bs, err := se.Serialize(item)
	assert.NoError(t, err)
	err = conn.Put(key, string(bs))
	assert.NoError(t, err)

	str, err := conn.Get(key)
	assert.NoError(t, err)
	var actualItem CouchDBItem
	err = se.Deserialize([]byte(str), &actualItem)
	assert.NoError(t, err)
	if !actualItem.Equal(item) {
		t.Errorf("\nexpect: \n%v, \nactual: \n%v", item, actualItem)
	}
}

// func TestCouchDBConnection_ReplaceAndGet(t *testing.T) {
// 	t.SkipNow()
// 	conn := NewCouchDBConnection(nil)
// 	conn.Connect()
// 	conn.Delete("test_key")
// 	se := serializer.NewJSONSerializer()

// 	key := "test_key"
// 	person := testutil.NewDefaultPerson()
// 	item := &CouchDBItem{
// 		CKey:       key,
// 		CValue:     util.ToJSONString(person),
// 		CGroupKeyList:     "1",
// 		CTxnState:  config.COMMITTED,
// 		CTValid:    time.Now().Add(-3 * time.Second),
// 		CTLease:    time.Now().Add(-2 * time.Second),
// 		CPrev:      "",
// 		CIsDeleted: false,
// 		// CVersion:   "2",
// 	}
// 	bs, err := se.Serialize(item)
// 	assert.NoError(t, err)
// 	err = conn.Put(key, string(bs))
// 	assert.NoError(t, err)

// 	item.CVersion = util.AddToString(item.CVersion, 1)
// 	bs, _ = se.Serialize(item)
// 	err = conn.Put(key, string(bs))
// 	assert.NoError(t, err)

// 	str, err := conn.Get(key)
// 	assert.NoError(t, err)
// 	var actualItem CouchDBItem
// 	err = se.Deserialize([]byte(str), &actualItem)
// 	assert.NoError(t, err)
// 	if !actualItem.Equal(item) {
// 		t.Errorf("\nexpect: \n%v, \nactual: \n%v", item, actualItem)
// 	}
// }

func TestCouchDBConnection_GetNoExist(t *testing.T) {
	conn := NewCouchDBConnection(nil)
	conn.Connect()

	key := "test_key"
	conn.Delete(key)

	_, err := conn.Get(key)
	assert.EqualError(t, err, txn.KeyNotFound.Error())
}

// func TestCouchDBConnection_PutDirectItem(t *testing.T) {
// 	conn := NewCouchDBConnection(nil)
// 	conn.Connect()

// 	key := "test_key"
// 	conn.Delete(key)

// 	person := testutil.NewDefaultPerson()
// 	item := trxn.DataItem{
// 		CKey:       key,
// 		CValue:     util.ToJSONString(person),
// 		CGroupKeyList:     "1",
// 		CTxnState:  config.COMMITTED,
// 		CTValid:    time.Now().Add(-3 * time.Second),
// 		CTLease:    time.Now().Add(-2 * time.Second),
// 		CPrev:      "",
// 		CIsDeleted: false,
// 		CVersion:   "2",
// 	}

// 	err := conn.Put(key, item)
// 	assert.NoError(t, err)

// 	// post check
// 	str, err := conn.Get(key)
// 	assert.NoError(t, err)
// 	var actualItem trxn.DataItem
// 	err = json.Unmarshal([]byte(str), &actualItem)
// 	assert.NoError(t, err)
// 	if !actualItem.Equal(item) {
// 		t.Errorf("\nexpect: \n%v, \nactual: \n%v", item, actualItem)
// 	}
// }

func TestCouchDBConnection_DeleteTwice(t *testing.T) {
	conn := NewCouchDBConnection(nil)
	conn.Connect()
	conn.Put("test_key", "test_value")
	err := conn.Delete("test_key")
	assert.NoError(t, err)
	err = conn.Delete("test_key")
	assert.NotNil(t, err)
}

func TestCouchDBConnection_ConditionalUpdateDoCreate(t *testing.T) {

	dbItem := &CouchDBItem{
		CKey:          "item1",
		CValue:        util.ToJSONString(testutil.NewTestItem("item1-db")),
		CGroupKeyList: "1",
		CTxnState:     config.COMMITTED,
		CTValid:       time.Now().Add(-3 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(-2 * time.Second),
		CPrev:         "",
		CIsDeleted:    false,
		CLinkedLen:    1,
		// CVersion:   "1",
	}

	cacheItem := &CouchDBItem{
		CKey:          "item1",
		CValue:        util.ToJSONString(testutil.NewTestItem("item1-cache")),
		CGroupKeyList: "2",
		CTxnState:     config.COMMITTED,
		CTValid:       time.Now().Add(-2 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(-1 * time.Second),
		CPrev:         util.ToJSONString(dbItem),
		CLinkedLen:    2,
		// CVersion:   "1",
	}

	t.Run("there is no item and doCreate is true ", func(t *testing.T) {
		conn := NewDefaultConnection()
		conn.Delete(cacheItem.Key())

		_, err := conn.ConditionalUpdate(cacheItem.Key(), cacheItem, true)
		assert.NoError(t, err)
	})

	t.Run("there is an item and doCreate is true ", func(t *testing.T) {
		conn := NewDefaultConnection()
		conn.Delete(dbItem.CKey)
		rev, err := conn.PutItem(dbItem.Key(), dbItem)
		assert.NoError(t, err)

		cacheItem.CVersion = rev
		_, err = conn.ConditionalUpdate(cacheItem.Key(), cacheItem, true)
		assert.EqualError(t, err, txn.VersionMismatch.Error())
	})

	t.Run("there is no item and doCreate is false ", func(t *testing.T) {
		conn := NewDefaultConnection()
		conn.Delete(cacheItem.Key())

		_, err := conn.ConditionalUpdate(cacheItem.Key(), cacheItem, false)
		assert.EqualError(t, err, txn.VersionMismatch.Error())
	})

	t.Run("there is an item and doCreate is false ", func(t *testing.T) {
		conn := NewDefaultConnection()
		conn.Delete(dbItem.CKey)
		rev, err := conn.PutItem(dbItem.Key(), dbItem)
		assert.NoError(t, err)

		cacheItem.CVersion = rev
		_, err = conn.ConditionalUpdate(cacheItem.Key(), cacheItem, false)
		assert.NoError(t, err)
	})
}

func TestMaxConnections(t *testing.T) {
	// num:=100
	num := 10
	for i := 1; i <= num; i++ {
		conn := NewDefaultConnection()
		err := conn.Connect()
		// err := conn.Put("test_key", "test_value")
		assert.NoError(t, err)
	}
}
