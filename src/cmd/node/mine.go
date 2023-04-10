package main

import (
	"errors"
	"strings"
	"sync"

	bc "github.com/antavelos/blockchain/src/blockchain"
)

func Mine() (bc.Block, error) {
	bdb := getBlockchainDb()
	ndb := getNodeDb()

	m := sync.Mutex{}

	m.Lock()
	defer m.Unlock()

	blockchain, err := bdb.LoadBlockchain()

	if err != nil {
		return bc.Block{}, errors.New("blockchain currently not available")
	}

	block, err := blockchain.NewBlock()
	if err != nil {
		return bc.Block{}, err
	}

	nodeErrors := ShareBlock(block)
	if nodeErrors != nil {
		errorStrings := ErrorsToStrings(nodeErrors)
		if len(errorStrings) > 0 {
			ErrorLogger.Printf("Failed to share the block with other nodes: \n%v", strings.Join(errorStrings, "\n"))
		}
	}

	nodes, _ := ndb.LoadNodes()
	if float64(len(nodeErrors)) >= (float64(len(nodes)) / 2.0) {
		return bc.Block{}, errors.New("new block was rejected by other nodes")
	}

	err = blockchain.AddBlock(block)
	if err != nil {
		return bc.Block{}, err
	}

	err = bdb.SaveBlockchain(*blockchain)
	if err != nil {
		return bc.Block{}, errors.New("failed to update blockchain")
	}

	return block, nil
}
