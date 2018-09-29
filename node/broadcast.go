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

// number of transactions for each block
const TransactionsNumber = 3

var port string

var urls []string

/*
 Sends object to nodes in url list
 msg - must be BlockChain or Transaction object
*/
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

/*
 Mines block of first n transactions and stops, when *stopSignal == true
*/
func miner(stopSignal *bool) {
	transQueue.Lock()
	b := blockchain.Block{PrevBlock: blockchain.GetLast().Hash(), Id: blockchain.GetLast().Id + 1, Timestamp: time.Now().Unix(), Nonce: 1}
	if *stopSignal == true {
		transQueue.Unlock()
		*stopSignal = false
		return
	}
	for i := 0; i < TransactionsNumber; i++ {
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
		color.New(color.BgGreen).Print("                     \n")
		color.New(color.BgGreen).Print(" I MADE NEW BLOCK!!! \n")
		color.New(color.BgGreen).Print("                     \n")
		fmt.Println()
		transQueue.Lock()
		transQueue.SetCurrent(&blockchain.Block{Id: -1})
		transQueue.Unlock()
		// Sends part of chain of the last day + new block to current node
		jsonStr, _ := json.Marshal(append(blockchain.GetSinceTime(time.Now().Unix()-(int64(time.Hour.Seconds())*24)).Sort(), b))
		req, _ := http.NewRequest("POST", "http://localhost:"+port+"/blocks/", bytes.NewBuffer(jsonStr))
		client := &http.Client{}
		client.Timeout = 0
		client.Do(req)
		return
	}
}

func main() {
	// Getting and checking [port] argument
	args := os.Args
	if len(args) != 1 {
		n, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Error 1: Port is not an integer.")
			fmt.Println(args[0] + " [port]")
			return
		}
		if !(n >= 1024 && n <= 65535) {
			fmt.Println("Error 2: Port is not in right range. \n[port] must be int between 1000 and 9999.")
			fmt.Println(args[0] + " [port]")
			return
		}
		port = args[1]
	} else {
		fmt.Println("Error 3: Port is not specified.")
		fmt.Println(args[0] + " [port]")
		return
	}
	// Retrieving url list from file
	file, err := os.Open("node/urls")
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	var stopSignal = false

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/blocks/", func(c *gin.Context) {
		c.JSON(200, blockchain.GetAll().Sort())
	})

	// Getting new part of blockchain
	router.POST("/blocks/", func(c *gin.Context) {
		var newChain blockchain.BlockChain
		err := c.ShouldBindJSON(&newChain)
		if err != nil || len(newChain) == 0 {
			return
		}
		// Part of local chain starting from first new chain's block timestamp
		todayChain := blockchain.GetSinceTime(newChain[0].Timestamp).Sort()
		if len(newChain) > len(todayChain) {
			if newChain.Check() {

				// If we got right chain with length more than ours
				go notifyNodes(newChain)
				// Stop miner
				stopSignal = true
				if (*transQueue.CurrentBlock).Id != -1 {
					for !stopSignal {
						// wait until miner goroutine return
					}
				}
				// Removing part from local and pushing new part
				for i := 0; i < len(todayChain); i++ {
					blockchain.RemoveBlock(todayChain[i].Hash())
					blockchain.PushBlock(newChain[i])
				}
				for i := len(todayChain); i < len(newChain); i++ {
					blockchain.PushBlock(newChain[i])
				}
				color.New(color.BgYellow).Println("GOT NEW PART OF CHAIN!")
				// Pushing all transactions from old part of chain back to queue
				for j := range todayChain {
					for _, i := range todayChain[j].Transactions {
						transQueue.Lock()
						transQueue.Push(i)
						transQueue.Unlock()
					}
				}
				// Removing ones which existing in new part
				for j := range newChain {
					for _, i := range newChain[j].Transactions {
						transQueue.Lock()
						transQueue.Remove(i.Sign)
						transQueue.Unlock()
					}
				}
				// If enough transactions returned back, start mining
				if transQueue.Size() >= TransactionsNumber && (*transQueue.CurrentBlock).Id == -1 {
					go miner(&stopSignal)
				}
			}
		}
	})

	// Handling transactions
	router.POST("/tran/", func(c *gin.Context) {
		var t blockchain.Transaction
		if err := c.ShouldBindJSON(&t); err != nil {
			return
		}
		if t.Verify() {
			transQueue.Lock()
			if transQueue.Push(t) {
				color.New(color.BgYellow).Println("New transaction!")
				go notifyNodes(t)
			}
			transQueue.Unlock()
			if transQueue.Size() >= TransactionsNumber && (*transQueue.CurrentBlock).Id == -1 {
				go miner(&stopSignal)
			}
		}
	})

	fmt.Println("RUNNING SERVER ON PORT :" + port)

	router.Run(":" + port)

}
