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
