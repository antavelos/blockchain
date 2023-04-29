package main

import (
	"net/http"

	"github.com/antavelos/blockchain/src/pkg/utils"
	"github.com/gin-gonic/gin"
)

const NewWalletEndpoint = "/wallets/new"

func apiNewWallet(c *gin.Context) {
	wrepo := getWalletRepo()

	wallet, err := wrepo.CreateWallet()
	if err != nil {
		utils.LogError("New wallet [FAIL]", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to create new wallet"})
		return
	}

	c.IndentedJSON(http.StatusCreated, wallet)
}

func InitRouter() *gin.Engine {
	router := gin.Default()

	router.SetTrustedProxies([]string{"localhost", "127.0.0.1"})

	router.GET(NewWalletEndpoint, apiNewWallet)

	return router
}
