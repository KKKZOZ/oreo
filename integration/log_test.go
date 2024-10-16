package integration

import (
	"testing"

	"github.com/oreo-dtx-lab/oreo/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNormalDebug(t *testing.T) {

	preTxn := NewTransactionWithSetup(REDIS)
	preTxn.Start()
	person := testutil.NewPerson("kkkzoz")
	preTxn.Write(REDIS, "kkkzoz", person)
	err := preTxn.Commit()
	assert.NoError(t, err)

	txn := NewTransactionWithSetup(REDIS)
	txn.Start()
	var p testutil.Person
	txn.Read(REDIS, "kkkzoz", &p)
	p.Age = 23
	txn.Write(REDIS, "kkkzoz", p)
	err = txn.Commit()
	assert.NoError(t, err)
}
