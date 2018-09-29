package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"fmt"
	"math/big"
	"strconv"
)

type Transaction struct {
	Owner     string
	Timestamp int64
	Data      string
	Sign      string
	hashCache string
}

/*
 Need for hashing block
*/
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
	pbKey, err := x509.ParsePKIXPublicKey(owner)
	if err != nil {
		fmt.Println(err)
		return false
	}
	publicKey := ecdsa.PublicKey{Curve: elliptic.P224(), X: pbKey.(*ecdsa.PublicKey).X, Y: pbKey.(*ecdsa.PublicKey).Y}
	if err != nil {
		return false
	}

	sign, err := base64.StdEncoding.DecodeString(t.Sign)
	if err != nil {
		return false
	}
	var rs [][]byte
	_, err = asn1.Unmarshal(sign, &rs)
	if err != nil {
		return false
	}
	r, s := big.NewInt(1).SetBytes(rs[0]), big.NewInt(1).SetBytes(rs[1])
	return ecdsa.Verify(&publicKey, hash[:], r, s)
}
