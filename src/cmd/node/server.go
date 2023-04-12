package main

import (
	"fmt"
	"net/http"
	"strings"

	bc "github.com/antavelos/blockchain/src/blockchain"

	"github.com/gin-gonic/gin"
)

const indexURL = "/"
const transactionsURL = "/transactions"
const sharedTransactionsURL = "/shared-transactions"
const sharedBlocksURL = "/shared-blocks"
const pingURL = "/ping"
const blockchainURL = "/blockchain"
const mineURL = "/mine"
const resolveURL = "/resolve"

func apiAddSharedBlock(c *gin.Context) {
	var block bc.Block

	if err := c.BindJSON(&block); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "invalid input")
		return
	}

	block, err := ioAddBlock(block)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusCreated, block)
}

func apiAddSharedTx(c *gin.Context) {
	var tx bc.Transaction

	if err := c.BindJSON(&tx); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "invalid input")
		return
	}

	tx, err := ioAddTx(tx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusCreated, tx)
}

func apiAddTx(c *gin.Context) {

	var tx bc.Transaction
	if err := c.BindJSON(&tx); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "invalid input")
		return
	}

	if tx.Body.Sender == "" || tx.Body.Recipient == "" || tx.Body.Amount == 0.0 {
		c.IndentedJSON(http.StatusInternalServerError, "invalid input")
		return
	}

	tx, err := ioAddTx(tx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	if nodeErrors := ShareTx(tx); nodeErrors != nil {
		errorStrings := ErrorsToStrings(nodeErrors)
		if len(errorStrings) > 0 {
			ErrorLogger.Printf("Failed to share the transaction with other nodes: \n%v", strings.Join(errorStrings, "\n"))
		}
	}

	c.IndentedJSON(http.StatusCreated, tx)
}

func apiGetBlockchain(c *gin.Context) {
	bdb := getBlockchainDb()

	blockchain, err := bdb.LoadBlockchain()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "blockchain currently not available")
		return
	}

	c.IndentedJSON(http.StatusOK, *blockchain)
}

func apiMine(c *gin.Context) {

	block, err := Mine()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusCreated, block)
}

func apiPing(c *gin.Context) {
	ndb := getNodeDb()

	var node bc.Node
	if err := c.BindJSON(&node); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	InfoLogger.Printf("Ping from %#v", node.Host)

	err := ndb.AddNode(node)
	if err != nil {
		ErrorLogger.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	nodes, err := ndb.LoadNodes()
	if err != nil {
		ErrorLogger.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func apiResolve(c *gin.Context) {
	bdb := getBlockchainDb()

	// TODO: refactor

	err := ResolveLongestBlockchain()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, fmt.Sprintf("failed to resolve blockchain: %v", err.Error()))
		return
	}

	blockchain, err := bdb.LoadBlockchain()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "blockchain currently not available")
		return
	}

	c.IndentedJSON(http.StatusOK, *blockchain)
}

func initRouter() *gin.Engine {
	router := gin.Default()

	router.SetTrustedProxies([]string{"localhost", "127.0.0.1"})

	router.POST(transactionsURL, apiAddTx)
	router.POST(sharedTransactionsURL, apiAddSharedTx)
	router.POST(sharedBlocksURL, apiAddSharedBlock)
	router.POST(pingURL, apiPing)
	router.GET(blockchainURL, apiGetBlockchain)
	router.GET(mineURL, apiMine)
	router.GET(resolveURL, apiResolve)

	return router
}
