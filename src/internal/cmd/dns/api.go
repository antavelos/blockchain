package main

import (
	"net/http"

	nd "github.com/antavelos/blockchain/src/internal/pkg/models/node"
	"github.com/antavelos/blockchain/src/internal/pkg/repos"
	"github.com/antavelos/blockchain/src/pkg/utils"

	"github.com/gin-gonic/gin"
)

const nodesURI = "/nodes"

type RouteHandler struct {
	NodeRepo repos.NodeRepo
}

func NewRouteHandler(nodeRepo *repos.NodeRepo) *RouteHandler {
	return &RouteHandler{NodeRepo: *nodeRepo}
}

func (h RouteHandler) getNodes(c *gin.Context) {
	nodes, err := h.NodeRepo.GetNodes()
	if err != nil {
		utils.LogError("nodes not available", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "nodes not available"})
		return
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func (h RouteHandler) addNode(c *gin.Context) {
	var node nd.Node
	if err := c.BindJSON(&node); err != nil {
		utils.LogError("invalid input", err.Error())
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	err := h.NodeRepo.AddNode(node)
	if err != nil {
		utils.LogError("failed to add node", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, node)
}

func (h *RouteHandler) InitRouter() *gin.Engine {
	router := gin.Default()
	router.SetTrustedProxies([]string{"localhost", "127.0.0.1"})
	router.GET(nodesURI, h.getNodes)
	router.POST(nodesURI, h.addNode)

	return router
}
