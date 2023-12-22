package main

import (
	"fmt"
	"time"

	"github.com/kkkzoz/oreo/datastore/memory"
	"github.com/kkkzoz/oreo/txn"
)

type User struct {
	Username string
	Password string
	Email    string
}

func insertTwoUsers(txn *txn.Transaction) {
	user1 := User{
		Username: "user1",
		Password: "password1",
		Email:    "user1@gmail.com",
	}
	user2 := User{
		Username: "user2",
		Password: "password2",
		Email:    "user2@gmail.com",
	}

	txn.Start()
	txn.Write("mem1", "user1", user1)
	txn.Write("mem2", "user2", user2)
	err := txn.Commit()
	if err != nil {
		panic(err)
	}
	fmt.Println("inserted two users")

}

func setupTransaction() *txn.Transaction {
	// create two new memory connection instances
	memConn1 := memory.NewMemoryConnection("localhost", 8321)
	memConn2 := memory.NewMemoryConnection("localhost", 8322)

	// create two datastore instances
	dst1 := memory.NewMemoryDatastore("mem1", memConn1)
	dst2 := memory.NewMemoryDatastore("mem2", memConn2)

	// create a new transaction
	txn := txn.NewTransaction()
	// add two datastores to the transaction
	txn.AddDatastore(dst1)
	txn.AddDatastore(dst2)
	// set one of them as global datastore
	txn.SetGlobalDatastore(dst1)

	return txn
}

func main() {

	// Create two new memory database instances
	memDB1 := memory.NewMemoryDatabase("localhost", 8321)
	go memDB1.Start()
	defer memDB1.Stop()
	memDB2 := memory.NewMemoryDatabase("localhost", 8322)
	go memDB2.Start()
	defer memDB2.Stop()
	time.Sleep(100 * time.Millisecond)

	// setup a transaction
	txn := setupTransaction()

	// execute some business logic
	insertTwoUsers(txn)

}
