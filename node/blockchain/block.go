package blockchain

import (
	"crypto/sha256"
	"fmt"
	"strconv"
)

type Block struct {
	PrevBlock    string
	Id           int
	Timestamp    int64
	Transactions []Transaction
	Nonce        int64
	hashCache    string
}

func (b Block) Hash() string {
	data := ""
	for _, t := range b.Transactions {
		data += t.getHash()
	}
	data = b.PrevBlock + strconv.Itoa(b.Id) + strconv.FormatInt(b.Timestamp, 10) + data + strconv.FormatInt(b.Nonce, 10)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
	return hash
}

func (b *Block) addTransaction(t Transaction) {
	if t.Verify() {
		b.Transactions = append(b.Transactions, t)
	}
}

func (b Block) Check() bool {
	hash := b.Hash()
	for i := 0; i < 4; i++ {
		if hash[i] != '0' {
			return false
		}
	}
	return true
}
