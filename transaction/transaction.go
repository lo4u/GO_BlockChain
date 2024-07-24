package transaction

import (
	"GOPreject/constcoe"
	"GOPreject/utils"
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/gob"
)

// Assume that inputs' source address can be repeatable, but outputs' target address can not.
// And a UTXO from a certain transaction must be used in an upcoming transaction.
type Transaction struct {
	ID      []byte //hash值
	Inputs  []TxInput
	Outputs []TxOutput
}

func BaseTx(toAddress []byte) *Transaction {
	//创建初始交易信息
	txIn := TxInput{[]byte{}, -1, []byte{}, nil}
	txOut := TxOutput{constcoe.INITCOIN, utils.Address2PubHash(toAddress)}
	tx := Transaction{[]byte("This is the Base Transaction"),
		[]TxInput{txIn},
		[]TxOutput{txOut}}
	return &tx
}

func (tx *Transaction) txHash() []byte {
	var encoded bytes.Buffer
	var hash [32]byte

	//gob是序列化用的，Newcoder接受一个缓冲区用来放序列化的结果
	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	utils.Handle(err)

	hash = sha256.Sum256(encoded.Bytes())
	return hash[:]
}

// Copy a transaction and remove all inputs' PubKey and Sign.
func (tx *Transaction) plainCopy() *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, txin := range tx.Inputs {
		inputs = append(inputs, TxInput{
			TxID:   txin.TxID,
			OutIdx: txin.OutIdx,
			PubKey: nil,
			Sign:   nil,
		})
	}
	for _, txout := range tx.Outputs {
		outputs = append(outputs, TxOutput{
			Value:      txout.Value,
			HashPubKey: txout.HashPubKey,
		})
	}

	return &Transaction{
		ID:      tx.ID,
		Inputs:  inputs,
		Outputs: outputs,
	}
}

// Specify a input, and set its PubKey to prevPubKey. Return the hash of the operation result.
func (tx *Transaction) plainHash(inIndex int, prevPubKey []byte) []byte {
	txCopy := tx.plainCopy()
	txCopy.Inputs[inIndex].PubKey = prevPubKey
	return txCopy.txHash()
}

// Set a transaction's ID value.
func (tx *Transaction) SetID() {
	tx.ID = tx.txHash()
}

// Check if the transaction is a base transaction.
func (tx *Transaction) IsBase() bool {
	return len(tx.Inputs) == 1 && tx.Inputs[0].OutIdx == -1
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey) {
	if tx.IsBase() {
		return
	}
	for idx, input := range tx.Inputs {
		plainHash := tx.plainHash(idx, input.PubKey)
		signature := utils.Sign(plainHash, privKey)
		tx.Inputs[idx].Sign = signature
	}
}

func (tx *Transaction) Verify() bool {
	for index, input := range tx.Inputs {
		plainHash := tx.plainHash(index, input.PubKey)
		if !utils.Verify(plainHash, input.PubKey, input.Sign) {
			return false
		}
	}
	return true
}
