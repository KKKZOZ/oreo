package workload

type WorkloadParameter struct {
	DBName       string
	TableName    string
	WorkloadName string

	ThreadCount       int
	DoBenchmark       bool
	PostCheckInterval int `yaml:"postcheckinterval"`

	RecordCount       int `yaml:"recordcount"`
	OperationCount    int `yaml:"operationcount"`
	TxnOperationGroup int `yaml:"txnoperationgroup"`

	ReadProportion            float64 `yaml:"readproportion"`
	UpdateProportion          float64 `yaml:"updateproportion"`
	InsertProportion          float64 `yaml:"insertproportion"`
	ReadModifyWriteProportion float64 `yaml:"readmodifywriteproportion"`
	ScanProportion            float64 `yaml:"scanproportion"`
	DoubleSeqCommitProportion float64 `yaml:"doubleseqcommitproportion"`

	// These parameters are for the data consistency test
	InitialAmountPerKey   int `yaml:"initialamountperkey"`
	TransferAmountPerTxn  int `yaml:"transferamountpertxn"`
	TotalAmount           int `yaml:"totalamount"`
	PostCheckWorkerThread int `yaml:"postcheckworkerthread"`

	// These parameters are for the data distribution test
	GlobalDatastoreName string  `yaml:"globaldatastorename"`
	Redis1Proportion    float64 `yaml:"redis1proportion"`
	Mongo1Proportion    float64 `yaml:"mongo1proportion"`
	Mongo2Proportion    float64 `yaml:"mongo2proportion"`
	KVRocksProportion   float64 `yaml:"kvrocksproportion"`
	CouchDBProportion   float64 `yaml:"couchdbproportion"`
	CassandraProportion float64 `yaml:"cassandraproportion"`

	Task1Proportion float64 `yaml:"task1proportion"`
	Task2Proportion float64 `yaml:"task2proportion"`
	Task3Proportion float64 `yaml:"task3proportion"`
}

// ----------------------------------------------------------------------------
