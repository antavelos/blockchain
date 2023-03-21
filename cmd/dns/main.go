package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var addresses []string = []string{
	"http://localhost:3001",
	"http://localhost:3002",
	"http://localhost:3003",
	"http://localhost:3004",
	"http://localhost:3005",
}

var host string = "localhost"
var port string = "3000"

func getAddresses(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, addresses)
}

func initRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/addresses", getAddresses)

	return router
}

func main() {
	router := initRouter()
	router.Run(fmt.Sprintf("%v:%v", host, port))
}
