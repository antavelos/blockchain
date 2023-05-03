package walletapi

import (
	"net/http"

	"github.com/antavelos/blockchain/src/internal/pkg/repos"
	"github.com/antavelos/blockchain/src/pkg/utils"
	"github.com/gin-gonic/gin"
)

const NewWalletEndpoint = "/wallets/new"

type RouteHandler struct {
	WalletRepo *repos.WalletRepo
}

func NewRouteHandler(walletRepo *repos.WalletRepo) *RouteHandler {
	return &RouteHandler{WalletRepo: walletRepo}
}

func (h RouteHandler) apiNewWallet(c *gin.Context) {
	wallet, err := h.WalletRepo.CreateWallet()
	if err != nil {
		utils.LogError("New wallet [FAIL]", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to create new wallet"})
		return
	}

	c.IndentedJSON(http.StatusCreated, wallet)
}

func (h *RouteHandler) InitRouter() *gin.Engine {
	router := gin.Default()

	router.SetTrustedProxies([]string{"localhost", "127.0.0.1"})

	router.GET(NewWalletEndpoint, h.apiNewWallet)

	return router
}
