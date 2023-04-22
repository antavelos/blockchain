package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	dns_client "github.com/antavelos/blockchain/pkg/clients/dns"
	node_client "github.com/antavelos/blockchain/pkg/clients/node"
	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/lib/rest"
	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
	nd "github.com/antavelos/blockchain/pkg/models/node"

	"github.com/gin-gonic/gin"
)

func getDnsHost() string {
	return fmt.Sprintf("http://%v:%v", os.Getenv("DNS_HOST"), os.Getenv("DNS_PORT"))
}

func getBlockchain() (*bc.Blockchain, error) {
	dnsHost := getDnsHost()
	nodes, err := dns_client.GetDnsNodes(dnsHost)
	if err != nil {
		return &bc.Blockchain{}, common.GenericError{Msg: "nodes not available"}
	}

	nodeBlockchains := getNodeBlockchains(nodes)

	return bc.GetMaxLengthBlockchain(nodeBlockchains), nil
}

func getNodeBlockchains(nodes []nd.Node) []*bc.Blockchain {
	responses := node_client.GetBlockchains(nodes)

	return common.Map(responses, func(response rest.Response) *bc.Blockchain {
		if response.Err != nil {
			return &bc.Blockchain{}
		}

		return response.Body.(*bc.Blockchain)
	})
}

func index(c *gin.Context) {
	data := gin.H{
		"title": "Blockchain",
	}
	blockchain, err := getBlockchain()
	if err != nil {
		common.LogError("Couldn't retrieve blockchain")
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
