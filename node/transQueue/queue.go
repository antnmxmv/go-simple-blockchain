package transQueue

import (
	"go-simple-blockchain/node/blockchain"
	"sync"
)

var trans struct {
	set          map[string]bool
	queue        []blockchain.Transaction
	currentBlock *blockchain.Block
	mux          sync.Mutex
}

var CurrentBlock = &trans.currentBlock

func SetCurrent(b *blockchain.Block) {
	trans.currentBlock = b
}

func init() {
	trans.queue = make([]blockchain.Transaction, 0)
	trans.set = make(map[string]bool)
	trans.currentBlock = &blockchain.Block{}
}

func Lock() {
	trans.mux.Lock()
}

func Unlock() {
	trans.mux.Unlock()
}

func Push(transaction blockchain.Transaction) bool {
	if _, ok := trans.set[transaction.Sign]; ok {
		return false
	}
	for _, i := range trans.currentBlock.Transactions {
		if i.Sign == transaction.Sign {
			return false
		}
	}
	trans.set[transaction.Sign] = true
	trans.queue = append(trans.queue, transaction)
	return true
}

func Pop() blockchain.Transaction {
	res := trans.queue[0]
	delete(trans.set, res.Sign)
	trans.queue = trans.queue[1:]
	return res
}

func Remove(hash string) bool {
	if _, ok := trans.set[hash]; !ok {
		return false
	}
	delete(trans.set, hash)
	for i := 0; i < len(trans.queue); i++ {
		if trans.queue[i].Sign == hash {
			trans.queue = append(trans.queue[:i], trans.queue[i+1:]...)
			i--
		}
	}
	return true
}

func Size() int {
	return len(trans.set)
}
