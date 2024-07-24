package wallet

import (
	"GOPreject/constcoe"
	"GOPreject/utils"
	"bytes"
	"encoding/gob"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// key --> address, value --> refname. Refname is repeatable, even empty, like "".
type RefList map[string]string

// Load a reflist from the local file named "ref_list.data"
func LoadRefList() *RefList {
	file := constcoe.WALLETSREFLIST + "ref_list.data"
	if !utils.FileExists(file) {
		refList := make(RefList)
		refList.Update()
		return &refList
	} else {
		content, err := os.ReadFile(file)
		utils.Handle(err)

		decoder := gob.NewDecoder(bytes.NewBuffer(content))
		refList := make(RefList)
		err = decoder.Decode(&refList)
		utils.Handle(err)
		return &refList
	}
}

// Save the reflist as a local file named "ref_list.data"
func (refList *RefList) Save() {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(refList)
	utils.Handle(err)

	file := constcoe.WALLETSREFLIST + "ref_list.data"
	err = os.WriteFile(file, buf.Bytes(), 0644)
	utils.Handle(err)
}

// Update the reflist object
//
// For those existing in reflist, this function will do nothing.
// And other wallets will be added to the reflist with a empty string as refname
func (refList *RefList) Update() {
	err := filepath.Walk(constcoe.WALLETSDIR, func(path string, info fs.FileInfo, err error) error {
		if (info == nil) || (info.IsDir()) || strings.Compare(filepath.Ext(path), ".wlt") != 0 {
			return err
		}
		fileName := info.Name()
		fileNameNoExtension := fileName[:len(fileName)-4]

		_, ok := (*refList)[fileNameNoExtension]
		if !ok {
			(*refList)[fileNameNoExtension] = ""
		}
		return err
	})
	utils.Handle(err)
}

// Bind a refname to a address in the reflist.
func (refList *RefList) BindRef(address, refName string) {
	(*refList)[address] = refName
}

// Find a address attached to a refname. This function acts like finding key by value.
func (refList *RefList) FindAddress(refName string) (string, error) {
	for key, val := range *refList {
		if val == refName && key != "" {
			return key, nil
		}
	}
	return "", errors.New("the refName is not found")
}
