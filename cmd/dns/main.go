package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	bc "github.com/antavelos/blockchain/pkg/blockchain"
	"github.com/antavelos/blockchain/pkg/db"

	"github.com/gin-gonic/gin"
)

var port string = os.Getenv("PORT")

func getNodeDb() *db.NodeDB {
	return &db.NodeDB{Filename: os.Getenv("NODES_FILENAME")}
}

func getNodes(c *gin.Context) {
	ndb := getNodeDb()

	nodes, err := ndb.LoadNodes()
	if err != nil {
		log.Printf("nodes not available: %v", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "nodes not available")
		return
	}

	c.IndentedJSON(http.StatusOK, nodes)
}

func addNode(c *gin.Context) {
	ndb := getNodeDb()

	var node bc.Node
	if err := c.BindJSON(&node); err != nil {
		log.Printf("invalid input: %v", err.Error())
		c.IndentedJSON(http.StatusBadRequest, "invalid input")
		return
	}

	err := ndb.AddNode(node)
	if err != nil {
		log.Printf("failed to add node: %v", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, node)
}

func initRouter() *gin.Engine {
	router := gin.Default()

	router.SetTrustedProxies([]string{"localhost", "127.0.0.1"})

	router.GET("/nodes", getNodes)
	router.POST("/nodes", addNode)

	return router
}

func main() {
	router := initRouter()
	router.Run(fmt.Sprintf(":%v", port))
}
