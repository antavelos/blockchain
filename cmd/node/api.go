package main

import (
	"net/http"

	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/db"
	"github.com/antavelos/blockchain/pkg/lib/bus"
	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
	nd "github.com/antavelos/blockchain/pkg/models/node"

	"github.com/gin-gonic/gin"
)

const transactionsEndpoint = "/transactions"
const sharedTransactionsEndpoint = "/shared-transactions"
const sharedBlocksEndpoint = "/shared-blocks"
const pingEndpoint = "/ping"
const blockchainEndpoint = "/blockchain"

func addSharedBlock(c *gin.Context) {
	var block bc.Block

	if err := c.BindJSON(&block); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "invalid input"})
		return
	}

	block, err := ioAddBlock(block)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, block)
}

func addSharedTx(c *gin.Context) {
	var tx bc.Transaction

	if err := c.BindJSON(&tx); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "invalid input"})
		return
	}

	tx, err := ioAddTx(tx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, tx)
}

func addTx(c *gin.Context) {

	var tx bc.Transaction
	if err := c.BindJSON(&tx); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "invalid input"})
		return
	}

	if tx.Body.Sender == "" || tx.Body.Recipient == "" || tx.Body.Amount == 0.0 {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "invalid input"})
		return
	}

	tx, err := ioAddTx(tx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	bus.Publish(ShareTransaction, tx)

	c.IndentedJSON(http.StatusCreated, tx)
}

func getBlockchain(c *gin.Context) {
	bdb := db.GetBlockchainDb()

	blockchain, err := bdb.LoadBlockchain()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "blockchain currently not available"})
		return
	}

	c.IndentedJSON(http.StatusOK, *blockchain)
}

func ping(c *gin.Context) {
	ndb := db.GetNodeDb()

	var node nd.Node
	if err := c.BindJSON(&node); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	common.LogInfo("Ping from %#v", node.GetHost())

	err := ndb.AddNode(node)
	if err != nil {
		common.LogError(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	nodes, err := ndb.LoadNodes()
	if err != nil {
		common.LogError(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func InitRouter() *gin.Engine {
	router := gin.Default()

	router.SetTrustedProxies([]string{"localhost", "127.0.0.1"})

	router.POST(transactionsEndpoint, addTx)
	router.POST(sharedTransactionsEndpoint, addSharedTx)
	router.POST(sharedBlocksEndpoint, addSharedBlock)
	router.POST(pingEndpoint, ping)
	router.GET(blockchainEndpoint, getBlockchain)

	return router
}
