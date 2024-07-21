package transaction

import (
	"GOPreject/constcoe"
	"GOPreject/utils"
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/gob"
)

type Transaction struct {
	ID      []byte //hash值
	Inputs  []TxInput
	Outputs []TxOutput
}

func (tx *Transaction) TxHash() []byte {
	var encoded bytes.Buffer
	var hash [32]byte

	//gob是序列化用的，Newcoder接受一个缓冲区用来放序列化的结果
	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	utils.Handle(err)

	hash = sha256.Sum256(encoded.Bytes())
	return hash[:]
}

func (tx *Transaction) SetID() {
	tx.ID = tx.TxHash()
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

func (tx *Transaction) IsBase() bool {
	return len(tx.Inputs) == 1 && tx.Inputs[0].OutIdx == -1
}

// 复制一份交易信息,去掉input的地址和签名
func (tx *Transaction) PlainCopy() *Transaction {
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

func (tx *Transaction) PlainHash(inIndex int, prevPubKey []byte) []byte {
	txCopy := tx.PlainCopy()
	txCopy.Inputs[inIndex].PubKey = prevPubKey
	return txCopy.TxHash()
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey) {
	if tx.IsBase() {
		return
	}
	for idx, input := range tx.Inputs {
		plainHash := tx.PlainHash(idx, input.PubKey)
		signature := utils.Sign(plainHash, privKey)
		tx.Inputs[idx].Sign = signature
	}
}

func (tx *Transaction) Verify() bool {
	for index, input := range tx.Inputs {
		plainHash := tx.PlainHash(index, input.PubKey)
		if !utils.Verify(plainHash, input.PubKey, input.Sign) {
			return false
		}
	}
	return true
}
