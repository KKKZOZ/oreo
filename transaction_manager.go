package main

import "github.com/google/uuid"

type TxnManager struct {
	TxnId string
}

var txnManager TxnManager = TxnManager{}

func (t *TxnManager) start() {
	globalDataStore.start()
	t.TxnId = uuid.NewString()
}
