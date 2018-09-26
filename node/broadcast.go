package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"go-simple-blockchain/node/blockchain"
	"go-simple-blockchain/node/transQueue"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var port = "1488"

var urls = make([]string, 0)

func notifyNodes(msg interface{}) {
	postfix := ""
	switch msg.(type) {
	case blockchain.BlockChain:
		postfix = "/blocks/"
	case blockchain.Transaction:
		postfix = "/tran/"
	default:
		return
	}
	for _, url := range urls {
		jsonStr, _ := json.Marshal(msg)
		req, _ := http.NewRequest("POST", url+postfix, bytes.NewBuffer(jsonStr))
		client := &http.Client{}
		client.Timeout = 0
		client.Do(req)
	}
}

func miner(stopSignal *bool) {
	transQueue.Lock()
	b := blockchain.Block{PrevBlock: blockchain.GetLast().Hash(), Id: blockchain.GetLast().Id + 1, Timestamp: time.Now().Unix(), Nonce: 1}
	if *stopSignal != true {
		transQueue.Unlock()
		return
	}
	for i := 0; i < 3; i++ {
		b.Transactions = append(b.Transactions, transQueue.Pop())
	}
	transQueue.SetCurrent(&b)
	transQueue.Unlock()
	for !b.Check() {
		if *stopSignal {
			transQueue.Lock()
			transQueue.SetCurrent(&blockchain.Block{})
			for _, i := range b.Transactions {
				transQueue.Push(i)
			}
			transQueue.Unlock()
			*stopSignal = false
			return
		}
		b.Nonce++
	}
	if b.Check() {
		fmt.Println("BLOCK IS READY. IT'S HASH - " + b.Hash())
		transQueue.Lock()
		transQueue.SetCurrent(&blockchain.Block{})
		transQueue.Unlock()
		jsonStr, _ := json.Marshal(append(blockchain.GetAll().Sort(), b))
		req, _ := http.NewRequest("POST", "http://localhost:"+port+"/blocks/", bytes.NewBuffer(jsonStr))
		client := &http.Client{}
		client.Timeout = 0
		client.Do(req)
		return
	}
}

func main() {

	args := os.Args
	file, err := os.Open("node/urls")
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
	var stopSignal = false
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}
	if len(args) != 1 {
		n, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println(args[0] + " [port]")
			fmt.Println("ERROR! [port] must be int between 1000 and 9999!")
			return
		}
		if !(n > 999 && n < 10000) {
			fmt.Println(args[0] + " [port]")
			fmt.Println("ERROR! [port] must be int between 1000 and 9999!")
			return
		}
		port = args[1]
	}
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/blocks/", func(c *gin.Context) {
		c.JSON(200, blockchain.GetAll().Sort())
	})

	router.POST("/blocks/", func(c *gin.Context) {
		var newChain blockchain.BlockChain
		err := c.ShouldBindJSON(&newChain)
		if err != nil || len(newChain) == 0 {
			return
		}
		todayChain := blockchain.GetSinceTime(newChain[0].Timestamp).Sort()
		if len(newChain) > len(todayChain) {
			if newChain.Check() {
				go notifyNodes(newChain)
				stopSignal = true
				if len((*transQueue.CurrentBlock).Transactions) != 0 {
					for stopSignal {
						// wait until miner goroutine return
					}
				}
				for i := 0; i < len(todayChain); i++ {
					blockchain.RemoveBlock(todayChain[i].Hash())
					blockchain.PushBlock(newChain[i])
				}
				for i := len(todayChain); i < len(newChain); i++ {
					blockchain.PushBlock(newChain[i])
				}
				fmt.Println("GOT NEW PART OF CHAIN!")
				for j := range todayChain {
					for _, i := range todayChain[j].Transactions {
						transQueue.Lock()
						transQueue.Push(i)
						transQueue.Unlock()
					}
				}
				for j := range newChain {
					for _, i := range newChain[j].Transactions {
						transQueue.Lock()
						transQueue.Remove(i.Sign)
						transQueue.Unlock()
					}
				}
				if transQueue.Size() >= 3 && len((*transQueue.CurrentBlock).Transactions) > 0 {
					go miner(&stopSignal)
				}
			}
		}
	})

	router.POST("/tran/", func(c *gin.Context) {
		var t blockchain.Transaction
		if err := c.ShouldBindJSON(&t); err != nil {
			return
		}
		if t.Verify() {
			transQueue.Lock()
			if transQueue.Push(t) {
				color.New(color.BgGreen).Println("New transaction!")
				go notifyNodes(t)
			}
			transQueue.Unlock()
			if transQueue.Size() >= 3 && len((*transQueue.CurrentBlock).Transactions) == 0 {
				go miner(&stopSignal)
			}
		}
	})

	fmt.Println("RUNNING SERVER ON PORT :" + port)

	router.Run(":" + port)

}
