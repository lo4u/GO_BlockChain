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

type RefList map[string]string

func (refList *RefList) Save() {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(refList)
	utils.Handle(err)

	file := constcoe.WALLETSREFLIST + "ref_list.data"
	err = os.WriteFile(file, buf.Bytes(), 0644)
	utils.Handle(err)
}

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

func (refList *RefList) BindRef(address, refName string) {
	(*refList)[address] = refName
}

func (refList *RefList) FindAddress(refName string) (string, error) {
	for key, val := range *refList {
		if val == refName && key != "" {
			return key, nil
		}
	}
	return "", errors.New("the refName is not found")
}
