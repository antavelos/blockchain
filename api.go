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

	blockchain.addTx(tx)

	c.IndentedJSON(http.StatusCreated, tx)
}

func apiChain(c *gin.Context) {
	c.IndentedJSON(http.StatusCreated, blockchain)
}
