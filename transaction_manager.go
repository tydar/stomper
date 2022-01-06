package main

import (
	"errors"
	"log"
)

type Transaction struct {
	frames []Frame
}

type TransactionManager struct {
	transactions map[string]Transaction
}

func NewTransactionManager() *TransactionManager {
	return &TransactionManager{
		transactions: make(map[string]Transaction),
	}
}

// StartTransaction is called when engine interprets a BEGIN frame
// adds a new transaction to the map iff no transaction with this id for this client exists
func (tm *TransactionManager) StartTransaction(id, connectionID string) error {
	internalTxId := internalTxId(id, connectionID)
	_, ok := tm.transactions[internalTxId]
	if ok {
		return errors.New("transaction with this ID already exists")
	}

	tm.transactions[internalTxId] = Transaction{
		frames: []Frame{},
	}

	log.Printf("client %s: created transaction %s\n", connectionID, id)

	return nil
}

// CommitTransaction is called when the engine handles a COMMIT frame
// it pulls the transaction by its ID and the client's ID, then sends the Transaction as one unit
// on the provided channel, at which point the Engine handles SENDing the messages
func (tm *TransactionManager) CommitTransaction(id, connectionID string, out chan Transaction) error {
	internalTxId := internalTxId(id, connectionID)
	tx, ok := tm.transactions[internalTxId]

	if ok {
		out <- tx
		delete(tm.transactions, internalTxId)
	} else {
		return errors.New("no such transaction found")
	}

	log.Printf("client %s: enacted commit for transaction %s\n", connectionID, id)

	return nil
}

// AbortTransaction is called when the engine handles an ABORT frame
// it pulls the transaction by its ID and the client's ID and deletes it
func (tm *TransactionManager) AbortTransaction(id, connectionID string) error {
	internalTxId := internalTxId(id, connectionID)
	_, ok := tm.transactions[internalTxId]

	if ok {
		delete(tm.transactions, internalTxId)
	} else {
		return errors.New("no such transaction found")
	}

	log.Printf("client %s: aborted transaction %s\n", connectionID, id)

	return nil
}

// AddFrame is called by the engine when it handles a frame that includes a transaction id header
// the frame is added iff a matching tx has been BEGINed by the same client
func (tm *TransactionManager) AddFrame(id, connectionID string, frame Frame) error {
	internalTxId := internalTxId(id, connectionID)

	tx, ok := tm.transactions[internalTxId]
	if ok {
		tm.transactions[internalTxId] = Transaction{
			frames: append(tx.frames, frame),
		}
	} else {
		return errors.New("no such transaction found")
	}

	log.Printf("client %s: added frame to transaction %s\n", connectionID, id)

	return nil
}

// internalTxId is used to return a consistent name for a given transaction
func internalTxId(id, connectionID string) string {
	return id + "_" + connectionID
}
