package main

import (
	// "encoding/json"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	bc "github.com/antavelos/blockchain/src/blockchain"
	cn "github.com/antavelos/blockchain/src/common"

	"github.com/gin-gonic/gin"
)

func getBlockchain() (*bc.Blockchain, error) {
	nodes, err := retrieveDnsNodes()
	if err != nil {
		return &bc.Blockchain{}, errors.New("nodes not available")
	}

	nodeBlockchains := getNodeBlockchains(nodes)

	return getMaxLengthBlockchain(nodeBlockchains), nil
}

func index(c *gin.Context) {
	blockchain, err := getBlockchain()
	if err != nil {
		cn.ErrorLogger.Println("Couldn't retrieve blockchain")

		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":      "Blockchain",
			"blockchain": "blockchain not available",
		})

		return
	}

	blockchainBytes, _ := json.MarshalIndent(blockchain, "", "  ")

	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":      "Blockchain",
		"blockchain": string(blockchainBytes),
	})
}

func getDnsHost() string {
	return fmt.Sprintf("http://%v:%v", os.Getenv("DNS_HOST"), os.Getenv("DNS_PORT"))
}

func getNodeBlockchain(node bc.Node) (*bc.Blockchain, error) {
	url := node.Host + "/blockchain"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var blockchain bc.Blockchain
	if err := json.Unmarshal(body, &blockchain); err != nil {
		return nil, err
	}

	return &blockchain, nil
}

func getNodeBlockchains(nodes []bc.Node) []*bc.Blockchain {
	return cn.Map(nodes, func(node bc.Node) *bc.Blockchain {
		b, err := getNodeBlockchain(node)
		if err != nil {
			return &bc.Blockchain{}
		}

		return b
	})
}

func getMaxLengthBlockchain(blockchains []*bc.Blockchain) *bc.Blockchain {
	if len(blockchains) == 0 {
		return nil
	}

	maxLengthBlockchain := blockchains[0]

	for _, blockchain := range blockchains[1:] {
		if len(blockchain.Blocks) > len(maxLengthBlockchain.Blocks) {
			maxLengthBlockchain = blockchain
		}
	}

	return maxLengthBlockchain
}

func retrieveDnsNodes() ([]bc.Node, error) {
	var nodes []bc.Node
	url := getDnsHost() + "/nodes"

	resp, err := http.Get(url)
	if err != nil {
		return nodes, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nodes, err
	}

	if err := json.Unmarshal(body, &nodes); err != nil {
		return nodes, err
	}

	return nodes, nil
}

func initRouter() *gin.Engine {
	router := gin.Default()

	router.SetTrustedProxies([]string{"localhost", "127.0.0.1"})

	router.LoadHTMLGlob("cmd/admin/templates/*")

	router.GET("/", index)

	return router
}
