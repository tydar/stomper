package main

import (
	"errors"
)

type Transaction struct {
	frames []Frame
}

type TransactionManager struct {
	transactions map[string]Transaction
}

// StartTransaction is called when engine interprets a begin frame
// adds a new transaction to the map iff no transaction with this id for this client exists
func (tm *TransactionManager) StartTransaction(id, connectionID string) error {
	internalTxId := id + "_" + "connectionID"
	_, ok := tm.transactions[internalTxId]
	if ok {
		return errors.New("transaction with this ID already exists")
	}

	tm.transactions[internalTxId] = Transaction{
		frames: []Frame{},
	}

	return nil
}

func (tm *TransactionManager) CommitTransaction(id, connectionID string, out chan Transaction) error {
	return nil
}

func (tm *TransactionManager) AbortTransaction(id, connectionID string) error {
	return nil
}
