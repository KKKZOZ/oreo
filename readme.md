# Oreo - An Easy-to-use Distributed Transaciotn Framework

<div align="center">

![Logo](./assets/img/logo.png)
</div>


Oreo is an open source implementation of Cherry Garcia protocol written in Go, it aims to provide an easy-to-use interface for enabling efficient multi-item transactions across heterogeneous key-value store.


It supports:

+ A simple MemoryDatabase(for test only)
+ Redis
+ Kvrocks
+ MongoDB
+ CouchDB

It provides:

+ Unified `Read(key string)`, `Write(key string, value any)`, `Delete(key string)` interfaces
+ ACID distributed transactions
  + Snapshot Isolation
  + Serializable

> ***It is currently alpha***.

## Main Features

+ An easy-to-use interface for distributed transactions.
+ Support heterogeneous key-value store.
+ Clear architecture.



## Project Structure

![Project Structure](./assets/img/project_structure.png)

## Getting Started

> Full example is in `examples/oreo_basic_with_memory_datastore`

```go
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
	// write to memory datastore 1
	txn.Write("mem1", "user1", user1)
	// write to memory datastore 2
	txn.Write("mem2", "user2", user2)
	err := txn.Commit()
	if err != nil {
		panic(err)
	}
	fmt.Println("inserted two users")

}
```

## Why this name?

I love Oreo and Oreo also has many different flavors of fillings!

## About Cherry Garicia

A client coordinated transaction protocol to enable efficient multi-item transactions across heterogeneous key-value store.

## Roadmap

- [ ] Support MongoDB
- [ ] Support CouchDB