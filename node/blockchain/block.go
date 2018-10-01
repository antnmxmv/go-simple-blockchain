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

/*
 Block stores whole transactions, but hashes only hashes of all transactions
*/
func (b *Block) Hash() string {
	data := ""
	if b.hashCache == "" {
		for _, t := range b.Transactions {
			data += t.getHash()
		}
		b.hashCache = data
	} else {
		data = b.hashCache
	}
	data = b.PrevBlock + strconv.Itoa(b.Id) + strconv.FormatInt(b.Timestamp, 10) + data + strconv.FormatInt(b.Nonce, 10)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
	return hash
}

func (b Block) Check() bool {
	hash := b.Hash()
	for i := 0; i < 5; i++ {
		if hash[i] != '0' {
			return false
		}
	}
	return true
}

func (a Block) Equal(b Block) bool {
	return b.Hash() == a.Hash()
}
