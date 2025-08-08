package cassandra

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/gocql/gocql"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/txn"
)

var _ txn.Connector = (*CassandraConnection)(nil)

type CassandraConnection struct {
	session      *gocql.Session
	config       ConnectionOptions
	hasConnected bool
}

type ConnectionOptions struct {
	Hosts    []string
	Keyspace string
	Username string
	Password string
}

var defaultOptions = ConnectionOptions{
	Hosts:    []string{"localhost"},
	Keyspace: "oreo",
}

// NewCassandraConnection creates a new Cassandra connection.
func NewCassandraConnection(config *ConnectionOptions) *CassandraConnection {
	// Start with a copy of the default configuration.
	finalConfig := defaultOptions

	// If the user provided a configuration, layer their values on top.
	if config != nil {
		if len(config.Hosts) > 0 {
			finalConfig.Hosts = config.Hosts
		}
		if config.Keyspace != "" {
			finalConfig.Keyspace = config.Keyspace
		}
		if config.Username != "" {
			finalConfig.Username = config.Username
		}
		if config.Password != "" {
			finalConfig.Password = config.Password
		}
	}

	return &CassandraConnection{
		config:       finalConfig,
		hasConnected: false,
	}
}

// Connect establishes a connection to the Cassandra cluster.
func (c *CassandraConnection) Connect() error {
	if c.hasConnected {
		return nil
	}

	cluster := gocql.NewCluster(c.config.Hosts...)
	cluster.Keyspace = c.config.Keyspace
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4

	if c.config.Username != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: c.config.Username,
			Password: c.config.Password,
		}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return err
	}

	c.session = session
	c.hasConnected = true

	// 预热连接池
	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			_ = c.session.Query("SELECT key FROM items WHERE key = ? LIMIT 1", "test").Exec()
		}()
	}
	wg.Wait()

	return nil
}

// GetItem retrieves a structured transaction item from Cassandra.
// It returns a txn.DataItem, which represents a full row with transaction metadata.
func (c *CassandraConnection) GetItem(key string) (txn.DataItem, error) {
	if !c.hasConnected {
		return &CassandraItem{}, fmt.Errorf("not connected to Cassandra")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	var item CassandraItem
	err := c.session.Query(`
        SELECT key, value, group_key_list, txn_state, t_valid, t_lease, prev, linked_len, is_deleted, version 
        FROM items WHERE key = ?`, key).Scan(
		&item.CKey, &item.CValue, &item.CGroupKeyList, &item.CTxnState,
		&item.CTValid, &item.CTLease, &item.CPrev, &item.CLinkedLen,
		&item.CIsDeleted, &item.CVersion)

	if err == gocql.ErrNotFound {
		return &CassandraItem{}, errors.New(txn.KeyNotFound)
	}
	if err != nil {
		return &CassandraItem{}, errors.New("version mismatch")
	}
	return &item, nil
}

// PutItem inserts or updates a transaction item in Cassandra.
func (c *CassandraConnection) PutItem(key string, value txn.DataItem) (string, error) {
	if !c.hasConnected {
		return "", fmt.Errorf("not connected to Cassandra")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	item, ok := value.(*CassandraItem)
	if !ok {
		return "", fmt.Errorf("invalid item type")
	}

	err := c.session.Query(`
        INSERT INTO items (key, value, group_key_list, txn_state, t_valid, t_lease, prev, linked_len, is_deleted, version)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		key, item.CValue, item.CGroupKeyList, item.CTxnState,
		item.CTValid, item.CTLease, item.CPrev, item.CLinkedLen,
		item.CIsDeleted, item.CVersion).Exec()
	if err != nil {
		return "", errors.New(fmt.Sprintf("PutItem key %s failed, err: %v", key, err))
	}
	return "", nil
}

// ConditionalUpdate atomically updates an item if the version matches using a lightweight transaction.
// If doCreate is true, it will create the item if it does not exist.
func (c *CassandraConnection) ConditionalUpdate(
	key string,
	value txn.DataItem,
	doCreate bool,
) (string, error) {
	if !c.hasConnected {
		return "", fmt.Errorf("not connected to Cassandra")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	newVer := util.AddToString(value.Version(), 1)

	if doCreate {
		if value.Version() != "" {
			return "", errors.New(txn.VersionMismatch)
		}

		// 使用 Cassandra 的轻量级事务(LWT)确保原子性
		applied, err := c.session.Query(`
            INSERT INTO items (key, value, group_key_list, txn_state, t_valid, t_lease, prev, linked_len, is_deleted, version)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            IF NOT EXISTS`,
			key, value.Value(), value.GroupKeyList(), value.TxnState(),
			value.TValid(), value.TLease(), value.Prev(), value.LinkedLen(),
			value.IsDeleted(), newVer).ScanCAS()
		if err != nil {
			return "", errors.New(
				fmt.Sprintf("ConditionalUpdate(doCreate) key %s failed, err: %v", key, err),
			)
		}
		if !applied {
			return "", errors.New("key exists")
		}
		return newVer, nil
	}

	// 更新现有记录，使用 LWT 确保版本匹配
	applied, err := c.session.Query(`
        UPDATE items 
        SET value = ?, group_key_list = ?, txn_state = ?, t_valid = ?, 
            t_lease = ?, prev = ?, linked_len = ?, is_deleted = ?, version = ?
        WHERE key = ?
        IF version = ?`,
		value.Value(), value.GroupKeyList(), value.TxnState(), value.TValid(),
		value.TLease(), value.Prev(), value.LinkedLen(), value.IsDeleted(),
		newVer, key, value.Version()).ScanCAS()
	// gocql: not enough columns to scan into: have 1 want 2
	// this is ok because it only occurs when the conditional update fails
	if err != nil {
		return "", errors.New(fmt.Sprintf("ConditionalUpdate key %s failed, err: %v", key, err))
	}
	if !applied {
		return "", errors.New(txn.VersionMismatch)
	}

	return newVer, nil
}

// ConditionalCommit atomically commits a transaction if the version matches.
func (c *CassandraConnection) ConditionalCommit(
	key string,
	version string,
	tCommit int64,
) (string, error) {
	if !c.hasConnected {
		return "", fmt.Errorf("not connected to Cassandra")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	applied, err := c.session.Query(`
        UPDATE items 
        SET txn_state = ?, t_valid = ?
        WHERE key = ?
        IF version = ?`,
		config.COMMITTED, tCommit, key, version).ScanCAS()
	if err != nil {
		return "", errors.New(fmt.Sprintf("ConditionalCommit key %s failed, err: %v", key, err))
	}
	if !applied {
		return "", errors.New(txn.VersionMismatch)
	}

	return version, nil
}

// AtomicCreate creates a key-value pair in the 'kv' table if the key does not already exist.
func (c *CassandraConnection) AtomicCreate(name string, value any) (string, error) {
	if !c.hasConnected {
		return "", fmt.Errorf("not connected to Cassandra")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	strValue := util.ToString(value)
	applied, err := c.session.Query(`
        INSERT INTO kv (key, value)
        VALUES (?, ?)
        IF NOT EXISTS`,
		name, strValue).ScanCAS()
	if err != nil {
		return "", err
	}
	if !applied {
		var existingValue string
		err = c.session.Query(`SELECT value FROM kv WHERE key = ?`, name).Scan(&existingValue)
		if err != nil {
			return "", errors.New(fmt.Sprintf("get key %s failed, err: %v", name, err))
		}
		return existingValue, errors.New(txn.KeyExists)
	}

	return "", nil
}

// Get retrieves a simple string value from the 'kv' table.
// This is a general-purpose getter, distinct from GetItem, which retrieves a structured txn.DataItem.
func (c *CassandraConnection) Get(name string) (string, error) {
	if !c.hasConnected {
		return "", fmt.Errorf("not connected to Cassandra")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	var value string
	err := c.session.Query(`SELECT value FROM kv WHERE key = ?`, name).Scan(&value)
	if err == gocql.ErrNotFound {
		return "", errors.New(txn.KeyNotFound)
	}
	if err != nil {
		return "", errors.New(fmt.Sprintf("get key %s failed, err: %v", name, err))
	}
	return value, nil
}

// Put sets the value for a given key in the 'kv' table.
func (c *CassandraConnection) Put(name string, value interface{}) error {
	if !c.hasConnected {
		return fmt.Errorf("not connected to Cassandra")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	strValue := util.ToString(value)
	err := c.session.Query(`
        INSERT INTO kv (key, value)
        VALUES (?, ?)`,
		name, strValue).Exec()
	if err != nil {
		return errors.New(fmt.Sprintf("put key %s failed, err: %v", name, err))
	}
	return nil
}

// Delete removes a key-value pair from the 'kv' table.
func (c *CassandraConnection) Delete(name string) error {
	if !c.hasConnected {
		return fmt.Errorf("not connected to Cassandra")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	err := c.session.Query(`DELETE FROM kv WHERE key = ?`, name).Exec()
	if err != nil {
		return errors.New(fmt.Sprintf("delete key %s failed, err: %v", name, err))
	}
	return nil
}
