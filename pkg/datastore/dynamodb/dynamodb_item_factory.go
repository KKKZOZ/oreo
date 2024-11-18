package dynamodb

import "github.com/oreo-dtx-lab/oreo/pkg/txn"

var _ txn.DataItemFactory = (*DynamoDBItemFactory)(nil)

type DynamoDBItemFactory struct{}

func (d *DynamoDBItemFactory) NewDataItem(options txn.ItemOptions) txn.DataItem {
	return NewDynamoDBItem(options)
}
