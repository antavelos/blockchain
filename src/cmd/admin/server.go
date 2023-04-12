package main

import (
	// "encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

func index(c *gin.Context) {
	// bdb := getBlockchainDb()

	// blockchain, err := bdb.LoadBlockchain()
	// if err != nil {
	// 	ErrorLogger.Println(err.Error())
	// 	c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	// marshalled, err := json.MarshalIndent(blockchain, "", "  ")
	// if err != nil {
	// 	ErrorLogger.Println(err.Error())
	// 	c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Blockchain",
		// "chain": string(marshalled),
	})
}

func initRouter() *gin.Engine {
	router := gin.Default()

	router.SetTrustedProxies([]string{"localhost", "127.0.0.1"})

	router.LoadHTMLGlob("cmd/admin/templates/*")

	router.GET("/", index)

	return router
}
