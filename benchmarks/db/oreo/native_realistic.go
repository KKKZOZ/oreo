package oreo

import (
	"benchmark/ycsb"
	"context"
)

var _ ycsb.DBCreator = (*OreoRedisCreator)(nil)

type NativeRealisticCreator struct {
	ConnMap map[string]ycsb.DB
}

func (nrc *NativeRealisticCreator) Create() (ycsb.DB, error) {
	return NewNativeRealisticDatastore(nrc.ConnMap), nil
}

var _ ycsb.DB = (*NativeRealisticDatastore)(nil)

var _ ycsb.TransactionDB = (*NativeRealisticDatastore)(nil)

type NativeRealisticDatastore struct {
	connMap map[string]ycsb.DB
}

func NewNativeRealisticDatastore(connMap map[string]ycsb.DB) *NativeRealisticDatastore {

	return &NativeRealisticDatastore{
		connMap: connMap,
	}
}

func (r *NativeRealisticDatastore) Start() error {
	return nil
}

func (r *NativeRealisticDatastore) Commit() error {
	return nil
}

func (r *NativeRealisticDatastore) Abort() error {
	return nil
}

func (r *NativeRealisticDatastore) NewTransaction() ycsb.TransactionDB {
	panic("implement me")
}

func (r *NativeRealisticDatastore) Close() error {
	return nil
}

func (r *NativeRealisticDatastore) InitThread(ctx context.Context, threadID int, threadCount int) context.Context {
	return ctx
}

func (r *NativeRealisticDatastore) CleanupThread(ctx context.Context) {
}

func (r *NativeRealisticDatastore) Read(ctx context.Context, table string, key string) (string, error) {
	key = addNativePrefix(key)

	value, err := r.connMap[table].Read(ctx, "", key)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r *NativeRealisticDatastore) Update(ctx context.Context, table string, key string, value string) error {
	key = addNativePrefix(key)
	return r.connMap[table].Update(ctx, "", key, value)
}

func (r *NativeRealisticDatastore) Insert(ctx context.Context, table string, key string, value string) error {
	key = addNativePrefix(key)
	return r.connMap[table].Insert(ctx, "", key, value)
}

func (r *NativeRealisticDatastore) Delete(ctx context.Context, table string, key string) error {
	key = addNativePrefix(key)
	return r.connMap[table].Delete(ctx, "", key)
}

func addNativePrefix(key string) string {
	return "native-" + key
}
