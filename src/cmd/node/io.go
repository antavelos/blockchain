package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	bc "github.com/antavelos/blockchain"
)

// const blockchainFilename = "./blockchain.json"
// const nodesFilename = "./nodes.json"

var jsonBlockchain []byte
var jsonNodes []byte

func getBlockchainFilename() string {
	return "./blockchain_" + *Port + ".json"
}

func getNodesFilename() string {
	return "./nodes_" + *Port + ".json"
}

func ioSaveBlockchain(blockchain bc.Blockchain) error {
	marshalled, err := json.MarshalIndent(blockchain, "", "  ")
	if err != nil {
		return err
	}

	jsonBlockchain = marshalled

	// return nil
	return ioutil.WriteFile(getBlockchainFilename(), jsonBlockchain, os.ModePerm)
}

func ioLoadBlockchain() (*bc.Blockchain, error) {
	var blockchain bc.Blockchain

	file, err := ioutil.ReadFile(getBlockchainFilename())
	if err != nil {
		return nil, err
	}

	json.Unmarshal(file, &blockchain)
	// json.Unmarshal(jsonBlockchain, &blockchain)

	return &blockchain, nil
}

func ioBlockchainExists() bool {
	_, err := os.Stat(getBlockchainFilename())

	return !errors.Is(err, os.ErrNotExist)
	// return false
}

func ioSaveNodes(nodes []bc.Node) error {
	marshalled, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}

	jsonNodes = marshalled
	// return nil
	return ioutil.WriteFile(getNodesFilename(), jsonNodes, os.ModePerm)

}

func ioLoadNodes() ([]bc.Node, error) {
	var nodes []bc.Node

	file, err := ioutil.ReadFile(getNodesFilename())
	if err != nil {
		return nil, err
	}

	json.Unmarshal(file, &nodes)
	// json.Unmarshal(jsonNodes, &nodes)

	return nodes, nil
}

func ioNewBlockchain() {
	blockchain := bc.NewBlockchain()
	ioSaveBlockchain(*blockchain)
}

func ioAddNode(node bc.Node) error {
	nodes, err := ioLoadNodes()
	if err != nil {
		return errors.New("nodes list not available")
	}

	if !containsNode(nodes, node) {
		nodes = append(nodes, node)
	}

	err = ioSaveNodes(nodes)
	if err != nil {
		return errors.New("couldn't update nodes' list")
	}

	return nil
}

func containsNode(nodes []bc.Node, node bc.Node) bool {
	for _, n := range nodes {
		if n.Host == node.Host {
			return true
		}
	}
	return false
}
