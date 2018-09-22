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
}

func (t Transaction) getHash() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(t.Owner+t.Data+t.Sign)))
}

func (t Transaction) Verify() bool {
	hash := sha256.Sum256([]byte(t.Owner + strconv.FormatInt(t.Timestamp, 10) + t.Data))
	owner, err := base64.StdEncoding.DecodeString(t.Owner)
	if err != nil {
		return false
	}

	pub_arr := strings.Split(string(owner), "+")
	if len(pub_arr) != 2 {
		return false
	}
	X, _ := new(big.Int).SetString(pub_arr[0], 16)
	Y, _ := new(big.Int).SetString(pub_arr[1], 16)
	public_key := ecdsa.PublicKey{elliptic.P224(), X, Y}
	if err != nil {
		return false
	}

	sign, err := base64.StdEncoding.DecodeString(t.Sign)

	arr := strings.Split(string(sign), "_")

	r, s := new(big.Int), new(big.Int)
	r, _ = r.SetString(arr[0], 16)
	s, _ = s.SetString(arr[1], 16)

	if err != nil {
		return false
	}

	return ecdsa.Verify(&public_key, hash[:], r, s)
}
