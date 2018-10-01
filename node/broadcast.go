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
const MaxTransactionsNumber = 10

var port string

var urls []string

var stopSignal = false

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
	lastBlock := blockchain.GetLast()
	b := blockchain.Block{PrevBlock: lastBlock.Hash(), Id: lastBlock.Id + 1, Timestamp: time.Now().Unix(), Nonce: 1}
	if *stopSignal == true {
		transQueue.Unlock()
		*stopSignal = false
		return
	}
	for i := 0; i < MaxTransactionsNumber && i < transQueue.Size(); i++ {
		b.Transactions = append(b.Transactions, transQueue.Pop())
	}
	transQueue.SetCurrent(&b)
	transQueue.Unlock()
	for !b.Check() {
		if *stopSignal {
			transQueue.Lock()
			transQueue.SetCurrent(&blockchain.Block{Id: -1})
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
		color.New(color.BgGreen).Println("                     ")
		color.New(color.BgGreen).Println(" I MADE NEW BLOCK!!! ")
		color.New(color.BgGreen).Println("                     ")
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

func stopMiner() {
	stopSignal = true
	if (*transQueue.CurrentBlock).Id != -1 {
		for !stopSignal {
			// wait until miner goroutine return
		}
	}
}

func main() {
	// Getting and checking [port] argument
	args := os.Args
	if len(args) != 1 {
		n, err := strconv.Atoi(args[1])
		if err != nil {
			panic(getErrorDesc(1))
		}
		if !(n >= 1024 && n <= 65535) {
			panic(getErrorDesc(2))
		}
		port = args[1]
	} else {
		panic(getErrorDesc(3))
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

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/blocks/", func(c *gin.Context) {
		c.JSON(200, blockchain.GetAll().Sort())
	})

	router.GET("/blocks/:hash/", func(c *gin.Context) {
		hash, ok := c.Params.Get("hash")
		if !ok {
			c.AbortWithStatus(400)
		}
		if hash == "last" {
			c.JSON(200, blockchain.GetLast())
			return
		}
		if block, ok := blockchain.GetByHash(hash); ok {
			c.JSON(200, block)
		} else {
			c.JSON(404, block)
		}
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

				stopMiner()
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
				if transQueue.Size() != 0 && (*transQueue.CurrentBlock).Id == -1 {
					go miner(&stopSignal)
				}
			} else {
				color.New(color.BgRed).Println(c.ClientIP() + " has tried to send wrong chain.")
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
			if transQueue.Size() != 0 && (*transQueue.CurrentBlock).Id == -1 {
				go miner(&stopSignal)
			}
		}
	})

	fmt.Println("RUNNING SERVER ON PORT :" + port)

	router.Run(":" + port)

}
