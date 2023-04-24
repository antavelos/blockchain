package main

import (
	"net/http"

	"github.com/antavelos/blockchain/pkg/common"
	"github.com/gin-gonic/gin"
)

const NewWalletEndpoint = "/wallets/new"

func apiNewWallet(c *gin.Context) {
	wdb := getWalletDb()

	wallet, err := wdb.CreateWallet()
	if err != nil {
		common.LogError("New wallet [FAIL]", err.Error())
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
