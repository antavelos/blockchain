package blockchain

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ApiAddTx(c *gin.Context) {
	var tx Transaction
	if err := c.BindJSON(&tx); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "invalid input")
		return
	}

	blockchain, err := DbLoadBlockchain()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "blockchain currently not available")
		return
	}

	tx, err = blockchain.AddTx(tx)
	if err != nil {
		c.IndentedJSON(http.StatusOK, err.Error())
		return
	}

	err = DbSaveBlockchain(*blockchain)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "couldn't update blockchain")
		return
	}

	c.IndentedJSON(http.StatusCreated, tx)
}

func ApiGetChain(c *gin.Context) {
	blockchain, err := DbLoadBlockchain()
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

func ApiMine(c *gin.Context) {
	blockchain, err := DbLoadBlockchain()
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
	err = DbSaveBlockchain(*blockchain)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "couldn't update blockchain")
		return
	}

	c.IndentedJSON(http.StatusCreated, block)
}

func ApiPing(c *gin.Context) {

	var node Node
	if err := c.BindJSON(&node); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "invalid input")
		return
	}
	log.Printf("ping from %#v", node.Host)

	err := DbAddNode(node)
	if err != nil {
		log.Println(err.Error())
	}

	var nodes []Node
	nodes, _ = DbLoadNodes()

	c.IndentedJSON(http.StatusOK, nodes)
}
