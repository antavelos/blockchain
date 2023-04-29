package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	dns_client "github.com/antavelos/blockchain/src/internal/pkg/clients/dns"
	node_client "github.com/antavelos/blockchain/src/internal/pkg/clients/node"
	bc "github.com/antavelos/blockchain/src/internal/pkg/models/blockchain"
	nd "github.com/antavelos/blockchain/src/internal/pkg/models/node"
	"github.com/antavelos/blockchain/src/pkg/rest"
	"github.com/antavelos/blockchain/src/pkg/utils"

	"github.com/gin-gonic/gin"
)

func getDNSHost() string {
	return fmt.Sprintf("http://%v:%v", config["DNS_HOST"], config["DNS_PORT"])
}

func getBlockchain() (*bc.Blockchain, error) {
	dnsHost := getDNSHost()
	nodes, err := dns_client.GetDNSNodes(dnsHost)
	if err != nil {
		return &bc.Blockchain{}, utils.GenericError{Msg: "nodes not available"}
	}

	nodeBlockchains := getNodeBlockchains(nodes)

	return bc.GetMaxLengthBlockchain(nodeBlockchains), nil
}

func getNodeBlockchains(nodes []nd.Node) []*bc.Blockchain {
	responses := node_client.GetBlockchains(nodes)

	return utils.Map(responses, func(response rest.Response) *bc.Blockchain {
		if response.Err != nil {
			return &bc.Blockchain{}
		}

		blockchain, err := bc.UnmarshalBlockchain(response.Body)
		if err != nil {
			return &bc.Blockchain{}
		}

		return &blockchain
	})
}

func index(c *gin.Context) {
	data := gin.H{
		"title": "Blockchain",
	}
	blockchain, err := getBlockchain()
	if err != nil {
		utils.LogError("Couldn't retrieve blockchain")
		data["blockchain"] = "blockchain not available"
		c.HTML(http.StatusOK, "index.html", data)
		return
	}

	blockchainBytes, _ := json.MarshalIndent(blockchain, "", "  ")

	data["blockchain"] = string(blockchainBytes)
	c.HTML(http.StatusOK, "index.html", data)
}

func initRouter() *gin.Engine {
	router := gin.Default()

	router.SetTrustedProxies([]string{"localhost", "127.0.0.1"})

	router.LoadHTMLGlob("cmd/admin/templates/*")

	router.GET("/", index)

	return router
}
