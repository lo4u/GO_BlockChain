package blockchain

import (
	"GOPreject/constcoe"
	"GOPreject/merkletree"
	"GOPreject/transaction"
	"GOPreject/utils"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"
)

type Block struct {
	Timestamp    int64                      //创建的时间戳
	Hash         []byte                     //自己的摘要值
	PrevHash     []byte                     //前一个区块的摘要
	Height       int64                      //区块高度
	Target       []byte                     //工作目标值，用于POW
	Nonce        int64                      //工作计算的结果
	MTree        *merkletree.MerkleTree     //merkleTree结构
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
		pBlock.getTransactionSummary(),
		pBlock.MTree.RootNode.HashData}, []byte{})
	hash := sha256.Sum256(infor)
	pBlock.Hash = hash[:]
}

func CreateBlock(prevHash []byte, height int64, transactions []*transaction.Transaction) *Block {
	block := Block{time.Now().Unix(), []byte{}, prevHash, height, []byte{}, 0, merkletree.CreateMerkleTree(transactions), transactions}
	block.Target = block.GetTarget()
	block.Nonce = block.FindNonce()
	block.SetHash()
	return &block
}

// 生成初始区块
func GenesisBlock(address []byte) *Block {
	block := CreateBlock([]byte(constcoe.PREVHASH), 0, []*transaction.Transaction{transaction.BaseTx(address)})
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
