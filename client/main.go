package main

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-simple-blockchain/node/blockchain"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const version = "0.9"

var keys struct {
	privateKey *ecdsa.PrivateKey
	publicKey  string
}

func header() {
	fmt.Println("###############################")
	fmt.Println("# Go-simple-blockchain Client #")
	fmt.Println("# Version: " + version + "                #")
	fmt.Println("###############################")
	fmt.Println()
	if keys.publicKey != "" {
		fmt.Println(keys.publicKey)
		fmt.Println()
	}
}

/*
 Generates ecdsa keys, marshals and encodes in base64
*/
func generateKeys() (*ecdsa.PrivateKey, string, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		return &ecdsa.PrivateKey{}, "", err
	}
	s, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return &ecdsa.PrivateKey{}, "", err
	}
	ioutil.WriteFile("private_key", []byte(base64.StdEncoding.EncodeToString(s)), os.ModePerm)
	marshaledKey, _ := x509.MarshalPKIXPublicKey(privateKey.Public())
	publicKey := base64.StdEncoding.EncodeToString(marshaledKey)
	return privateKey, publicKey, nil
}

/*
 Returns transaction with current timestamp and sign
*/
func sign(msg string) blockchain.Transaction {
	t := blockchain.Transaction{Owner: keys.publicKey, Timestamp: time.Now().Unix(), Data: msg}
	hash := sha256.Sum256([]byte(t.Owner + strconv.FormatInt(t.Timestamp, 10) + t.Data))
	r, s, err := ecdsa.Sign(rand.Reader, keys.privateKey, hash[:])
	if err != nil {
		fmt.Println(err.Error())
	}
	asn, _ := asn1.Marshal([][]byte{r.Bytes(), s.Bytes()})
	t.Sign = base64.StdEncoding.EncodeToString(asn)
	return t
}

// Clear bash command
func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	file, err := ioutil.ReadFile("private_key")
	if err == nil {
		key, err := base64.StdEncoding.DecodeString(string(file))
		keys.privateKey, err = x509.ParseECPrivateKey(key)
		if err == nil {
			marshaledKey, _ := x509.MarshalPKIXPublicKey(keys.privateKey.Public())
			keys.publicKey = base64.StdEncoding.EncodeToString(marshaledKey)
		}
	}
	clear()
	header()
	for {

		fmt.Print("1. Generate new keys\n2. Make a transaction\n3. Exit\n\n")
		str, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		action, _ := strconv.Atoi(strings.TrimRight(str, "\n"))

		switch action {
		case 1:
			keys.privateKey, keys.publicKey, err = generateKeys()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		case 2:
			if keys.publicKey == "" {
				fmt.Print("You need to generate keys!\n\n")
				break
			}
			fmt.Print("Write transaction body:\n\n")
			str, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			str = strings.TrimRight(str, "\n")

			t := sign(str)

			// Marshal and send to specified node address
			// TODO: Another way to specify url

			url := "http://localhost:1488/tran/"

			jsonStr, _ := json.Marshal(t)
			req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
			client := &http.Client{}
			client.Do(req)

			fmt.Print("\nPress any key to continue...\n")
			bufio.NewReader(os.Stdin).ReadString('\n')
		case 3:
			return
		default:
			header()
		}
		clear()
		header()
	}

}
