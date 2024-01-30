package ycsb

type WorkloadParameter struct {
	TableName string

	RecordCount int

	OperationCount int

	ReadProportion   float64
	UpdateProportion float64
	InsertProportion float64
	ScanProportion   float64

	ReadModifyWriteProportion float64
}

// ----------------------------------------------------------------------------
