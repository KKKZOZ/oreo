package workload

import (
	"benchmark/pkg/generator"
	"benchmark/pkg/util"
	"benchmark/ycsb"
	"context"
	"fmt"
	"sync"
)

type SocialWorkload struct {
	mu sync.Mutex

	Randomizer
	taskChooser             *generator.Discrete
	wp                      *WorkloadParameter
	MongoDBNamespace        string
	CassandraNamespace      string
	RedisNamespaceAnalytics string
	RedisNamespaceSession   string
	task1Count              int
	task2Count              int
	task3Count              int
}

var _ Workload = (*SocialWorkload)(nil)

func NewSocialWorkload(wp *WorkloadParameter) *SocialWorkload {
	return &SocialWorkload{
		mu:                      sync.Mutex{},
		Randomizer:              *NewRandomizer(wp),
		taskChooser:             createTaskGenerator(wp),
		wp:                      wp,
		MongoDBNamespace:        "users",
		CassandraNamespace:      "posts",
		RedisNamespaceAnalytics: "analytics",
		RedisNamespaceSession:   "session",
	}
}

func (wl *SocialWorkload) ContentFeedRetrieval(ctx context.Context, db ycsb.TransactionDB) {

	db.Start()

	post_id := wl.NextKeyName()
	_, _ = db.Read(ctx, "Cassandra", fmt.Sprintf("%v:%v", wl.CassandraNamespace, post_id))
	_, _ = db.Read(ctx, "Redis", fmt.Sprintf("%v:%v", wl.RedisNamespaceAnalytics, post_id))

	db.Commit()
}

func (wl *SocialWorkload) ContentCreation(ctx context.Context, db ycsb.TransactionDB) {

	db.Start()

	post_id := wl.NextKeyName()
	db.Update(ctx, "Cassandra", fmt.Sprintf("%v:%v", wl.CassandraNamespace, post_id), wl.RandomValue())
	db.Update(ctx, "Redis", fmt.Sprintf("%v:%v", wl.RedisNamespaceAnalytics, post_id), wl.RandomValue())
	db.Update(ctx, "Redis", fmt.Sprintf("%v:%v", wl.RedisNamespaceSession, post_id), wl.RandomValue())

	db.Commit()
}

func (wl *SocialWorkload) ProfileUpdate(ctx context.Context, db ycsb.TransactionDB) {

	db.Start()
	user_id := wl.NextKeyName()
	db.Update(ctx, "MongoDB", fmt.Sprintf("%v:%v", wl.MongoDBNamespace, user_id), wl.RandomValue())
	db.Update(ctx, "Redis", fmt.Sprintf("%v:%v", wl.RedisNamespaceSession, user_id), wl.RandomValue())
	db.Commit()
}

func (wl *SocialWorkload) Load(ctx context.Context, opCount int,
	db ycsb.DB) {
	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		fmt.Println("The DB does not support transactions")
		return
	}
	txnDB.Start()
	for i := 0; i < opCount; i++ {
		key := wl.NextKeyNameFromSequence()
		txnDB.Insert(ctx, "MongoDB", fmt.Sprintf("%v:%v", wl.MongoDBNamespace, key), wl.RandomValue())
		txnDB.Insert(ctx, "Cassandra", fmt.Sprintf("%v:%v", wl.CassandraNamespace, key), wl.RandomValue())
		txnDB.Insert(ctx, "Redis", fmt.Sprintf("%v:%v", wl.RedisNamespaceAnalytics, key), wl.RandomValue())
	}
	err := txnDB.Commit()
	if err != nil {
		fmt.Printf("Error when committing transaction: %v\n", err)
	}

}

func (wl *SocialWorkload) Run(ctx context.Context, opCount int,
	db ycsb.DB) {

	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		fmt.Println("The DB does not support transactions")
		return
	}
	for i := 0; i < opCount; i++ {
		switch wl.NextTask() {
		case 1:
			wl.ContentFeedRetrieval(ctx, txnDB)
		case 2:
			wl.ContentCreation(ctx, txnDB)
		case 3:
			wl.ProfileUpdate(ctx, txnDB)
		default:
			panic("Invalid task")
		}
	}

}

func (wl *SocialWorkload) Cleanup() {}

func (wl *SocialWorkload) NeedPostCheck() bool {
	return true
}

func (wl *SocialWorkload) NeedRawDB() bool {
	return false
}

func (wl *SocialWorkload) PostCheck(context.Context, ycsb.DB, chan int) {
}

func (wl *SocialWorkload) DisplayCheckResult() {
	fmt.Printf("Task 1 count: %v\n", wl.task1Count)
	fmt.Printf("Task 2 count: %v\n", wl.task2Count)
	fmt.Printf("Task 3 count: %v\n", wl.task3Count)
}

func (wl *SocialWorkload) NextTask() int64 {
	wl.mu.Lock()
	defer wl.mu.Unlock()
	idx := wl.taskChooser.Next(wl.r)
	switch idx {
	case 1:
		wl.task1Count++
	case 2:
		wl.task2Count++
	case 3:
		wl.task3Count++
	default:
		panic("Invalid task")
	}
	return idx
}

func (wl *SocialWorkload) RandomValue() string {
	value := wl.r.Intn(10000)
	return util.ToString(value)
}
