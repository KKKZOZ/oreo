package ycsb

type WorkloadParameter struct {
	DBName    string
	TableName string

	ThreadCount int
	DoBenchmark bool

	RecordCount       int
	OperationCount    int
	TxnOperationGroup int

	ReadProportion            float64
	UpdateProportion          float64
	InsertProportion          float64
	ScanProportion            float64
	ReadModifyWriteProportion float64

	// These parameters are for the data consistency test
	DataConsistencyTest   bool
	InitialAmountPerKey   int
	TransferAmountPerTxn  int
	TotalAmount           int
	PostCheckWorkerThread int

	TxnPerformanceTest bool

	// These parameters are for the data distribution test
	AcrossDatastoreTest bool
	GlobalDatastoreName string
	RedisProportion     float64
	MongoProportion     float64
}

// ----------------------------------------------------------------------------
