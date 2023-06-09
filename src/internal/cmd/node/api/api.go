package api

import (
	"net/http"

	"github.com/antavelos/blockchain/src/internal/cmd/node/events"
	bc "github.com/antavelos/blockchain/src/internal/pkg/models/blockchain"
	nd "github.com/antavelos/blockchain/src/internal/pkg/models/node"
	rep "github.com/antavelos/blockchain/src/internal/pkg/repos"
	"github.com/antavelos/blockchain/src/pkg/eventbus"
	"github.com/antavelos/blockchain/src/pkg/utils"

	"github.com/gin-gonic/gin"
)

const transactionsEndpoint = "/transactions"
const sharedTransactionsEndpoint = "/shared-transactions"
const sharedBlocksEndpoint = "/shared-blocks"
const pingEndpoint = "/ping"
const blockchainEndpoint = "/blockchain"

type RouteHandler struct {
	Bus   *eventbus.Bus
	Repos *rep.Repos
}

func NewRouteHandler(bus *eventbus.Bus, repos *rep.Repos) *RouteHandler {
	return &RouteHandler{Bus: bus, Repos: repos}
}

func (h *RouteHandler) addSharedBlock(c *gin.Context) {
	var block bc.Block

	if err := c.BindJSON(&block); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	err := h.Repos.BlockchainRepo.AddBlock(block)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, block)
}

func (h *RouteHandler) addSharedTx(c *gin.Context) {
	var tx bc.Transaction

	if err := c.BindJSON(&tx); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	tx, err := h.Repos.BlockchainRepo.AddTx(tx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, tx)
}

func (h *RouteHandler) addTx(c *gin.Context) {

	var tx bc.Transaction
	if err := c.BindJSON(&tx); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	if tx.Body.Sender == "" || tx.Body.Recipient == "" || tx.Body.Amount == 0.0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	tx, err := h.Repos.BlockchainRepo.AddTx(tx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.Bus.Handle(eventbus.DataEvent{Ev: events.TransactionReceivedEvent, Data: tx})

	c.IndentedJSON(http.StatusCreated, tx)
}

func (h *RouteHandler) getBlockchain(c *gin.Context) {
	blockchain, err := h.Repos.BlockchainRepo.GetBlockchain()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "blockchain currently not available"})
		return
	}

	c.IndentedJSON(http.StatusOK, *blockchain)
}

func (h *RouteHandler) ping(c *gin.Context) {
	var node nd.Node
	if err := c.BindJSON(&node); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	utils.LogInfo("Ping from", node.GetHost())

	err := h.Repos.NodeRepo.AddNode(node)
	if err != nil {
		utils.LogError(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	nodes, err := h.Repos.NodeRepo.GetNodes()
	if err != nil {
		utils.LogError(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func (routeHandler *RouteHandler) InitRouter() *gin.Engine {
	router := gin.Default()

	router.SetTrustedProxies([]string{"localhost", "127.0.0.1"})

	router.POST(transactionsEndpoint, routeHandler.addTx)
	router.POST(sharedTransactionsEndpoint, routeHandler.addSharedTx)
	router.POST(sharedBlocksEndpoint, routeHandler.addSharedBlock)
	router.POST(pingEndpoint, routeHandler.ping)
	router.GET(blockchainEndpoint, routeHandler.getBlockchain)

	return router
}
