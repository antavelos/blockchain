package main

import (
	"net/http"

	bc "github.com/antavelos/blockchain/src/internal/pkg/models/blockchain"
	nd "github.com/antavelos/blockchain/src/internal/pkg/models/node"
	"github.com/antavelos/blockchain/src/pkg/bus"
	"github.com/antavelos/blockchain/src/pkg/utils"

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
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	brepo := getBlockchainRepo()
	err := brepo.AddBlock(block)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, block)
}

func addSharedTx(c *gin.Context) {
	var tx bc.Transaction

	if err := c.BindJSON(&tx); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	brepo := getBlockchainRepo()
	tx, err := brepo.AddTx(tx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, tx)
}

func addTx(c *gin.Context) {

	var tx bc.Transaction
	if err := c.BindJSON(&tx); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if tx.Body.Sender == "" || tx.Body.Recipient == "" || tx.Body.Amount == 0.0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	brepo := getBlockchainRepo()
	tx, err := brepo.AddTx(tx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	bus.Publish(ShareTransactionTopic, tx)

	c.IndentedJSON(http.StatusCreated, tx)
}

func getBlockchain(c *gin.Context) {
	brepo := getBlockchainRepo()

	blockchain, err := brepo.GetBlockchain()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "blockchain currently not available"})
		return
	}

	c.IndentedJSON(http.StatusOK, *blockchain)
}

func ping(c *gin.Context) {
	nrepo := getNodeRepo()

	var node nd.Node
	if err := c.BindJSON(&node); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	utils.LogInfo("Ping from", node.GetHost())

	err := nrepo.AddNode(node)
	if err != nil {
		utils.LogError(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	nodes, err := nrepo.GetNodes()
	if err != nil {
		utils.LogError(err.Error())
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
