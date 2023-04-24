package main

import (
	"net/http"

	"github.com/antavelos/blockchain/pkg/common"
	nd "github.com/antavelos/blockchain/pkg/models/node"

	"github.com/gin-gonic/gin"
)

const nodesURI = "/nodes"

func getNodes(c *gin.Context) {
	ndb := getNodeDb()

	nodes, err := ndb.LoadNodes()
	if err != nil {
		common.LogError("nodes not available", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "nodes not available"})
		return
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func addNode(c *gin.Context) {
	ndb := getNodeDb()

	var node nd.Node
	if err := c.BindJSON(&node); err != nil {
		common.LogError("invalid input", err.Error())
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	err := ndb.AddNode(node)
	if err != nil {
		common.LogError("failed to add node", err.Error())
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
