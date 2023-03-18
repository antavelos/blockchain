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

	tx, err := blockchain.addTx(tx)
	if err != nil {
		c.IndentedJSON(http.StatusOK, err.Error())
		return
	}

	err = dbSaveBlockchain(*blockchain)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "couldn't update blockchain")
		return
	}

	c.IndentedJSON(http.StatusCreated, tx)
}

func apiGetChain(c *gin.Context) {
	blockchain, err := dbLoadBlockchain()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "blockchain currently not available")
		return
	}

	result := map[string]any{
		"blockchain": blockchain,
		"blocksNum":  len(blockchain.Blocks),
		"isValid":    blockchain.isValid(),
	}

	c.IndentedJSON(http.StatusCreated, result)
}

func apiMine(c *gin.Context) {
	blockchain, err := dbLoadBlockchain()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "blockchain currently not available")
		return
	}

	block, err := blockchain.newBlock()
	if err != nil {
		c.IndentedJSON(http.StatusOK, err.Error())
		return
	}

	// TODO: to be done after network consensus
	blockchain.addBlock(block)
	err = dbSaveBlockchain(*blockchain)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "couldn't update blockchain")
		return
	}

	c.IndentedJSON(http.StatusCreated, block)
}
