package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-kivik/kivik/v4/couchdb" // The CouchDB driver
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/couchdb"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

func main() {
	// a random string
	key := randomString()
	couchItem := couchdb.NewCouchDBItem(txn.ItemOptions{
		Key:          key,
		Value:        "value1",
		GroupKeyList: "txn1",
		TValid:       time.Now().UnixMicro(),
		TLease:       time.Now(),
	})

	conn := couchdb.NewCouchDBConnection(nil)
	err := conn.Connect()
	if err != nil {
		panic(err)
	}

	rev, err := conn.PutItem(couchItem.CKey, couchItem)
	if err != nil {
		panic(err)
	}
	fmt.Printf("rev: %s\n", rev)

	resItem, err := conn.GetItem(couchItem.CKey)
	if err != nil {
		panic(err)
	}
	fmt.Println(resItem)

	resItem.SetValue("value2")
	rev, err = conn.PutItem(couchItem.CKey, resItem)
	if err != nil {
		panic(err)
	}
	resItem, err = conn.GetItem(couchItem.CKey)
	if err != nil {
		panic(err)
	}

	fmt.Printf("rev: %s\n", rev)

	num := 100
	var wg sync.WaitGroup
	wg.Add(num)
	successNum := 0
	var mu sync.Mutex

	for i := 0; i < num; i++ {
		item := resItem
		go func(ii int) {
			defer wg.Done()
			idx := fmt.Sprintf("%d", ii)
			item.SetValue(idx)
			rev, err := conn.ConditionalUpdate(item.Key(), item, false)
			if err == nil {
				mu.Lock()
				successNum++
				fmt.Printf("rev: %s\n", rev)
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("successNum: %d\n", successNum)
}

func randomString() string {
	return "item" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
