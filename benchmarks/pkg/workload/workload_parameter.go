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
	Redis1Proportion    float64 `yaml:"redis1proportion" oreo:"Redis"`
	Mongo1Proportion    float64 `yaml:"mongo1proportion" oreo:"MongoDB1"`
	Mongo2Proportion    float64 `yaml:"mongo2proportion" oreo:"MongoDB2"`
	KVRocksProportion   float64 `yaml:"kvrocksproportion" oreo:"KVRocks"`
	CouchDBProportion   float64 `yaml:"couchdbproportion" oreo:"CouchDB"`
	CassandraProportion float64 `yaml:"cassandraproportion" oreo:"Cassandra"`
	DynamoDBProportion  float64 `yaml:"dynamodbproportion" oreo:"DynamoDB"`
	TiKVProportion      float64 `yaml:"tikvproportion" oreo:"TiKV"`

	Task1Proportion float64 `yaml:"task1proportion"`
	Task2Proportion float64 `yaml:"task2proportion"`
	Task3Proportion float64 `yaml:"task3proportion"`
}

// ----------------------------------------------------------------------------
