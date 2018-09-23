package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"go-simple-blockchain/node/blockchain"
	"net/http"
	"sort"
	"sync"
	"time"
)

var urls = []string{"http://127.0.0.1:1488"}

var Trans struct {
	Queue []blockchain.Transaction
	Mux   sync.Mutex
}

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
		client.Timeout = time.Second
		client.Do(req)
	}
}

func miner(n int) {
	for {
		if len(Trans.Queue) >= n {
			Trans.Mux.Lock()
			sort.Slice(Trans.Queue, func(i, j int) bool {
				return Trans.Queue[i].Timestamp < Trans.Queue[j].Timestamp
			})
			arr := Trans.Queue[:n]
			Trans.Queue = Trans.Queue[n:]
			Trans.Mux.Unlock()
			b := blockchain.Block{PrevBlock: blockchain.GetLast().Hash(), Id: blockchain.GetLast().Id + 1, Timestamp: time.Now().Unix(), Transactions: arr, Nonce: 1}
			fmt.Printf("GOT %d TRANSACTIONS. START MINING\n", n)
			for !b.Check() {
				b.Nonce++
			}
			fmt.Println("BLOCK IS READY. IT'S HASH - " + b.Hash())
			blockchain.PushBlock(b)
			notifyNodes(blockchain.GetAfterTime(time.Now().Unix() - int64(time.Hour.Seconds())*24).Sort())
		}
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/blocks/", func(c *gin.Context) {
		c.JSON(200, blockchain.GetAll().Sort())
	})

	router.POST("/blocks/", func(c *gin.Context) {
		var newChain blockchain.BlockChain
		err := c.ShouldBindJSON(&newChain)
		if err != nil {
			return
		}
		todayChain := blockchain.GetAfterTime(time.Now().Unix() - int64(time.Hour.Seconds())*24).Sort()
		if len(newChain) > len(todayChain) {
			if newChain.Check() {
				for i := 0; i < len(todayChain); i++ {
					if todayChain[i].Timestamp != newChain[i].Timestamp {
						blockchain.RemoveBlock(todayChain[i].Hash())
						blockchain.PushBlock(newChain[i])
					}
				}
				for i := len(todayChain); i < len(newChain); i++ {
					blockchain.PushBlock(newChain[i])
				}
				fmt.Println("GOT NEW PART OF CHAIN!")
			}
		}
	})

	router.POST("/tran/", func(c *gin.Context) {
		var t blockchain.Transaction
		if err := c.ShouldBindJSON(&t); err != nil {
			return
		}
		if t.Verify() {
			Trans.Mux.Lock()
			Trans.Queue = append(Trans.Queue, t)
			Trans.Mux.Unlock()
			fmt.Println(color.GreenString("Transaction submitted successfully!"))
		} else {
			fmt.Println(color.RedString("Transaction not submitted!"))
		}
	})

	go miner(3)

	router.Run(":2228")

}
