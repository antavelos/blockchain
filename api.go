package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func apiAddTx(c *gin.Context) {
	var tx Transaction
	if err := c.BindJSON(&tx); err != nil {
		return
	}

	blockchain, _ := dbLoadBlockchain()
	blockchain.addTx(tx)
	dbSaveBlockchain(*blockchain)

	c.IndentedJSON(http.StatusCreated, tx)
}

func apiGetChain(c *gin.Context) {
	blockchain, _ := dbLoadBlockchain()
	c.IndentedJSON(http.StatusCreated, blockchain)
}

func apiMine(c *gin.Context) {
	blockchain, _ := dbLoadBlockchain()

	block, err := blockchain.newBlock()
	if err != nil {
		c.IndentedJSON(http.StatusOK, err.Error())
	} else {
		blockchain.addBlock(block)
		dbSaveBlockchain(*blockchain)

		c.IndentedJSON(http.StatusCreated, block)
	}
}
