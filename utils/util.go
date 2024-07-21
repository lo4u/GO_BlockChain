package utils

import (
	"GOPreject/constcoe"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"log"
	"math/big"
	"os"

	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"
)

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func Int2Bytes(num int64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buf.Bytes()
}

func FileExists(fileAddr string) bool {
	if _, err := os.Stat(fileAddr); os.IsNotExist(err) {
		return false
	}
	return true
}

func PubKeyHash(pubKey []byte) []byte {
	hashedPubKey := sha256.Sum256(pubKey)
	hasher := ripemd160.New()
	_, err := hasher.Write(hashedPubKey[:])
	Handle(err)
	return hasher.Sum(nil)
}

func CheckSum(ripemdHash []byte) []byte {
	checkSum := sha256.Sum256(ripemdHash)
	checkSum = sha256.Sum256(checkSum[:])
	return checkSum[:constcoe.CHECKSUMLEN]
}

// base58编码
func Base58Encode(bytes []byte) []byte {
	return []byte(base58.Encode(bytes))
}

func Base58Decode(bytes []byte) []byte {
	ret, err := base58.Decode(string(bytes))
	Handle(err)
	return ret
}

func PubHash2Address(pubKeyHash []byte) []byte {
	nettedPubKeyHash := append([]byte{constcoe.NETWORKVERSION}, pubKeyHash...)
	address := append(nettedPubKeyHash, CheckSum(nettedPubKeyHash)...)
	return Base58Encode(address)
}

func Address2PubHash(address []byte) []byte {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-constcoe.CHECKSUMLEN]
	return pubKeyHash
}

// 签名函数
func Sign(msg []byte, privKey ecdsa.PrivateKey) []byte {
	r, s, err := ecdsa.Sign(rand.Reader, &privKey, msg)
	Handle(err)
	signature := append(r.Bytes(), s.Bytes()...)
	return signature
}

func Verify(msg []byte, pubkey []byte, signature []byte) bool {
	curve := elliptic.P256()
	r := big.Int{}
	s := big.Int{}
	sigLen := len(signature)
	r.SetBytes(signature[:(sigLen / 2)])
	s.SetBytes(signature[(sigLen / 2):])

	x := big.Int{}
	y := big.Int{}
	keyLen := len(pubkey)
	x.SetBytes(pubkey[:(keyLen / 2)])
	y.SetBytes(pubkey[keyLen/2:])

	rawPubKey := ecdsa.PublicKey{
		Curve: curve,
		X:     &x,
		Y:     &y,
	}
	return ecdsa.Verify(&rawPubKey, msg, &r, &s)
}
