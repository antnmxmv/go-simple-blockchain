package blockchain

import (
	"crypto/sha256"
	"fmt"
)

type Transaction struct {
	Owner     string
	Timestamp int
	Data      string
	Sign      string
}

func (t Transaction) getHash() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(t.Owner+t.Data+t.Sign)))
}

func (t Transaction) Verify() bool {
	// TODO: implement
	return true
}
