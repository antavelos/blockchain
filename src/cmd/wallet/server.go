package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func apiAddWallet(c *gin.Context) {
	wallet, err := CreateWallet()
	if err != nil {
		ErrorLogger.Printf("New wallet [FAIL]: %v", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to create a new wallet"})
		return
	}

	c.IndentedJSON(http.StatusCreated, wallet)
}

func InitRouter() *gin.Engine {
	router := gin.Default()

	router.SetTrustedProxies([]string{"localhost", "127.0.0.1"})

	router.POST("/wallets", apiAddWallet)

	return router
}