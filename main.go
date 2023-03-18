package main

import (
	"github.com/gin-gonic/gin"
)

var blockchain *Blockchain

func initBlockchain() *Blockchain {
	blockchain := Blockchain{}

	blockchain.createGenesisBlock()

	var users []string
	for i := 0; i < 10; i++ {
		users = append(users, newUuid())
	}

	for _, user := range users {
		blockchain.addTx(Transaction{
			Id:        newUuid(),
			Sender:    god,
			Recipient: user,
			Amount:    10.0,
		})
	}

	return &blockchain
}

func main() {
	if !dbBlockchainExists() {
		blockchain = initBlockchain()
		dbSaveBlockchain(*blockchain)
	}

	router := gin.Default()
	router.POST("/transactions", apiAddTx)
	router.GET("/blockchain", apiGetChain)
	router.POST("/mine", apiMine)

	router.Run("localhost:8080")
}
