package workload

import (
	"benchmark/ycsb"
	"context"
	"errors"
)

var _ ycsb.DB = (*DBSet)(nil)

var _ ycsb.DBCreator = (*DBSetCreator)(nil)

type DBSet struct {
	dbMap map[string]ycsb.DB
}

type DBSetCreator struct {
	CreatorMap map[string]ycsb.DBCreator
}

func (dc *DBSetCreator) Create() (ycsb.DB, error) {
	dbMap := make(map[string]ycsb.DB)
	for dbName, creator := range dc.CreatorMap {
		db, err := creator.Create()
		if err != nil {
			return nil, err
		}
		dbMap[dbName] = db
	}
	return NewDBSet(dbMap), nil
}

func NewDBSet(dbMap map[string]ycsb.DB) *DBSet {
	return &DBSet{
		dbMap: dbMap,
	}
}

func (m *DBSet) Close() error {
	return nil
}

func (m *DBSet) InitThread(context context.Context, threadID int, threadCount int) context.Context {
	return context
}

func (m *DBSet) CleanupThread(context context.Context) {
}

func (m *DBSet) Read(context context.Context, table string, key string) (string, error) {
	if db, ok := m.dbMap[table]; ok {
		return db.Read(context, table, key)
	} else {
		return "", errors.New("table not found")
	}

}

func (m *DBSet) Update(context context.Context, table string, key string, value string) error {
	if db, ok := m.dbMap[table]; ok {
		return db.Update(context, table, key, value)
	} else {
		return errors.New("table not found")
	}
}

func (m *DBSet) Insert(context context.Context, table string, key string, value string) error {
	if db, ok := m.dbMap[table]; ok {
		return db.Insert(context, table, key, value)
	} else {
		return errors.New("table not found")
	}
}

func (m *DBSet) Delete(context context.Context, table string, key string) error {
	if db, ok := m.dbMap[table]; ok {
		return db.Delete(context, table, key)
	} else {
		return errors.New("table not found")
	}
}
