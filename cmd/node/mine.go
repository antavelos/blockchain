package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	bc "github.com/antavelos/blockchain/pkg/blockchain"
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
	if len(nodeErrors) > 0 {
		errorStrings := ErrorsToStrings(nodeErrors)
		return bc.Block{}, fmt.Errorf("failed to share the block with other nodes: \n%v", strings.Join(errorStrings, "\n"))
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
