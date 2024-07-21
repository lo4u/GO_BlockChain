package blockchain

import (
	"GOPreject/constcoe"
	"GOPreject/utils"
	"bytes"
	"crypto/sha256"
	"math"
	"math/big"
)

func (b *Block) GetTarget() []byte {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-constcoe.DIFFICUTY))
	return target.Bytes()
}

func (b *Block) GetBase4Nonce(nonce int64) []byte {
	data := bytes.Join([][]byte{
		utils.Int2Bytes(b.Timestamp),
		b.PrevHash,
		b.Target,
		utils.Int2Bytes(nonce),
		b.getTransactionSummary(),
	},
		[]byte{},
	)
	return data
}

func (b *Block) FindNonce() int64 {
	//32位byte数组
	hash := [32]byte{}
	var nonce int64 = 0
	var intHash, intTarget big.Int
	intTarget.SetBytes(b.Target)

	for nonce < math.MaxInt64 {
		data := b.GetBase4Nonce(nonce)
		hash = sha256.Sum256(data)
		intHash.SetBytes(hash[:])
		if intHash.Cmp(&intTarget) == -1 {
			break
		} else {
			nonce++
		}
	}
	return nonce
}

func (b *Block) ValidatePoW() bool {
	var intHash big.Int
	var hash [32]byte
	var intTarget big.Int
	hash = sha256.Sum256(b.GetBase4Nonce(b.Nonce))
	intHash.SetBytes(hash[:])
	intTarget.SetBytes(b.Target)
	return intHash.Cmp(&intTarget) == -1
}
