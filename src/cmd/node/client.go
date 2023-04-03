package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	bc "github.com/antavelos/blockchain/src/blockchain"
)

func ping(node bc.Node) error {
	host := node.Host + "/ping"
	selfNode := bc.Node{Host: "http://localhost:" + *Port}
	jsonSelfNode, err := json.Marshal(selfNode)
	if err != nil {
		return err
	}

	resp, err := http.Post(host, "application/json", bytes.NewBuffer(jsonSelfNode))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var nodes []bc.Node
	if err := json.Unmarshal(body, &nodes); err != nil {
		return err
	}

	for _, node := range nodes {
		if node.Host != selfNode.Host {
			ioAddNode(node)
		}
	}

	return nil
}

func getBlockchain(node bc.Node) (*bc.Blockchain, error) {
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

func pingDns() ([]bc.Node, error) {
	var nodes []bc.Node
	url := dnsHost + "/nodes"

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

func resolveLongestBlockchain(nodes []bc.Node) {
	maxLengthBlockchain := getMaxLengthBlockchain(nodes)

	if len(maxLengthBlockchain.Blocks) == 0 {
		return
	}
	blockchain, _ := ioLoadBlockchain()

	blockchain.Blocks = maxLengthBlockchain.Blocks

	ioSaveBlockchain(*blockchain)
}

func getMaxLengthBlockchain(nodes []bc.Node) *bc.Blockchain {
	blockchain, _ := ioLoadBlockchain()

	maxLengthBlockchain := blockchain

	for _, node := range nodes {
		nodeBlockchain, err := getBlockchain(node)
		if err != nil {
			ErrorLogger.Printf("Couldn't retrieve blockchain from node %v: %v", node.Host, err.Error())
			continue
		}

		if len(nodeBlockchain.Blocks) > len(maxLengthBlockchain.Blocks) {
			maxLengthBlockchain = nodeBlockchain
		}
	}
	return maxLengthBlockchain
}
