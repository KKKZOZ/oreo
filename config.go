package main

type DataStoreType string

const (
	MEMORY DataStoreType = "memory"
)

var globalDataStore Datastore

const leastTime = 1000
