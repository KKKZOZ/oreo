# Oreo - An Easy-to-use Distributed Transaction Framework

[![Go Report Card](https://goreportcard.com/badge/github.com/kkkzoz/oreo)](https://goreportcard.com/report/github.com/kkkzoz/oreo)
[![Go Reference](https://pkg.go.dev/badge/github.com/kkkzoz/oreo.svg)](https://pkg.go.dev/github.com/kkkzoz/oreo)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


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

> ***It is currently beta***.

## Main Features

+ An easy-to-use interface for distributed transactions.
+ Support heterogeneous key-value store.
+ Clean architecture.



## Project Structure

![Project Structure](./assets/img/project_structure.png)

## Getting Started

### Installation

Oreo supports 2 last Go versions and requires a Go version with
[modules](https://github.com/golang/go/wiki/Modules) support. So make sure to initialize a Go module:

```shell
go mod init github.com/my/repo
```

Then install oreo:

```shell
go get github.com/kkkzoz/oreo
```


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

## Roadmap

Can be found [here](https://trello.com/b/Vl2H7Aqg/oreo-roadmap)

## Why this name?

I love Oreo and Oreo also has many different flavors of fillings!

## About Cherry Garicia

A client coordinated transaction protocol to enable efficient multi-item transactions across heterogeneous key-value store.

## License
This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.