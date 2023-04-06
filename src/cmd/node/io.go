package main

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	bc "github.com/antavelos/blockchain/src/blockchain"
)

func getBlockchainFilename() string {
	return "./blockchain_" + *Port + ".json"
}

func getNodesFilename() string {
	return "./nodes_" + *Port + ".json"
}

func ioSaveBlockchain(blockchain bc.Blockchain) error {
	jsonBlockchain, err := json.MarshalIndent(blockchain, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getBlockchainFilename(), jsonBlockchain, os.ModePerm)
}

func ioLoadBlockchain() (*bc.Blockchain, error) {
	var blockchain bc.Blockchain

	file, err := os.ReadFile(getBlockchainFilename())
	if err != nil {
		return nil, err
	}

	json.Unmarshal(file, &blockchain)

	return &blockchain, nil
}

func ioAddTx(tx bc.Transaction) (bc.Transaction, error) {
	m := sync.Mutex{}

	m.Lock()
	defer m.Unlock()

	blockchain, err := ioLoadBlockchain()
	if err != nil {
		return tx, errors.New("blockchain currently not available")
	}

	tx, err = blockchain.AddTx(tx)
	if err != nil {
		return tx, err
	}

	if err := ioSaveBlockchain(*blockchain); err != nil {
		return tx, errors.New("couldn't update blockchain")
	}

	return tx, nil
}

func ioAddBlock(block bc.Block) (bc.Block, error) {
	m := sync.Mutex{}

	m.Lock()
	defer m.Unlock()

	blockchain, err := ioLoadBlockchain()
	if err != nil {
		return block, errors.New("blockchain currently not available")
	}

	err = blockchain.AddBlock(block)
	if err != nil {
		return block, err
	}

	if err := ioSaveBlockchain(*blockchain); err != nil {
		return block, errors.New("couldn't update blockchain")
	}

	return block, nil
}

func ioSaveNodes(nodes []bc.Node) error {
	jsonNodes, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(getNodesFilename(), jsonNodes, os.ModePerm)
}

func ioLoadNodes() ([]bc.Node, error) {
	var nodes []bc.Node

	file, err := os.ReadFile(getNodesFilename())
	if err != nil {
		return nil, err
	}

	json.Unmarshal(file, &nodes)

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
