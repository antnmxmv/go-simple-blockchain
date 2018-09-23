package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type Transaction struct {
	Owner     string
	Timestamp int64
	Data      string
	Sign      string
	hashCache string
}

func (t *Transaction) getHash() string {
	if t.hashCache == "" {
		t.hashCache = fmt.Sprintf("%x", sha256.Sum256([]byte(t.Owner+t.Data+t.Sign)))
	}
	return t.hashCache
}

func (t Transaction) Verify() bool {
	hash := sha256.Sum256([]byte(t.Owner + strconv.FormatInt(t.Timestamp, 10) + t.Data))
	owner, err := base64.StdEncoding.DecodeString(t.Owner)
	if err != nil {
		return false
	}
	pubArr := strings.Split(string(owner), "+")
	if len(pubArr) != 2 {
		return false
	}
	X, ok1 := new(big.Int).SetString(pubArr[0], 16)
	Y, ok2 := new(big.Int).SetString(pubArr[1], 16)
	if !ok1 || !ok2 {
		return false
	}
	publicKey := ecdsa.PublicKey{Curve: elliptic.P224(), X: X, Y: Y}
	if err != nil {
		return false
	}

	sign, err := base64.StdEncoding.DecodeString(t.Sign)
	if err != nil {
		return false
	}
	arr := strings.Split(string(sign), "_")
	r, s := new(big.Int), new(big.Int)
	r, ok1 = r.SetString(arr[0], 16)
	s, ok2 = s.SetString(arr[1], 16)
	if !ok1 || !ok2 {
		return false
	}
	return ecdsa.Verify(&publicKey, hash[:], r, s)
}
