package testsuite

import (
	"fmt"
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

type Helper interface {
	MakeItem(txn.ItemOptions) txn.DataItem
	NewInstance() txn.DataItem
}

func testConnection_Connect(t *testing.T, conn txn.Connector, _ Helper) {
	assert.NoError(t, conn.Connect())
}

func testTimestamp(t *testing.T, _ txn.Connector, _ Helper) {
	tValid := time.Now().Add(-3 * time.Second)
	t1, _ := time.Parse(time.RFC3339Nano, tValid.Format(time.RFC3339Nano))
	if !t1.Equal(tValid) {
		t.Error("Not Equal")
	}
}

func testConnection_GetItemNotFound(t *testing.T, conn txn.Connector, _ Helper) {
	key := "test_key_not_found"
	conn.Delete(key)
	_, err := conn.GetItem(key)
	assert.EqualError(t, err, txn.KeyNotFound.Error())
}

func testConnectionPutItemAndGetItem(t *testing.T, conn txn.Connector, h Helper) {
	key := "test_key_put_get"
	conn.Delete(key)

	expectedValue := testutil.NewDefaultPerson()
	expectedItem := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        expectedValue,
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -1,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "2",
	})

	_, err := conn.PutItem(key, expectedItem)
	assert.NoError(t, err)

	item, err := conn.GetItem(key)
	assert.NoError(t, err)
	assert.True(t, item.Equal(expectedItem))
}

func testConnectionReplaceAndGetItem(t *testing.T, conn txn.Connector, h Helper) {
	key := "test_key_replace_get"
	conn.Delete(key)

	older := testutil.NewDefaultPerson()
	olderItem := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        older,
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -3,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "2",
	})

	_, _ = conn.PutItem(key, olderItem)

	newer := testutil.NewDefaultPerson()
	newer.Name = "newer"
	newerItem := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        newer,
		GroupKeyList: "2",
		TxnState:     config.COMMITTED,
		TValid:       -1,
		TLease:       time.Now().Add(1 * time.Second),
		Prev:         util.ToJSONString(olderItem),
		Version:      "3",
	})

	_, _ = conn.PutItem(key, newerItem)

	item, err := conn.GetItem(key)
	assert.NoError(t, err)
	assert.True(t, item.Equal(newerItem))
}

func testConnectionConditionalUpdateSuccess(t *testing.T, conn txn.Connector, h Helper) {
	key := "test_key_cu_success"
	conn.Delete(key)

	// old item
	olderItem := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        testutil.NewDefaultPerson(),
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -3,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "2",
	})
	_, _ = conn.PutItem(key, olderItem)

	// new item
	newerP := testutil.NewDefaultPerson()
	newerP.Name = "newer"
	newerItem := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        newerP,
		GroupKeyList: "2",
		TxnState:     config.COMMITTED,
		TValid:       -2,
		TLease:       time.Now().Add(-1 * time.Second),
		Version:      "2",
	})

	_, err := conn.ConditionalUpdate(key, newerItem, false)
	assert.NoError(t, err)

	item, _ := conn.GetItem(key)
	newerItem.SetVersion(util.AddToString(newerItem.Version(), 1))
	assert.True(t, item.Equal(newerItem))
}

func testConnectionConditionalUpdateFail(t *testing.T, conn txn.Connector, h Helper) {
	key := "test_key_cu_fail"
	conn.Delete(key)

	olderItem := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        testutil.NewDefaultPerson(),
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -3,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "2",
	})
	_, _ = conn.PutItem(key, olderItem)

	newerItem := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        testutil.NewDefaultPerson(),
		GroupKeyList: "2",
		TxnState:     config.COMMITTED,
		TValid:       -2,
		TLease:       time.Now().Add(-1 * time.Second),
		Version:      "3", // version mismatch
	})

	_, err := conn.ConditionalUpdate(key, newerItem, false)
	assert.EqualError(t, err, "version mismatch")

	item, _ := conn.GetItem(key)
	assert.True(t, item.Equal(olderItem))
}

func testConnectionConditionalUpdateNonExist(t *testing.T, conn txn.Connector, h Helper) {
	key := "test_key_cu_nonexist"
	conn.Delete(key)

	newer := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        testutil.NewDefaultPerson(),
		GroupKeyList: "2",
		TxnState:     config.COMMITTED,
		TValid:       -2,
		TLease:       time.Now().Add(-1 * time.Second),
		Version:      "1",
	})

	_, err := conn.ConditionalUpdate(key, newer, true)
	assert.NoError(t, err)

	item, _ := conn.GetItem(key)
	newer.SetVersion(util.AddToString(newer.Version(), 1))
	assert.True(t, item.Equal(newer))
}

func testConnectionConditionalUpdateConcurrently(t *testing.T, conn txn.Connector, h Helper) {
	// update existing key in parallel
	t.Run("update-in-parallel", func(t *testing.T) {
		key := "test_key_concurrent_update"
		conn.Delete(key)

		base := h.MakeItem(txn.ItemOptions{
			Key:          key,
			Value:        testutil.NewDefaultPerson(),
			GroupKeyList: "1",
			TxnState:     config.COMMITTED,
			TValid:       -3,
			TLease:       time.Now().Add(-2 * time.Second),
			Version:      "2",
		})
		_, _ = conn.PutItem(key, base)

		resCh := make(chan bool)
		currentNum := 100
		globalID := 0

		for i := 1; i <= currentNum; i++ {
			go func(id int) {
				p := testutil.NewDefaultPerson()
				p.Name = "newer"
				item := h.MakeItem(txn.ItemOptions{
					Key:          key,
					Value:        p,
					GroupKeyList: strconv.Itoa(id),
					TxnState:     config.COMMITTED,
					TValid:       -2,
					TLease:       time.Now().Add(-1 * time.Second),
					Version:      "2",
				})

				_, err := conn.ConditionalUpdate(key, item, false)
				if err == nil {
					globalID = id
					resCh <- true
				} else {
					resCh <- false
				}
			}(i)
		}

		successCnt := 0
		for i := 0; i < currentNum; i++ {
			if <-resCh {
				successCnt++
			}
		}
		assert.Equal(t, 1, successCnt)

		got, _ := conn.GetItem(key)
		assert.Equal(t, strconv.Itoa(globalID), got.GroupKeyList())
	})

	// create concurrently
	t.Run("create-in-parallel", func(t *testing.T) {
		key := "test_key_concurrent_create"
		conn.Delete(key)

		resCh := make(chan bool)
		currentNum := 100
		globalID := 0

		for i := 1; i <= currentNum; i++ {
			go func(id int) {
				p := testutil.NewDefaultPerson()
				p.Name = "newer"
				item := h.MakeItem(txn.ItemOptions{
					Key:          key,
					Value:        p,
					GroupKeyList: strconv.Itoa(id),
					TxnState:     config.COMMITTED,
					TValid:       -2,
					TLease:       time.Now().Add(-1 * time.Second),
					Version:      "2",
				})

				_, err := conn.ConditionalUpdate(key, item, true)
				if err == nil {
					globalID = id
					resCh <- true
				} else {
					resCh <- false
				}
			}(i)
		}

		successCnt := 0
		for i := 0; i < currentNum; i++ {
			if <-resCh {
				successCnt++
			}
		}
		assert.Equal(t, 1, successCnt)

		got, _ := conn.GetItem(key)
		assert.Equal(t, strconv.Itoa(globalID), got.GroupKeyList())
	})
}

func testConnectionPutAndGet(t *testing.T, conn txn.Connector, h Helper) {
	key := "test_key_put_get_raw"
	conn.Delete(key)

	se := serializer.NewJSONSerializer()
	item := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        testutil.NewDefaultPerson(),
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -3,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "2",
	})

	bs, _ := se.Serialize(item)
	assert.NoError(t, conn.Put(key, bs))

	str, _ := conn.Get(key)
	got := h.NewInstance()
	_ = se.Deserialize([]byte(str), &got)
	assert.True(t, got.Equal(item))
}

func testConnectionReplaceAndGet(t *testing.T, conn txn.Connector, h Helper) {
	key := "test_key_replace_get_raw"
	conn.Delete(key)

	se := serializer.NewJSONSerializer()
	item := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        testutil.NewDefaultPerson(),
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -3,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "2",
	})

	bs, _ := se.Serialize(item)
	_ = conn.Put(key, bs)

	item.SetVersion(util.AddToString(item.Version(), 1))
	bs, _ = se.Serialize(item)
	_ = conn.Put(key, bs)

	str, _ := conn.Get(key)
	got := h.NewInstance()
	_ = se.Deserialize([]byte(str), &got)
	assert.True(t, got.Equal(item))
}

func testConnectionGetNoExist(t *testing.T, conn txn.Connector, _ Helper) {
	key := "test_key_get_no_exist"
	conn.Delete(key)
	_, err := conn.Get(key)
	assert.EqualError(t, err, "key not found")
}

func testConnectionPutDirectItem(t *testing.T, conn txn.Connector, h Helper) {
	key := "test_key_put_direct"
	conn.Delete(key)

	item := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        testutil.NewDefaultPerson(),
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -3,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "2",
	})

	_, err := conn.PutItem(key, item)
	assert.NoError(t, err)

	got, _ := conn.GetItem(key)
	fmt.Println("Retrieved item:", got)

	// got := h.NewInstance()
	// _ = json.Unmarshal([]byte(str), &got)
	assert.Truef(t, got.Equal(item), "expected %+v, got %+v", item, got)
}

func testConnectionDeleteTwice(t *testing.T, conn txn.Connector, _ Helper) {
	key := "test_key_delete_twice"
	_ = conn.Put(key, "test_value")
	assert.NoError(t, conn.Delete(key))
	assert.NoError(t, conn.Delete(key))
}

func testConnectionConditionalUpdateDoCreate(t *testing.T, conn txn.Connector, h Helper) {
	dbItem := h.MakeItem(txn.ItemOptions{
		Key:          "item1",
		Value:        testutil.NewTestItem("item1-db"),
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -3,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "1",
	})

	cacheItem := h.MakeItem(txn.ItemOptions{
		Key:          "item1",
		Value:        testutil.NewTestItem("item1-cache"),
		GroupKeyList: "2",
		TxnState:     config.COMMITTED,
		TValid:       -2,
		TLease:       time.Now().Add(-1 * time.Second),
		Prev:         util.ToJSONString(dbItem),
		Version:      "1",
	})

	t.Run("no item & doCreate true", func(t *testing.T) {
		conn.Delete(cacheItem.Key())
		_, err := conn.ConditionalUpdate(cacheItem.Key(), cacheItem, true)
		assert.NoError(t, err)
	})

	t.Run("has item & doCreate true", func(t *testing.T) {
		conn.PutItem(dbItem.Key(), dbItem)
		_, err := conn.ConditionalUpdate(cacheItem.Key(), cacheItem, true)
		assert.EqualError(t, err, txn.VersionMismatch.Error())
	})

	t.Run("no item & doCreate false", func(t *testing.T) {
		conn.Delete(cacheItem.Key())
		_, err := conn.ConditionalUpdate(cacheItem.Key(), cacheItem, false)
		assert.EqualError(t, err, txn.VersionMismatch.Error())
	})

	t.Run("has item & doCreate false", func(t *testing.T) {
		conn.PutItem(dbItem.Key(), dbItem)
		_, err := conn.ConditionalUpdate(cacheItem.Key(), cacheItem, false)
		assert.NoError(t, err)
	})
}

func testConnectionConditionalCommit(t *testing.T, conn txn.Connector, h Helper) {
	dbItem := h.MakeItem(txn.ItemOptions{
		Key:          "item1",
		Value:        testutil.NewTestItem("item1-db"),
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -3,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "1",
	})

	conn.PutItem(dbItem.Key(), dbItem)

	_, err := conn.ConditionalCommit(dbItem.Key(), dbItem.Version(), 100)
	assert.NoError(t, err)

	item, _ := conn.GetItem(dbItem.Key())
	dbItem.SetVersion(util.AddToString(dbItem.Version(), 1))

	dbItem.SetTxnState(config.COMMITTED)
	dbItem.SetTValid(100)
	assert.True(t, dbItem.Equal(item))
}

// ---------------------------------------------------------------------------
// unified entry – reuse a single connection and helper across all tests
// ---------------------------------------------------------------------------

func TestConnectorSuite(t *testing.T, conn txn.Connector, h Helper) {
	tests := []struct {
		name string
		fn   func(*testing.T, txn.Connector, Helper)
	}{
		{"Connection_Connect", testConnection_Connect},
		{"Timestamp", testTimestamp},
		{"Connection_GetItemNotFound", testConnection_GetItemNotFound},
		{"ConnectionPutItemAndGetItem", testConnectionPutItemAndGetItem},
		{"ConnectionReplaceAndGetItem", testConnectionReplaceAndGetItem},
		{"ConnectionConditionalUpdateSuccess", testConnectionConditionalUpdateSuccess},
		{"ConnectionConditionalUpdateFail", testConnectionConditionalUpdateFail},
		{"ConnectionConditionalUpdateNonExist", testConnectionConditionalUpdateNonExist},
		{"ConnectionConditionalUpdateConcurrently", testConnectionConditionalUpdateConcurrently},
		{"ConnectionPutAndGet", testConnectionPutAndGet},
		{"ConnectionReplaceAndGet", testConnectionReplaceAndGet},
		{"ConnectionGetNoExist", testConnectionGetNoExist},
		{"ConnectionPutDirectItem", testConnectionPutDirectItem},
		{"ConnectionDeleteTwice", testConnectionDeleteTwice},
		{"ConnectionConditionalUpdateDoCreate", testConnectionConditionalUpdateDoCreate},
		{"ConnectionConditionalCommit", testConnectionConditionalCommit},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			tc.fn(t, conn, h)
		})
	}
}
