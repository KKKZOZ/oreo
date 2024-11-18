package dynamodb

import "github.com/oreo-dtx-lab/oreo/pkg/txn"

type DynamoDBDatastore struct {
	*txn.Datastore
}

func NewDynamoDBDatastore(name string, conn txn.Connector) txn.Datastorer {
	return txn.NewDatastore(name, conn, &DynamoDBItemFactory{})
}
