package api

import (
	"net/http"

	"github.com/antavelos/blockchain/src/internal/cmd/node/events"
	bc "github.com/antavelos/blockchain/src/internal/pkg/models/blockchain"
	nd "github.com/antavelos/blockchain/src/internal/pkg/models/node"
	"github.com/antavelos/blockchain/src/pkg/eventbus"
	"github.com/antavelos/blockchain/src/pkg/utils"

	bc_repo "github.com/antavelos/blockchain/src/internal/pkg/repos/blockchain"
	node_repo "github.com/antavelos/blockchain/src/internal/pkg/repos/node"
	"github.com/gin-gonic/gin"
)

const transactionsEndpoint = "/transactions"
const sharedTransactionsEndpoint = "/shared-transactions"
const sharedBlocksEndpoint = "/shared-blocks"
const pingEndpoint = "/ping"
const blockchainEndpoint = "/blockchain"

type RouteHandler struct {
	Bus            *eventbus.Bus
	BlockchainRepo *bc_repo.BlockchainRepo
	NodeRepo       *node_repo.NodeRepo
}

func NewRouteHandler(bus *eventbus.Bus, br *bc_repo.BlockchainRepo, nr *node_repo.NodeRepo) *RouteHandler {
	return &RouteHandler{Bus: bus, BlockchainRepo: br, NodeRepo: nr}
}

func (h *RouteHandler) addSharedBlock(c *gin.Context) {
	var block bc.Block

	if err := c.BindJSON(&block); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	err := h.BlockchainRepo.AddBlock(block)
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

	tx, err := h.BlockchainRepo.AddTx(tx)
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

	tx, err := h.BlockchainRepo.AddTx(tx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.Bus.Handle(events.TransactionReceivedEvent{Tx: tx})

	c.IndentedJSON(http.StatusCreated, tx)
}

func (h *RouteHandler) getBlockchain(c *gin.Context) {
	blockchain, err := h.BlockchainRepo.GetBlockchain()
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

	err := h.NodeRepo.AddNode(node)
	if err != nil {
		utils.LogError(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	nodes, err := h.NodeRepo.GetNodes()
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
