package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var blockchain *Blockchain

func printBlockchain(bc Blockchain) {
	marshaled, _ := json.MarshalIndent(bc, "", "  ")
	log.Println(string(marshaled))
}

func getUuid() string {
	return fmt.Sprintf("%v", uuid.New())
}

func addBlock(bc *Blockchain) {
	block, err := bc.newBlock()
	if err == nil {
		bc.addBlock(block)
	} else {
		log.Println(err)
	}
}

func initBlockchain() *Blockchain {
	blockchain := Blockchain{}

	blockchain.createGenesisBlock()

	var users []string
	for i := 0; i < 10; i++ {
		users = append(users, getUuid())
	}

	for _, user := range users {
		blockchain.addTx(Transaction{
			Id:        getUuid(),
			Sender:    god,
			Recipient: user,
			Amount:    10.0,
		})
	}

	return &blockchain
}

func main() {

	blockchain = initBlockchain()

	router := gin.Default()
	router.POST("/transactions", apiAddTx)

	router.Run("localhost:8080")
}
