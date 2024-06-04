package workload

type WorkloadParameter struct {
	DBName       string
	TableName    string
	WorkloadName string

	ThreadCount int
	DoBenchmark bool

	RecordCount       int `yaml:"recordcount"`
	OperationCount    int `yaml:"operationcount"`
	TxnOperationGroup int `yaml:"txnoperationgroup"`

	ReadProportion            float64 `yaml:"readproportion"`
	UpdateProportion          float64 `yaml:"updateproportion"`
	InsertProportion          float64 `yaml:"insertproportion"`
	ScanProportion            float64 `yaml:"scanproportion"`
	ReadModifyWriteProportion float64 `yaml:"readmodifywriteproportion"`
	DoubleSeqCommitProportion float64 `yaml:"doubleseqcommitproportion"`

	// These parameters are for the data consistency test
	InitialAmountPerKey   int `yaml:"initialamountperkey"`
	TransferAmountPerTxn  int `yaml:"transferamountpertxn"`
	TotalAmount           int `yaml:"totalamount"`
	PostCheckWorkerThread int `yaml:"postcheckworkerthread"`

	// These parameters are for the data distribution test
	GlobalDatastoreName string  `yaml:"globaldatastorename"`
	Redis1Proportion    float64 `yaml:"redis1proportion"`
	Mongo1Proportion    float64 `yaml:"mongo2proportion"`
	Mongo2Proportion    float64 `yaml:"mongo2proportion"`
	CouchDBProportion   float64 `yaml:"couchdbproportion"`
}

// ----------------------------------------------------------------------------
