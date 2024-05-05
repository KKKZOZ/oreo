package workload

type WorkloadParameter struct {
	DBName       string
	TableName    string
	WorkloadName string

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
	DoubleSeqCommitProportion float64

	// These parameters are for the data consistency test
	InitialAmountPerKey   int
	TransferAmountPerTxn  int
	TotalAmount           int
	PostCheckWorkerThread int

	// These parameters are for the data distribution test
	GlobalDatastoreName string
	RedisProportion     float64
	MongoProportion     float64
}

// ----------------------------------------------------------------------------
