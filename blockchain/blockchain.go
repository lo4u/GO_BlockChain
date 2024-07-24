package blockchain

import (
	"GOPreject/constcoe"
	"GOPreject/transaction"
	"GOPreject/utils"
	"GOPreject/utxoset"
	"GOPreject/wallet"
	"bytes"
	"encoding/hex"
	"fmt"
	"runtime"

	"github.com/dgraph-io/badger"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

// It is like C++'s iterator object. However, there are a few function implemented.
type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

// a new iterator points to the newest block.
func (blockChain *BlockChain) NewIterator() *BlockChainIterator {
	iterator := BlockChainIterator{blockChain.LastHash, blockChain.Database}
	return &iterator
}

// Return a iterator which points to the initial block's previous address.
func (pBlockChain *BlockChain) End() *BlockChainIterator {
	var ogPrevHash []byte

	err := pBlockChain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("ogprevhash"))
		utils.Handle(err)
		err = item.Value(func(val []byte) error {
			ogPrevHash = val
			return nil
		})
		utils.Handle(err)
		return nil
	})
	utils.Handle(err)

	iterator := BlockChainIterator{ogPrevHash, pBlockChain.Database}
	return &iterator
}

func (iterator *BlockChainIterator) Next() *Block {
	var pBlock *Block

	err := iterator.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iterator.CurrentHash)
		utils.Handle(err)
		err = item.Value(func(val []byte) error {
			pBlock = DeSerialize(val)
			return nil
		})
		utils.Handle(err)
		return nil
	})
	utils.Handle(err)

	iterator.CurrentHash = pBlock.PrevHash
	return pBlock
}

func (iter1 *BlockChainIterator) Equal(iter2 *BlockChainIterator) bool {
	return bytes.Equal(iter1.CurrentHash, iter2.CurrentHash) && (iter1.Database == iter2.Database)
}

// Add a block to the blockchain
func (pBlockChain *BlockChain) AddBlock(newBlock *Block) {
	var lastHash []byte

	err := pBlockChain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		utils.Handle(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		utils.Handle(err)
		return nil
	})
	utils.Handle(err)
	if !bytes.Equal(newBlock.PrevHash, lastHash) {
		fmt.Println("This block is out of order")
		runtime.Goexit()
	}

	err = pBlockChain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		utils.Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)
		utils.Handle(err)
		pBlockChain.LastHash = newBlock.Hash
		return nil
	})
	utils.Handle(err)
}

// initialize a blockchain and give a original fund to the specified address
func InitBlockChain(address []byte) *BlockChain {
	var lashHash []byte

	if utils.FileExists(constcoe.BCFILE) {
		fmt.Println("blockchain already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(constcoe.BCPATH)
	opts.Logger = nil

	db, err := badger.Open(opts)
	utils.Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		genesis := GenesisBlock(address)
		fmt.Println("Genesis Created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		utils.Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash)
		utils.Handle(err)
		err = txn.Set([]byte("ogprevhash"), genesis.PrevHash)
		utils.Handle(err)
		lashHash = genesis.Hash
		return nil
	})
	utils.Handle(err)

	blockchain := BlockChain{
		LastHash: lashHash,
		Database: db,
	}
	return &blockchain
}

// Load the local blockchain file
func ContinueBlockChain() *BlockChain {
	if !utils.FileExists(constcoe.BCFILE) {
		fmt.Println("No blockchain found, please create one firse")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(constcoe.BCPATH)
	opts.Logger = nil
	db, err := badger.Open(opts)
	utils.Handle(err)

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		utils.Handle(err)
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		utils.Handle(err)
		return nil
	})
	utils.Handle(err)

	chain := BlockChain{lastHash, db}
	return &chain
}

// Find all transactions those include UTXO. Return these transaction as a pointer slice.
func (pBlockChain *BlockChain) FindUTXs(address []byte) []*transaction.Transaction {
	var UTX []*transaction.Transaction //所有 存在输出未被使用的交易（记录）
	STXO := make(map[string][]int)     //已使用的交易输出
	iter := pBlockChain.NewIterator()  //创建一个区块的迭代器
	end := pBlockChain.End()
	//遍历每个区块
	for !iter.Equal(end) {
		pBlock := iter.Next()
		//遍历这个区块的每个交易
		for _, ptx := range pBlock.Transactions {

			//key不可以是slice类型, 下面转为16进制字串
			txID := hex.EncodeToString(ptx.ID)

		IterOutputs:
			//遍历这个交易的每个输出
			for outIdx, out := range ptx.Outputs {
				if STXO[txID] != nil {
					for _, spentOut := range STXO[txID] {
						if spentOut == outIdx {
							//看这个输出是不是用了
							continue IterOutputs
						}
					}
				}
				if out.ToAddressEqual(address) {
					//看这个输出是属于我的吗？
					UTX = append(UTX, ptx)
				}
			}
			//遍历这个交易的每个输入，看之前的余额（交易输出）是否被用掉。
			//这里对输出是否被使用的检测放在后面，并无不妥，因为要使用交易输出，
			//必须在下一次交易中使用，而不能在同一个交易中使用。
			if !ptx.IsBase() {
				for _, in := range ptx.Inputs {
					if in.FromAddressEqual(address) {
						inTxID := hex.EncodeToString(in.TxID)
						STXO[inTxID] = append(STXO[inTxID], in.OutIdx)
					}
				}
			}
		}
	}
	return UTX
}

// Discarded
// func (pBlockChain *BlockChain) FindUTXOs(address []byte) (int, map[string]int) {
// 	unspentOut := make(map[string]int)
// 	//用FindUTX找到未使用的交易输出所在的交易
// 	unspentTX := pBlockChain.FindUTXs(address)
// 	//计算余额
// 	sum := 0
//
// Work:
// 	for _, ptx := range unspentTX {
// 		txID := hex.EncodeToString(ptx.ID)
// 		for outIdx, out := range ptx.Outputs {
// 			if out.ToAddressEqual(address) {
// 				unspentOut[txID] = outIdx
// 				sum += out.Value
// 				continue Work
// 			}
// 		}
// 	}
//
// 	return sum, unspentOut
// }

// Search for avaliable UTXOs. Return the money amount found and a map of transaction ID and corresponding output index.
//
// Note: the amount returned might be less than required amount, if the sum of all UTXOs is limited.
func (pBlockChain *BlockChain) FindSpendableOutputs(address []byte, amount int) (int, map[string]int) {
	unspentOut := make(map[string]int)
	//用FindUTX找到未使用的交易输出所在的交易
	unspentTX := pBlockChain.FindUTXs(address)
	//计算余额
	sum := 0

Work:
	for _, ptx := range unspentTX {
		txID := hex.EncodeToString(ptx.ID)
		for outIdx, out := range ptx.Outputs {
			if out.ToAddressEqual(address) {
				unspentOut[txID] = outIdx
				sum += out.Value
				if sum >= amount {
					break Work
				}
				continue Work
			}
		}
	}

	//要是钱不够，就会遍历所有的交易信息，所以调用本函数的时候还是需要检测返回的数目够不够的
	return sum, unspentOut
}

// Create a Transaction and return a boolean value indicating whether the creation was successful.
//
// Note: this function retrive the private key automatically and perform the signing process,
// without requiring the caller to explicitly provide it.
func (pBlockChain *BlockChain) CreateTransaction(fromAddress, toAddress []byte, amount int) (*transaction.Transaction, bool) {
	var inputs []transaction.TxInput
	var outputs []transaction.TxOutput
	fromWallet := wallet.LoadWallet(fromAddress)

	money, validOutputs := pBlockChain.FindSpendableOutputs(fromAddress, amount)
	if money < amount {
		fmt.Printf("Not enougn coins!\n")
		return &transaction.Transaction{}, false
	}

	for txID_Enc, outIdx := range validOutputs {
		//创建每个input结构体
		txID_Dec, err := hex.DecodeString(txID_Enc)
		utils.Handle(err)
		inputs = append(inputs, transaction.TxInput{
			TxID:   txID_Dec,
			OutIdx: outIdx,
			PubKey: fromWallet.PublicKey,
			Sign:   nil,
		})
	}

	outputs = append(outputs, transaction.TxOutput{
		Value:      amount,
		HashPubKey: utils.Address2PubHash(toAddress),
	})
	//如果有多，就打给自己
	if money > amount {
		outputs = append(outputs, transaction.TxOutput{
			Value:      money - amount,
			HashPubKey: utils.Address2PubHash(fromAddress),
		})
	}

	tx := transaction.Transaction{
		ID:      nil,
		Inputs:  inputs,
		Outputs: outputs,
	}
	tx.SetID()
	tx.Sign(fromWallet.PrivateKey)

	return &tx, true
}

// Get the newest block
func (blockChain *BlockChain) GetCurrentBlock() *Block {
	var block *Block
	err := blockChain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(blockChain.LastHash)
		utils.Handle(err)

		err = item.Value(func(val []byte) error {
			block = DeSerialize(val)
			return nil
		})
		utils.Handle(err)
		return nil
	})
	utils.Handle(err)
	return block
}

// Get the height of current block. The initial block's height is 0.
func (blockChain *BlockChain) GetHeight() int64 {
	return blockChain.GetCurrentBlock().Height
}

// On this blockchain, find all UTXOs and concatenate a slice.
func (pBlockChain *BlockChain) GetUTXOs(address []byte) []transaction.UTXO {
	var utxos []transaction.UTXO
	unspentTxs := pBlockChain.FindUTXs(address)

Work:
	for _, tx := range unspentTxs {
		for i, output := range tx.Outputs {
			if output.ToAddressEqual(address) {
				utxos = append(utxos, transaction.UTXO{
					TxID:   tx.ID,
					OutIdx: i,
					Output: output,
				})
				continue Work
			}
		}
	}
	return utxos
}

// Create a UTXO set from the blockchain corresponding the specified address.
func (pBlockChain *BlockChain) CreataUTXOSet(address []byte) *utxoset.UTXOSet {
	pWallet := wallet.LoadWallet(address)
	utxos := pBlockChain.GetUTXOs(pWallet.Address())
	return utxoset.CreateUTXOSet(pWallet.Address(), pWallet.GetUTXOSetDirName(), utxos, pBlockChain.GetHeight())
}

// Try to update the UTXO set corresponding the specified address.
//
// If the difference in height is greater than one, then a new UTXO set will be created.
func (pBlockChain *BlockChain) UpdateUTXOSet(address []byte) {
	pWallet := wallet.LoadWallet(address)
	utxoSet := pWallet.LoadUTXOSet()

	if pBlockChain.GetHeight() > utxoSet.Height+1 {
		utxoSet.DB.Close()
		err := pWallet.RemoveUTXOSet()
		utils.Handle(err)
		newUTXOSet := pBlockChain.CreataUTXOSet(address)
		newUTXOSet.DB.Close()
		return
	} else if pBlockChain.GetHeight() == utxoSet.Height+1 {
		pBlock := pBlockChain.GetCurrentBlock()
		for _, tx := range pBlock.Transactions {
			for _, input := range tx.Inputs {
				if input.FromAddressEqual(address) {
					utxoSet.DelUTXO(input.TxID, input.OutIdx)
				}
			}
			for outIdx, output := range tx.Outputs {
				if output.ToAddressEqual(address) {
					utxoSet.AddUTXO(&transaction.UTXO{
						TxID:   tx.ID,
						OutIdx: outIdx,
						Output: output,
					})
				}
			}
		}
		utxoSet.UpdateHeight(pBlock.Height)
	}
	utxoSet.DB.Close()
}
