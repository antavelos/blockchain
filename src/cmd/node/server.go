package main

import (
	"encoding/json"
	"net/http"
	"sync"

	bc "github.com/antavelos/blockchain/src/blockchain"

	"github.com/gin-gonic/gin"
)

func apiAddTx(c *gin.Context) {
	m := sync.Mutex{}

	var tx bc.Transaction
	if err := c.BindJSON(&tx); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "invalid input")
		return
	}

	if tx.Body.Sender == "" || tx.Body.Recipient == "" || tx.Body.Amount == 0.0 {
		c.IndentedJSON(http.StatusInternalServerError, "invalid input")
		return
	}

	m.Lock()
	defer m.Unlock()

	blockchain, err := ioLoadBlockchain()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "blockchain currently not available")
		return
	}

	tx, err = blockchain.AddTx(tx)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := ioSaveBlockchain(*blockchain); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "couldn't update blockchain")
		return
	}

	c.IndentedJSON(http.StatusCreated, tx)
}

func apiGetBlockchain(c *gin.Context) {
	blockchain, err := ioLoadBlockchain()
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

	var node bc.Node
	if err := c.BindJSON(&node); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	InfoLogger.Printf("Ping from %#v", node.Host)

	err := ioAddNode(node)
	if err != nil {
		ErrorLogger.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	nodes, err := ioLoadNodes()
	if err != nil {
		ErrorLogger.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func index(c *gin.Context) {
	blockchain, err := ioLoadBlockchain()
	if err != nil {
		ErrorLogger.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	marshalled, err := json.MarshalIndent(blockchain, "", "  ")
	if err != nil {
		ErrorLogger.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title": "Blockchain",
		"chain": string(marshalled),
	})
}

func apiResolve(c *gin.Context) {
	nodes, err := ioLoadNodes()
	if err != nil {
		ErrorLogger.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resolveLongestBlockchain(nodes)

	blockchain, err := ioLoadBlockchain()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "blockchain currently not available")
		return
	}

	c.IndentedJSON(http.StatusOK, *blockchain)
}

func initRouter() *gin.Engine {
	router := gin.Default()

	router.SetTrustedProxies([]string{"localhost", "127.0.0.1"})

	router.LoadHTMLGlob("cmd/node/templates/*")

	router.POST("/transactions", apiAddTx)
	router.POST("/ping", apiPing)
	router.GET("/", index)
	router.GET("/blockchain", apiGetBlockchain)
	router.GET("/mine", apiMine)
	router.GET("/resolve", apiResolve)

	return router
}
