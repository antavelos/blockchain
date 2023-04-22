package main

import (
	"net/http"

	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/db"
	nd "github.com/antavelos/blockchain/pkg/models/node"

	"github.com/gin-gonic/gin"
)

const nodesURI = "/nodes"

func getNodes(c *gin.Context) {
	ndb := db.GetNodeDb()

	nodes, err := ndb.LoadNodes()
	if err != nil {
		common.ErrorLogger.Printf("nodes not available"}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "nodes not available"})
		return
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func addNode(c *gin.Context) {
	ndb := db.GetNodeDb()

	var node nd.Node
	if err := c.BindJSON(&node); err != nil {
		common.ErrorLogger.Printf("invalid input"}
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	err := ndb.AddNode(node)
	if err != nil {
		common.ErrorLogger.Printf("failed to add node"}
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, node)
}

func InitRouter() *gin.Engine {
	router := gin.Default()
	router.SetTrustedProxies([]string{"localhost", "127.0.0.1"})
	router.GET(nodesURI, getNodes)
	router.POST(nodesURI, addNode)

	return router
}
