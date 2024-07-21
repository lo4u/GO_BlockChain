package transaction

import (
	"GOPreject/utils"
	"bytes"
)

type TxOutput struct {
	//隐含：
	//1. 每次交易输出地址不会出现两次，意思是要打给某人的前一次输出打完
	//2. 可能存在打给自己的钱，但该笔交易中的输出不能在同一次交易中被使用
	Value      int
	HashPubKey []byte
}

type TxInput struct {
	TxID   []byte //前一个交易的ID
	OutIdx int    //前一个交易中，指示这个Input是第几个输出
	PubKey []byte
	Sign   []byte
}

func (in *TxInput) FromAddressEqual(address []byte) bool {
	inAddress := utils.PubKeyHash(in.PubKey)
	inAddress = utils.PubHash2Address(inAddress)
	return bytes.Equal(inAddress, address)
}

func (out *TxOutput) ToAddressEqual(address []byte) bool {
	return bytes.Equal(utils.PubHash2Address(out.HashPubKey), address)
}
