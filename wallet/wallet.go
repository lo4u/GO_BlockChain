package wallet

import (
	"GOPreject/constcoe"
	"GOPreject/utils"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"math/big"
	"os"
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	utils.Handle(err)
	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	return *privateKey, publicKey
}

func NewWallet() *Wallet {
	privateKey, pubKey := NewKeyPair()
	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  pubKey,
	}
}

func (w *Wallet) Address() []byte {
	pubKeyHash := utils.PubKeyHash(w.PublicKey)
	return utils.PubHash2Address(pubKeyHash)
}

func (w *Wallet) Save() {
	gob.Register(elliptic.P256())

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	privPubKey := struct {
		D big.Int
		X big.Int
		Y big.Int
	}{*w.PrivateKey.D, *w.PrivateKey.PublicKey.X, *w.PrivateKey.PublicKey.Y}
	err := encoder.Encode(&privPubKey)
	utils.Handle(err)

	file := constcoe.WALLETSDIR + string(w.Address()) + ".wlt"
	err = os.WriteFile(file, buf.Bytes(), 0644)
	utils.Handle(err)
}

func LoadWallet(address []byte) *Wallet {
	file := constcoe.WALLETSDIR + string(address) + ".wlt"
	if !utils.FileExists(file) {
		utils.Handle(errors.New("no wallet with such address"))
	}
	content, err := os.ReadFile(file)
	utils.Handle(err)

	pBuf := bytes.NewBuffer(content)
	decoder := gob.NewDecoder(pBuf)
	var w Wallet
	var privPubKey struct {
		D big.Int
		X big.Int
		Y big.Int
	}
	decoder.Decode(&privPubKey)
	utils.Handle(err)
	w.PrivateKey.D = &privPubKey.D
	w.PrivateKey.PublicKey.Curve = elliptic.P256()
	w.PrivateKey.PublicKey.X = &privPubKey.X
	w.PrivateKey.PublicKey.Y = &privPubKey.Y
	w.PublicKey = append(w.PrivateKey.PublicKey.X.Bytes(), w.PrivateKey.PublicKey.Y.Bytes()...)
	return &w
}
