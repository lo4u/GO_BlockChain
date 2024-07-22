package utxoset

import (
	"GOPreject/transaction"
	"GOPreject/utils"
	"bytes"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

const (
	INFO         = "INFO:"
	INFONAME     = INFO + "NAME"
	INFOHEIGHT   = INFO + "HIGT"
	UTXOKEY      = "UTXO:"
	UTXOKEYORDER = ":ORDER:"
)

type UTXOSet struct {
	Name   []byte //标识符
	DB     *badger.DB
	Height int64
}

func GetUTXOSetFile(dir string) string {
	fileAddress := dir + "/" + "MANIFEST"
	return fileAddress
}

func ToUTXOKey(txID []byte, order int) []byte {
	utxoKey := bytes.Join([][]byte{
		[]byte(UTXOKEY),
		txID,
		[]byte(UTXOKEYORDER),
		utils.Int2Bytes(int64(order)),
	}, []byte{})
	return utxoKey
}

func CreateUTXOSet(name []byte, dir string, utxos []transaction.UTXO, height int64) *UTXOSet {
	if utils.FileExists(GetUTXOSetFile(dir)) {
		fmt.Println("UTXOSet has already existed, now rebuild it.")
		err := os.RemoveAll(dir)
		utils.Handle(err)
	}

	opts := badger.DefaultOptions(dir)
	opts.Logger = nil
	db, err := badger.Open(opts)
	utils.Handle(err)

	utxoSet := UTXOSet{name, db, height}

	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(INFONAME), name)
		if err != nil {
			return err
		}
		err = txn.Set([]byte(INFOHEIGHT), utils.Int2Bytes(height))
		if err != nil {
			return err
		}
		for _, utxo := range utxos {
			utxoKey := ToUTXOKey(utxo.TxID, utxo.OutIdx)
			err = txn.Set(utxoKey, utxo.Serialize())
			if err != nil {
				return err
			}
		}
		return nil
	})
	utils.Handle(err)
	return &utxoSet
}

func LoadUTXOSet(dir string) *UTXOSet {
	if !utils.FileExists(GetUTXOSetFile(dir)) {
		fmt.Println("No UTXOSet found, please create one first.")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dir)
	opts.Logger = nil
	db, err := badger.Open(opts)
	utils.Handle(err)

	var name []byte
	var height int64
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(INFONAME))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			name = val
			return nil
		})
		if err != nil {
			return err
		}

		//获取高度
		item, err = txn.Get([]byte(INFOHEIGHT))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			height = utils.Bytes2Int(val)
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	utils.Handle(err)
	return &UTXOSet{name, db, height}
}

func (utxoSet *UTXOSet) AddUTXO(utxo *transaction.UTXO) {
	err := utxoSet.DB.Update(func(txn *badger.Txn) error {
		utxoKey := ToUTXOKey(utxo.TxID, utxo.OutIdx)
		err := txn.Set(utxoKey, utxo.Serialize())
		if err != nil {
			return err
		}
		return nil
	})
	utils.Handle(err)
}

func (utxoSet *UTXOSet) DelUTXO(txID []byte, order int) {
	err := utxoSet.DB.Update(func(txn *badger.Txn) error {
		utxoKey := ToUTXOKey(txID, order)
		err := txn.Delete(utxoKey)
		return err
	})
	utils.Handle(err)
}

func (utxoSet *UTXOSet) UpdateHeight(height int64) {
	utxoSet.Height = height
	err := utxoSet.DB.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(INFOHEIGHT), utils.Int2Bytes(height))
		return err
	})
	utils.Handle(err)
}

// 判断是否是UTXOSEt的name或者height， 还是UTXO信息
func IsInfo(inkey []byte) bool {
	return bytes.HasPrefix(inkey, []byte(INFO))
}
