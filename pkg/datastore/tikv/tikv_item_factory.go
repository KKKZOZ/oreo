package tikv

import "github.com/kkkzoz/oreo/pkg/txn"

var _ txn.DataItemFactory = (*TiKVItemFactory)(nil)

type TiKVItemFactory struct{}

func (t *TiKVItemFactory) NewDataItem(options txn.ItemOptions) txn.DataItem {
	return NewTiKVItem(options)
}
