package wallet

import (
	"GOPreject/constcoe"
	"GOPreject/transaction"
	"GOPreject/utils"
	"GOPreject/utxoset"
	"os"

	"github.com/dgraph-io/badger"
)

func (wallet *Wallet) GetUTXOSetDir() string {
	address_str := string(wallet.Address())
	path := constcoe.UTXOSET + address_str
	return path
}

func (wt *Wallet) LoadUTXOSet() *utxoset.UTXOSet {
	return utxoset.LoadUTXOSet(wt.GetUTXOSetDir())
}

func (pWallet *Wallet) RemoveUTXOSet() error {
	file := pWallet.GetUTXOSetDir()
	err := os.Remove(file)
	return err
}

func (pWallet *Wallet) GetBalance() int {
	balance := 0
	pUTXOSet := pWallet.LoadUTXOSet()
	defer pUTXOSet.DB.Close()

	err := pUTXOSet.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		iter := txn.NewIterator(opts)
		defer iter.Close()

		for iter.Rewind(); iter.Valid(); iter.Next() {
			item := iter.Item()
			if utxoset.IsInfo(item.Key()) {
				continue
			}
			err := item.Value(func(val []byte) error {
				pUTXO := transaction.Deserialize(val)
				balance += pUTXO.Output.Value
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	utils.Handle(err)
	return balance
}
