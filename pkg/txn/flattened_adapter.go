package txn

import "time"

type FlattenedItem interface {
	TxnMetadata
}

type FlattenedAdapter struct {
	FlattenedItem
}

// assert FlattenedAdapter implements TxnValueOperator
var _ TxnValueOperator = (*FlattenedAdapter)(nil)

var _ DataItemNext = (*FlattenedAdapter)(nil)

func (fa *FlattenedAdapter) ParseValue(any) error {
	panic("implement me")
}

func (fa *FlattenedAdapter) SetValue(any) error {
	panic("implement me")
}

func (fa *FlattenedAdapter) UpdateMetadata(DataItemNext, int64, time.Time) error {
	panic("implement me")
}

func (fa *FlattenedAdapter) GetValidItem(TStart int64) (DataItemNext, bool) {
	panic("implement me")
}

func (fa *FlattenedAdapter) GetPrevItem() (DataItemNext, error) {
	panic("implement me")
}
