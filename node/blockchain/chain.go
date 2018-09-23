package blockchain

import (
	"sort"
)

type BlockChain []Block

func (chain BlockChain) Sort() BlockChain {
	sort.Slice(chain, func(i, j int) bool {
		return chain[i].Id < chain[j].Id
	})
	return chain
}

func (chain BlockChain) Check() bool {
	for i := 0; i < len(chain)-1; i++ {
		if chain[i+1].PrevBlock != chain[i].Hash() {
			return false
		}
	}
	return true
}
