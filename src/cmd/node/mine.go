package main

import (
	"errors"
	"strings"
	"sync"

	bc "github.com/antavelos/blockchain/src/blockchain"
)

func Mine() (bc.Block, error) {
	m := sync.Mutex{}

	m.Lock()
	defer m.Unlock()

	blockchain, err := ioLoadBlockchain()

	if err != nil {
		return bc.Block{}, errors.New("blockchain currently not available")
	}

	block, err := blockchain.NewBlock()
	if err != nil {
		return bc.Block{}, err
	}

	err = blockchain.AddBlock(block)
	if err != nil {
		return bc.Block{}, err
	}

	err = ioSaveBlockchain(*blockchain)
	if err != nil {
		return bc.Block{}, errors.New("failed to update blockchain")
	}

	if nodeErrors := ShareBlock(block); nodeErrors != nil {
		errorStrings := ErrorsToStrings(nodeErrors)
		if len(errorStrings) > 0 {
			ErrorLogger.Printf("Failed to share the block with other nodes: \n%v", strings.Join(errorStrings, "\n"))
		}
	}

	return block, nil
}
