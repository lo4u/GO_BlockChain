package blockchain

import (
	"GOPreject/transaction"
	"GOPreject/utils"
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
)

// 如果该input使用的UTXO是在这个交易池中的某个交易中的输出，返回true，和一个数字，表明该UTXO的面额
// 如果不是该池中任何一个交易的，则返回false
func isInputRight(txs []*transaction.Transaction, in transaction.TxInput) (bool, int) {
	//遍历每个交易
	for _, tx := range txs {
		if bytes.Equal(tx.ID, in.TxID) && bytes.Equal(utils.PubKeyHash(in.PubKey), tx.Outputs[in.OutIdx].HashPubKey) {
			return true, tx.Outputs[in.OutIdx].Value
		}
	}
	return false, 0
}

// Verify each transaction
func (pBlockChain *BlockChain) VerifyTransactions(txs []*transaction.Transaction) bool {
	if len(txs) == 0 {
		return true
	}
	//记录这个交易池中使用过的UTXO
	spentOutputs := make(map[string]int)
	for _, tx := range txs {
		pubKey := tx.Inputs[0].PubKey
		fromAddress := utils.PubHash2Address(utils.PubKeyHash(pubKey))
		UTXs := pBlockChain.FindUTXs(fromAddress)
		inputAmount := 0
		outputAmount := 0

		//遍历每个input
		//firstly, 确认在这个交易池子中未被使用
		for _, input := range tx.Inputs {
			if outIndex, ok := spentOutputs[hex.EncodeToString(input.TxID)]; ok && outIndex == input.OutIdx {
				log.Println("Warning: The fund might be utilized on multiple occasions")
				return false
			}
			//确认在这个交易池中这个input的来源UTXO是存在的。
			ok, amount := isInputRight(UTXs, input)
			if !ok {
				log.Println("falls in 2")
				return false
			}
			inputAmount += amount
			spentOutputs[hex.EncodeToString(input.TxID)] = input.OutIdx
		}
		for _, output := range tx.Outputs {
			outputAmount += output.Value
		}
		if inputAmount != outputAmount {
			log.Println("falls in 3")
			return false
		}
		if !tx.Verify() {
			log.Println("falls in 4")
			return false
		}
	}
	return true
}

// Run mine process and then remove the local transaction pool file.
func (pBlockChain *BlockChain) RunMine() bool {
	pTransactionPool := transaction.GetTransactionPool()
	if !pBlockChain.VerifyTransactions(pTransactionPool.PubTx) {
		log.Println("falls in transaction verification")
		err := transaction.RemoveTransactionPoolFile()
		utils.Handle(err)
		return false
	}
	pCandidateBlock := CreateBlock(pBlockChain.LastHash, pBlockChain.GetHeight()+1, pTransactionPool.PubTx)
	if pCandidateBlock.ValidatePoW() {
		pBlockChain.AddBlock(pCandidateBlock)
		err := transaction.RemoveTransactionPoolFile()
		utils.Handle(err)
		return true
	} else {
		fmt.Println("Block has invalid nonce")
		return false
	}
}
