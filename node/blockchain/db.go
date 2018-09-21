package blockchain

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"
)

func GetLast() Block {
	var res Block
	files, _ := ioutil.ReadDir("node/blockchain/blocks/")
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Unix() < files[j].ModTime().Unix()
	})
	file, _ := ioutil.ReadFile("node/blockchain/blocks/" + files[len(files)-1].Name())
	json.Unmarshal(file, &res)
	return res
}

func GetAll() []Block {
	files, _ := ioutil.ReadDir("node/blockchain/blocks/")
	var res = make([]Block, len(files))
	for i := range res {
		file, _ := ioutil.ReadFile("node/blockchain/blocks/" + files[i].Name())
		json.Unmarshal(file, &res[i])
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Id < res[j].Id
	})
	return res
}

func SendBlock(b Block) {
	msg, _ := json.Marshal(b)
	ioutil.WriteFile("node/blockchain/blocks/"+b.Hash()+".json", msg, os.ModePerm)
}
