## Introduction

There are four kinds benchmark available:

+ YCSB benchmark
+ Data Consistency Test
+ Transaction Performance Test
+ Across Datastore Test

Running a benchmark includes two phase:

1. Load data to datastore
2. Run the benchmark

You are expected to run the two commands one after another.


## YCSB benchmark

Change the corresponding parameters in `main.go`:

```go
wp := &ycsb.WorkloadParameter{
		RecordCount:               100,
		OperationCount:            1,
		TxnOperationGroup:         10,
		ReadProportion:            0.5,
		UpdateProportion:          0.5,
		InsertProportion:          0,
		ScanProportion:            0,
		ReadModifyWriteProportion: 0,
	}
```

Then run the two commands below:

```bash
# load data
go run main.go redis load 2

# run benchmark
go run main.go redis run 2
```

## Start a benchmark

```bash
# load data
# 2 indicates 2 threads
go run main.go redis load 2
# test redids
# 2 indicates 2 threads
go run main.go redis run 2

# Data consistency test
go run main.go redis load 2 -dc
```