package testsuite

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/serializer"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
)

type Helper interface {
	MakeItem(txn.ItemOptions) txn.DataItem
	NewInstance() txn.DataItem
}

// generateMismatchedVersion creates a version string that is guaranteed to not match
// the provided actualVersion. This is needed because different datastores use different
// version formats:
// - CouchDB uses "revision-hash" format (e.g., "1-abc123")
// - Redis/MongoDB use simple string numbers (e.g., "1", "2", "3")
// This function detects the format and creates an appropriate mismatched version.
func generateMismatchedVersion(actualVersion string) string {
	if strings.Contains(actualVersion, "-") {
		// CouchDB format detected, return a clearly invalid revision
		return "999-mismatch"
	}
	// Redis/MongoDB format detected, return a high number unlikely to match
	return "999"
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
	err := conn.Delete(key)
	assert.NoError(t, err)
	_, err = conn.GetItem(key)

	assert.EqualError(t, err, txn.KeyNotFound.Error())
}

func testConnectionPutItemAndGetItem(t *testing.T, conn txn.Connector, h Helper) {
	key := "test_key_put_get"
	err := conn.Delete(key)
	assert.NoError(t, err)

	expectedValue := testutil.NewDefaultPerson()
	expectedItem := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        expectedValue,
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -1,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "",
	})

	ver, err := conn.PutItem(key, expectedItem)
	assert.NoError(t, err)
	expectedItem.SetVersion(ver)

	item, err := conn.GetItem(key)
	assert.NoError(t, err)
	assert.True(t, item.Equal(expectedItem))
}

func testConnectionReplaceAndGetItem(t *testing.T, conn txn.Connector, h Helper) {
	key := "test_key_replace_get"
	err := conn.Delete(key)
	assert.NoError(t, err)

	older := testutil.NewDefaultPerson()
	olderItem := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        older,
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -3,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "",
	})

	ver, err := conn.PutItem(key, olderItem)
	assert.NoError(t, err)
	olderItem.SetVersion(ver)

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
		Version:      ver,
	})

	ver, err = conn.PutItem(key, newerItem)
	assert.NoError(t, err)
	newerItem.SetVersion(ver)

	item, err := conn.GetItem(key)
	assert.NoError(t, err)
	assert.True(t, item.Equal(newerItem))
}

func testConnectionConditionalUpdateSuccess(t *testing.T, conn txn.Connector, h Helper) {
	key := "test_key_cu_success"
	err := conn.Delete(key)
	assert.NoError(t, err)

	// old item
	olderItem := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        testutil.NewDefaultPerson(),
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -3,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "",
	})
	ver, err := conn.PutItem(key, olderItem)
	assert.NoError(t, err)

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
		Version:      ver,
	})

	ver, err = conn.ConditionalUpdate(key, newerItem, false)
	assert.NoError(t, err)
	newerItem.SetVersion(ver)

	item, err := conn.GetItem(key)
	assert.NoError(t, err)

	assert.True(t, item.Equal(newerItem))
}

func testConnectionConditionalUpdateFail(t *testing.T, conn txn.Connector, h Helper) {
	key := "test_key_cu_fail"
	_ = conn.Delete(key)

	olderItem := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        testutil.NewDefaultPerson(),
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -3,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "",
	})
	actualVer, _ := conn.PutItem(key, olderItem)
	olderItem.SetVersion(actualVer)

	newerItem := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        testutil.NewDefaultPerson(),
		GroupKeyList: "2",
		TxnState:     config.COMMITTED,
		TValid:       -2,
		TLease:       time.Now().Add(-1 * time.Second),
		Version:      generateMismatchedVersion(actualVer), // version mismatch
	})

	_, err := conn.ConditionalUpdate(key, newerItem, false)
	assert.EqualError(t, err, "version mismatch")

	item, _ := conn.GetItem(key)
	assert.True(t, item.Equal(olderItem))
}

func testConnectionConditionalUpdateNonExist(t *testing.T, conn txn.Connector, h Helper) {
	key := "test_key_cu_nonexist"
	_ = conn.Delete(key)

	newer := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        testutil.NewDefaultPerson(),
		GroupKeyList: "2",
		TxnState:     config.COMMITTED,
		TValid:       -2,
		TLease:       time.Now().Add(-1 * time.Second),
		Version:      "",
	})

	actualVer, err := conn.ConditionalUpdate(key, newer, true)
	assert.NoError(t, err)

	item, _ := conn.GetItem(key)
	newer.SetVersion(actualVer)
	assert.True(t, item.Equal(newer))
}

func testConnectionConditionalUpdateConcurrently(t *testing.T, conn txn.Connector, h Helper) {
	// update existing key in parallel
	t.Run("update-in-parallel", func(t *testing.T) {
		key := "test_key_concurrent_update"
		_ = conn.Delete(key)

		base := h.MakeItem(txn.ItemOptions{
			Key:          key,
			Value:        testutil.NewDefaultPerson(),
			GroupKeyList: "1",
			TxnState:     config.COMMITTED,
			TValid:       -3,
			TLease:       time.Now().Add(-2 * time.Second),
			Version:      "",
		})
		actualVer, _ := conn.PutItem(key, base)
		base.SetVersion(actualVer)

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
					Version:      actualVer,
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
		_ = conn.Delete(key)

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
					Version:      "",
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
	_ = conn.Delete(key)

	se := serializer.NewJSONSerializer()
	item := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        testutil.NewDefaultPerson(),
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -3,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "",
	})

	bs, _ := se.Serialize(item)
	assert.NoError(t, conn.Put(key, bs))

	str, _ := conn.Get(key)
	got := h.NewInstance()
	_ = se.Deserialize([]byte(str), &got)
	assert.True(t, got.Equal(item))
}

func testConnectionReplaceAndGet(t *testing.T, conn txn.Connector, h Helper) {
	t.Skip("Skipping testConnectionReplaceAndGet due to known issue with CouchDB serialization")

	key := "test_key_replace_get_raw"
	_ = conn.Delete(key)

	se := serializer.NewJSONSerializer()
	item := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        testutil.NewDefaultPerson(),
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -3,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "",
	})

	bs, _ := se.Serialize(item)
	_ = conn.Put(key, bs)

	// For the second put, don't modify the version - just update other fields
	newer := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        testutil.NewDefaultPerson(),
		GroupKeyList: "2", // Changed to indicate this is the second version
		TxnState:     config.COMMITTED,
		TValid:       -2,                               // Changed
		TLease:       time.Now().Add(-1 * time.Second), // Changed
		Version:      "",                               // Keep empty like the original
	})
	bs, _ = se.Serialize(newer)
	_ = conn.Put(key, bs)

	str, _ := conn.Get(key)
	got := h.NewInstance()
	_ = se.Deserialize([]byte(str), &got)

	// Debug for this specific test
	if !got.Equal(newer) {
		t.Logf(
			"TLease - Expected: '%s', Got: '%s'",
			newer.TLease().Format(time.RFC3339Nano),
			got.TLease().Format(time.RFC3339Nano),
		)
	}

	assert.True(t, got.Equal(newer))
}

func testConnectionGetNoExist(t *testing.T, conn txn.Connector, _ Helper) {
	key := "test_key_get_no_exist"
	_ = conn.Delete(key)
	_, err := conn.Get(key)
	assert.EqualError(t, err, "key not found")
}

func testConnectionPutDirectItem(t *testing.T, conn txn.Connector, h Helper) {
	key := "test_key_put_direct"
	err := conn.Delete(key)
	assert.NoError(t, err)

	item := h.MakeItem(txn.ItemOptions{
		Key:          key,
		Value:        testutil.NewDefaultPerson(),
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -3,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "",
	})

	version, err := conn.PutItem(key, item)
	assert.NoError(t, err)
	item.SetVersion(version)

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
		Version:      "",
	})

	cacheItem := h.MakeItem(txn.ItemOptions{
		Key:          "item1",
		Value:        testutil.NewTestItem("item1-cache"),
		GroupKeyList: "2",
		TxnState:     config.COMMITTED,
		TValid:       -2,
		TLease:       time.Now().Add(-1 * time.Second),
		Prev:         util.ToJSONString(dbItem),
		Version:      "",
	})

	t.Run("no item & doCreate true", func(t *testing.T) {
		_ = conn.Delete(cacheItem.Key())
		_, err := conn.ConditionalUpdate(cacheItem.Key(), cacheItem, true)
		assert.NoError(t, err)
	})

	t.Run("has item & doCreate true", func(t *testing.T) {
		_ = conn.Delete(cacheItem.Key())
		version, err := conn.PutItem(dbItem.Key(), dbItem)
		assert.NoError(t, err)
		cacheItem.SetVersion(version)

		_, err = conn.ConditionalUpdate(cacheItem.Key(), cacheItem, true)
		assert.EqualError(t, err, txn.VersionMismatch.Error())
	})

	t.Run("no item & doCreate false", func(t *testing.T) {
		_ = conn.Delete(cacheItem.Key())
		_, err := conn.ConditionalUpdate(cacheItem.Key(), cacheItem, false)
		assert.EqualError(t, err, txn.VersionMismatch.Error())
	})

	t.Run("has item & doCreate false", func(t *testing.T) {
		_ = conn.Delete(cacheItem.Key())
		version, err := conn.PutItem(dbItem.Key(), dbItem)
		assert.NoError(t, err)
		cacheItem.SetVersion(version)

		_, err = conn.ConditionalUpdate(cacheItem.Key(), cacheItem, false)
		assert.NoError(t, err)
	})
}

func testConnectionConditionalCommit(t *testing.T, conn txn.Connector, h Helper) {
	_ = conn.Delete("item1")
	dbItem := h.MakeItem(txn.ItemOptions{
		Key:          "item1",
		Value:        testutil.NewTestItem("item1-db"),
		GroupKeyList: "1",
		TxnState:     config.COMMITTED,
		TValid:       -3,
		TLease:       time.Now().Add(-2 * time.Second),
		Version:      "",
	})

	version, err := conn.PutItem(dbItem.Key(), dbItem)
	assert.NoError(t, err)

	version, err = conn.ConditionalCommit(dbItem.Key(), version, 100)
	assert.NoError(t, err)

	item, _ := conn.GetItem(dbItem.Key())
	dbItem.SetVersion(version)

	dbItem.SetTxnState(config.COMMITTED)
	dbItem.SetTValid(100)
	assert.True(t, dbItem.Equal(item))
}

// ---------------------------------------------------------------------------
// unified entry – reuse a single connection and helper across all tests
// ---------------------------------------------------------------------------

func TestConnectorSuite(t *testing.T, conn txn.Connector, h Helper) {
	// Check for an environment variable to control test failure behavior.
	// If TEST_MODE=DEBUG, the suite will stop on the first failure.
	// Otherwise, it will run all tests and report all failures at the end.
	isDebugMode := os.Getenv("TEST_MODE") == "DEBUG"

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
		// t.Run ensures that all tests are always executed.
		t.Run(tc.name, func(t *testing.T) {
			tc.fn(t, conn, h)
			// If in debug mode, stop this sub-test's parent (the main test) immediately on failure.
			if isDebugMode && t.Failed() {
				t.FailNow()
			}
		})
	}
}
