package main

import (
	"net/http"

	nd "github.com/antavelos/blockchain/src/internal/pkg/models/node"
	"github.com/antavelos/blockchain/src/pkg/utils"

	"github.com/gin-gonic/gin"
)

const nodesURI = "/nodes"

func getNodes(c *gin.Context) {
	nrepo := getNodeRepo()

	nodes, err := nrepo.GetNodes()
	if err != nil {
		utils.LogError("nodes not available", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "nodes not available"})
		return
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func addNode(c *gin.Context) {
	nrepo := getNodeRepo()

	var node nd.Node
	if err := c.BindJSON(&node); err != nil {
		utils.LogError("invalid input", err.Error())
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	err := nrepo.AddNode(node)
	if err != nil {
		utils.LogError("failed to add node", err.Error())
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
