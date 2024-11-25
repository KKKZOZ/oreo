package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
)

const (
	keyspace = "oreo"
)

// CQL statements
var createStatements = []string{
	`CREATE KEYSPACE IF NOT EXISTS oreo
     WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`,

	`CREATE TABLE IF NOT EXISTS oreo.items (
        key text,
        value text,
        group_key_list text,
        txn_state int,
        t_valid bigint,
        t_lease timestamp,
        prev text,
        linked_len int,
        is_deleted boolean,
        version text,
        PRIMARY KEY (key)
    ) WITH gc_grace_seconds = 172800`,

	`CREATE TABLE IF NOT EXISTS oreo.kv(
        key text,
        value text,
        PRIMARY KEY ( key )
    )`,
}

var truncateStatements = []string{
	"TRUNCATE oreo.items",
	"TRUNCATE oreo.kv",
}

var op = ""
var ip = ""

func main() {
	// Parse command line arguments
	flag.StringVar(&op, "op", "", "Operation to perform: create or clear")
	flag.StringVar(&ip, "ip", "127.0.0.1", "IP address of Cassandra node")
	flag.Parse()

	if op == "" {
		log.Fatal("Please specify an operation using -op flag")
	}

	// Initialize cluster
	cluster := gocql.NewCluster(ip)
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = time.Second * 5
	cluster.ProtoVersion = 4

	// For initial connection, use system keyspace
	cluster.Keyspace = "system"
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal("Error creating session:", err)
	}
	defer session.Close()

	switch op {
	case "create":
		// Execute create statements
		for _, stmt := range createStatements {
			if err := session.Query(stmt).Exec(); err != nil {
				log.Fatalf("Error executing create statement: %v\nStatement: %s", err, stmt)
			}
		}
		fmt.Println("Tables created successfully")

	case "clear":
		// Switch to oreo keyspace for truncate operations
		cluster.Keyspace = keyspace
		session.Close()
		session, err = cluster.CreateSession()
		if err != nil {
			log.Fatal("Error creating session with oreo keyspace:", err)
		}

		// Execute truncate statements
		for _, stmt := range truncateStatements {
			if err := session.Query(stmt).Exec(); err != nil {
				log.Fatalf("Error executing truncate statement: %v\nStatement: %s", err, stmt)
			}
		}
		fmt.Println("Tables truncated successfully")

	default:
		log.Fatal("Invalid operation. Use 'create' or 'clear'")
	}
}
