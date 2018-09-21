package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-simple-blockchain/node/blockchain"
	"sort"
	"sync"
	"time"
)

var Trans struct {
	Queue []blockchain.Transaction
	Mux   sync.Mutex
}

func miner() {
	for {
		time.Sleep(time.Second * 10)
		if len(Trans.Queue) >= 3 {
			Trans.Mux.Lock()
			sort.Slice(Trans.Queue, func(i, j int) bool {
				return Trans.Queue[i].Timestamp < Trans.Queue[j].Timestamp
			})
			arr := Trans.Queue[:3]

			Trans.Queue = Trans.Queue[3:]

			Trans.Mux.Unlock()

			b := blockchain.Block{blockchain.GetLast().Hash(), blockchain.GetLast().Id + 1, time.Now().Unix(), arr, 1}

			fmt.Println("start mining...")

			for !b.Check() {
				b.Nonce++
			}

			fmt.Println("got block! it's hash - " + b.Hash())

			blockchain.SendBlock(b)
		}
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/blocks/", func(c *gin.Context) {
		c.JSON(200, blockchain.GetAll())
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
		}
	})

	go miner()

	router.Run(":1488")

}
