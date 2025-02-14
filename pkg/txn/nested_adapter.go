package txn

import (
	"fmt"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/errs"
)

type NestedItem interface {
	TxnMetadata

	Value() []byte
	SetValue([]byte) error

	Prev() []byte
	SetPrev([]byte)
	LinkedLen() int
	SetLinkedLen(int)

	// If prev field is empty, return an error of errs.NotFoundInAVC
	GetPrevItem() (NestedItem, error)
}

// assert NestedAdapter implements TxnValueOperator
var _ TxnValueOperator = (*NestedAdapter)(nil)

var _ DataItemNext = (*NestedAdapter)(nil)

type NestedAdapter struct {
	NestedItem
}

func (na *NestedAdapter) ParseValue(valuePtr any) error {
	err := config.Config.Serializer.Deserialize(na.Value(), valuePtr)
	if err != nil {
		return errs.NewSerializerError(errs.DeserializeFailed, err)
	}
	return nil
}

func (na *NestedAdapter) SetValue(v any) error {
	value, err := config.Config.Serializer.Serialize(v)
	if err != nil {
		return errs.NewSerializerError(errs.SerializeFailed, err)
	}
	na.SetValue(value)
	return nil
}

func (na *NestedAdapter) GetPrevItem() (DataItemNext, error) {
	item, err := na.NestedItem.GetPrevItem()
	if err != nil {
		return nil, err
	}
	return &NestedAdapter{item}, nil
}

// truncate truncates the linked list of DataItems to the max length
func truncate(newItem NestedItem) (NestedItem, error) {
	maxLen := config.Config.MaxRecordLength

	if newItem.LinkedLen() <= maxLen {
		return newItem, nil
	}

	stack := util.NewStack[NestedItem]()
	stack.Push(newItem)
	curItem := &newItem
	for i := 1; i <= maxLen-1; i++ {
		preItem, err := (*curItem).GetPrevItem()
		// preItem, err := getPrevItem(*curItem)
		if err != nil {
			return nil, fmt.Errorf("truncate error: %w", err)
		}
		curItem = &preItem
		stack.Push(*curItem)
	}

	tarItem, err := stack.Pop()
	if err != nil {
		return nil, fmt.Errorf("stack Pop error: %w", err)
	}
	tarItem.SetPrev([]byte(""))
	tarItem.SetLinkedLen(1)

	for !stack.IsEmpty() {
		item, err := stack.Pop()
		if err != nil {
			return nil, fmt.Errorf("stack Pop error: %w", err)
		}
		bs, err := config.Config.Serializer.Serialize(tarItem)
		if err != nil {
			return nil, errs.NewSerializerError(errs.SerializeFailed, err)
		}
		item.SetPrev(bs)
		item.SetLinkedLen(tarItem.LinkedLen() + 1)
		tarItem = item
	}
	return tarItem, nil

}

func (na *NestedAdapter) UpdateMetadata(oldItem DataItemNext, TCommit int64, TLease time.Time) error {
	oldNestedItem, ok := oldItem.(NestedItem)
	if !ok {
		// WTF
		return errs.NewTypeAssertionError("oldItem", "NestedItem", fmt.Sprintf("%T", oldItem))
	}

	if oldItem == nil {
		na.SetLinkedLen(1)
	} else {
		na.SetLinkedLen(oldNestedItem.LinkedLen() + 1)
		bs, err := config.Config.Serializer.Serialize(oldItem)
		if err != nil {
			return err
		}
		na.SetPrev(bs)
		na.SetVersion(oldItem.Version())
	}

	// truncate the record
	newItem, err := truncate(na.NestedItem)
	if err != nil {
		return err
	}

	newItem.SetTxnState(config.PREPARED)
	newItem.SetTValid(TCommit)
	// TODO: time.Now() is temporary
	newItem.SetTLease(TLease)
	na.NestedItem = newItem
	// ni = newItem.(*NestedRedisItem)
	return nil
}

func (na *NestedAdapter) GetValidItem(TStart int64) (DataItemNext, bool) {
	curItem := na
	for i := 1; i <= config.Config.MaxRecordLength; i++ {

		if curItem.TValid() < TStart {
			return curItem, true
		}

		// TODO: check if necessary
		if i == config.Config.MaxRecordLength {
			break
		}

		preItem, err := curItem.GetPrevItem()
		if err != nil {
			return curItem, false
		}
		curItem = preItem.(*NestedAdapter)
	}
	return curItem, false
}
