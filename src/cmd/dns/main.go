package main

import (
	"fmt"
	"net/http"

	bc "github.com/antavelos/blockchain/src/blockchain"

	"github.com/gin-gonic/gin"
)

var nodes []bc.Node = []bc.Node{
	{
		Host: "http://localhost:3001",
	},
	{
		Host: "http://localhost:3002",
	},
	// {
	// 	Host: "http://localhost:3003",
	// },
	// {
	// 	Host: "http://localhost:3004",
	// },
	// {
	// 	Host: "http://localhost:3005",
	// },
}

var host string = "localhost"
var port string = "3000"

func Nodes(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, nodes)
}

func initRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/nodes", Nodes)

	return router
}

func main() {
	router := initRouter()
	router.Run(fmt.Sprintf("%v:%v", host, port))
}
