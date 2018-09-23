package blockchain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
)

var chainPath string

func init() {
	chainPath = "node/blockchain/blocks/"
}

func GetLast() Block {
	var res Block

	files, err := ioutil.ReadDir(chainPath)
	if err != nil {
		fmt.Println(err.Error())
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Unix() < files[j].ModTime().Unix()
	})
	file, _ := ioutil.ReadFile(chainPath + files[len(files)-1].Name())
	json.Unmarshal(file, &res)
	return res
}

func GetAll() BlockChain {
	files, _ := ioutil.ReadDir(chainPath)
	var res = make(BlockChain, len(files))
	for i := range res {
		file, _ := ioutil.ReadFile(chainPath + files[i].Name())
		json.Unmarshal(file, &res[i])
	}
	return res
}

func GetAfterTime(timestamp int64) BlockChain {
	files, _ := ioutil.ReadDir(chainPath)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Unix() > files[j].ModTime().Unix()
	})
	var res BlockChain
	for i := range files {
		file, _ := ioutil.ReadFile(chainPath + files[i].Name())
		block := &Block{}
		json.Unmarshal(file, &block)
		if block.Timestamp < timestamp {
			break
		}
		res = append(res, *block)
	}
	return res
}

func RemoveBlock(hash string) error {
	return os.Remove(chainPath + hash + ".json")
}

func PushBlock(b Block) {
	msg, _ := json.Marshal(b)
	ioutil.WriteFile(chainPath+b.Hash()+".json", msg, os.ModePerm)
}
