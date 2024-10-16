package main

import (
	"fmt"
	"time"

	_ "github.com/go-kivik/kivik/v4/couchdb" // The CouchDB driver
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/couchdb"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

func main() {

	couchItem := couchdb.NewCouchDBItem(txn.ItemOptions{
		Key:     "item1",
		Value:   "value1",
		TxnId:   "txn1",
		TValid:  time.Now().UnixMicro(),
		TLease:  time.Now(),
		Version: "1",
	})

	conn := couchdb.NewCouchDBConnection(nil)
	err := conn.Connect()
	if err != nil {
		panic(err)
	}

	_, err = conn.PutItem(couchItem.CKey, couchItem)
	if err != nil {
		panic(err)
	}

	resItem, err := conn.GetItem(couchItem.CKey)
	if err != nil {
		panic(err)
	}
	fmt.Println(resItem)

}
