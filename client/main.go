package main

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const version = "0.1"

var keys struct {
	privateKey *ecdsa.PrivateKey
	publicKey  string
}

type Transaction struct {
	Owner     string
	Timestamp int64
	Data      string
	Sign      string
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

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	file, err := ioutil.ReadFile("private_key")
	if err == nil {
		keys.privateKey, err = x509.ParseECPrivateKey(file)
		if err == nil {
			keys.publicKey = base64.StdEncoding.EncodeToString([]byte(keys.privateKey.PublicKey.X.Text(16) + "+" + keys.privateKey.PublicKey.Y.Text(16)))
			fmt.Println(keys.privateKey.PublicKey.X.Text(16) + "+" + keys.privateKey.PublicKey.Y.Text(16))
		}
	}
	clear()
	header()
	for {

		fmt.Print("1. Generate new keys\n2. Make a transaction\n3. Exit\n\n")
		str, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		action, err := strconv.Atoi(strings.TrimRight(str, "\n"))

		if err != nil {
			clear()
			continue
		}

		switch action {
		case 1:
			keys.privateKey, _ = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
			s, _ := x509.MarshalECPrivateKey(keys.privateKey)
			ioutil.WriteFile("private_key", s, os.ModePerm)
			p := keys.privateKey.PublicKey.X.Text(16) + "+" + keys.privateKey.PublicKey.Y.Text(16)
			keys.publicKey = base64.StdEncoding.EncodeToString([]byte(p))
			clear()
			header()
		case 2:
			clear()
			header()
			if keys.publicKey == "" {
				fmt.Print("You need to generate keys!\n\n")
				break
			}
			fmt.Print("Write transaction body:\n\n")
			str, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			str = strings.TrimRight(str, "\n")

			t := Transaction{keys.publicKey, time.Now().Unix(), str, ""}
			hash := sha256.Sum256([]byte(t.Owner + strconv.FormatInt(t.Timestamp, 10) + t.Data))
			r, s, _ := ecdsa.Sign(rand.Reader, keys.privateKey, hash[:])

			t.Sign = base64.StdEncoding.EncodeToString([]byte(r.Text(16) + "_" + s.Text(16)))

			url := "http://localhost:1488/tran/"

			jsonStr, _ := json.Marshal(t)
			req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			client.Do(req)

			fmt.Print("\nPress any key to continue...\n")
			bufio.NewReader(os.Stdin).ReadString('\n')
		case 3:
			return
		}
	}

}
