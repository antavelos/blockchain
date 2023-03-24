package main

import (
	"log"
	"net/http"

	bc "github.com/antavelos/blockchain"

	"github.com/gin-gonic/gin"
)

func apiAddTx(c *gin.Context) {
	var tx bc.Transaction
	if err := c.BindJSON(&tx); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "invalid input")
		return
	}

	blockchain, err := ioLoadBlockchain()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "blockchain currently not available")
		return
	}

	tx, err = blockchain.AddTx(tx)
	if err != nil {
		c.IndentedJSON(http.StatusOK, err.Error())
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
	// nodes, _ := DbLoadNodes()
	// if err != nil {
	// 	return errors.New("nodes list not available")
	// }

	// result := map[string]any{
	// 	"blockchain": *blockchain,
	// 	"blocksNum":  len(blockchain.Blocks),
	// 	"isValid":    isValid(*blockchain),
	// 	"nodes":      nodes,
	// }

	c.IndentedJSON(http.StatusOK, *blockchain)
}

func apiMine(c *gin.Context) {
	blockchain, err := ioLoadBlockchain()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "blockchain currently not available")
		return
	}

	block, err := blockchain.NewBlock()
	if err != nil {
		c.IndentedJSON(http.StatusOK, err.Error())
		return
	}

	// TODO: to be done after network consensus
	blockchain.AddBlock(block)
	err = ioSaveBlockchain(*blockchain)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "couldn't update blockchain")
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
	log.Printf("ping from %#v", node.Host)

	err := ioAddNode(node)
	if err != nil {
		log.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	nodes, err := ioLoadNodes()
	if err != nil {
		log.Println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func initRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/blockchain", apiGetBlockchain)
	router.POST("/transactions", apiAddTx)
	router.POST("/mine", apiMine)
	router.POST("/ping", apiPing)

	return router
}
