package blockchain

import (
	"GOPreject/constcoe"
	"GOPreject/transaction"
	"GOPreject/utils"
	"bytes"
	"encoding/gob"
	"os"
)

type TransactionPool struct {
	PubTx []*transaction.Transaction
}

// Load the local transaction pool file. If nothing found, create a new one.
func GetTransactionPool() *TransactionPool {
	transactionPool := TransactionPool{}
	err := transactionPool.LoadFile()
	utils.Handle(err)
	return &transactionPool
}

// remove the local transaction pool file.
func RemoveTransactionPoolFile() error {
	err := os.Remove(constcoe.TRANSACTIONPOOLFILE)
	return err
}

// Add a transaction to the pool.
func (tp *TransactionPool) AddTransaction(tx *transaction.Transaction) {
	tp.PubTx = append(tp.PubTx, tx)
}

// save the transaction pool as a local file.
func (tp *TransactionPool) SaveFile() {
	var content bytes.Buffer
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(tp)
	utils.Handle(err)
	err = os.WriteFile(constcoe.TRANSACTIONPOOLFILE, content.Bytes(), 0644)
	utils.Handle(err)
}

// Load a transaction pool file and
func (tp *TransactionPool) LoadFile() error {
	if !utils.FileExists(constcoe.TRANSACTIONPOOLFILE) {
		//文件不存在返回nil
		return nil
	}

	var transactionPool TransactionPool
	fileContent, err := os.ReadFile(constcoe.TRANSACTIONPOOLFILE)
	if err != nil {
		return err
	}

	decoder := gob.NewDecoder(bytes.NewBuffer(fileContent))
	err = decoder.Decode(&transactionPool)
	if err != nil {
		return err
	}

	tp.PubTx = transactionPool.PubTx
	return nil
}
