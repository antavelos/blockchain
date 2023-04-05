package main

import (
	"errors"
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

	blockchain.AddBlock(block)
	err = ioSaveBlockchain(*blockchain)
	if err != nil {
		return bc.Block{}, errors.New("failed to update blockchain")
	}

	// TODO: broadcast the block to the other nodes

	return block, nil
}
