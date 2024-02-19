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

	DataConsistencyTest   bool
	InitialAmountPerKey   int
	TransferAmountPerTxn  int
	TotalAmount           int
	PostCheckWorkerThread int

	TxnPerformanceTest bool
}

// ----------------------------------------------------------------------------
