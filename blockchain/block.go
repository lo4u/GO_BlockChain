package blockchain

import (
	"GOPreject/constcoe"
	"GOPreject/transaction"
	"GOPreject/utils"
	"encoding/gob"

	"bytes"
	"crypto/sha256"
	"time"
)

type Block struct {
	Timestamp    int64                      //创建的时间戳
	Hash         []byte                     //自己的摘要值
	PrevHash     []byte                     //前一个区块的摘要
	Target       []byte                     //工作目标值，用于POW
	Nonce        int64                      //工作计算的结果
	Transactions []*transaction.Transaction //载荷
}

func (pBlock *Block) getTransactionSummary() []byte {
	txIDs := make([][]byte, 0)
	for _, tx := range pBlock.Transactions {
		txIDs = append(txIDs, tx.ID)
	}
	summary := bytes.Join(txIDs, []byte{})
	return summary
}

func (pBlock *Block) SetHash() {
	infor := bytes.Join([][]byte{
		utils.Int2Bytes(pBlock.Timestamp),
		pBlock.PrevHash,
		pBlock.Target,
		utils.Int2Bytes(pBlock.Nonce),
		pBlock.getTransactionSummary()}, []byte{})
	hash := sha256.Sum256(infor)
	pBlock.Hash = hash[:]
}

func CreateBlock(prevHash []byte, transaction []*transaction.Transaction) *Block {
	block := Block{time.Now().Unix(), []byte{}, prevHash, []byte{}, 0, transaction}
	block.Target = block.GetTarget()
	block.Nonce = block.FindNonce()
	block.SetHash()
	return &block
}

// 生成初始区块
func GenesisBlock(address []byte) *Block {
	block := CreateBlock([]byte(constcoe.PREVHASH), []*transaction.Transaction{transaction.BaseTx(address)})
	block.SetHash()
	return block
}

func (pBlock *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	utils.Handle(encoder.Encode(pBlock))
	return res.Bytes()
}

func DeSerialize(data []byte) *Block {
	var pBlock *Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	utils.Handle(decoder.Decode(&pBlock))
	return pBlock
}
