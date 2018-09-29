package blockchain

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var chainPath string

/*
 File of functions, which works with database.
 For this simple blockchain, I use separate json files for each block
*/

var lastBlock = Block{Id: -1}

func init() {
	chainPath = "node/blockchain/blocks/"
}

func GetLast() Block {
	if lastBlock.Id == -1 {
		files, _ := ioutil.ReadDir(chainPath)
		lastBlock = GetAll().Sort()[len(files)-1]
	}
	return lastBlock
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

func GetSinceTime(timestamp int64) BlockChain {
	var res []Block = GetAll().Sort()
	for i := range res {
		if res[i].Timestamp >= timestamp {
			return res[i:]
		}
	}
	return res
}

func RemoveBlock(hash string) ([]Transaction, error) {
	file, err := ioutil.ReadFile(chainPath + hash + ".json")
	if err != nil {
		return nil, nil
	}
	block := &Block{}
	json.Unmarshal(file, &block)
	return block.Transactions, os.Remove(chainPath + hash + ".json")
}

func PushBlock(b Block) {
	msg, err := json.Marshal(b)
	if err != nil {
		panic(err.Error())
	}
	ioutil.WriteFile(chainPath+b.Hash()+".json", msg, os.ModePerm)
	lastBlock = b
}
