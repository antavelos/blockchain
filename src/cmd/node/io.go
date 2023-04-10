package main

import (
	"errors"
	"os"
	"sync"

	bc "github.com/antavelos/blockchain/src/blockchain"
	"github.com/antavelos/blockchain/src/db"
)

func getBlockchainDb() *db.BlockchainDB {
	return &db.BlockchainDB{Filename: os.Getenv("BLOCKCHAIN_FILENAME")}
}

func getNodeDb() *db.NodeDB {
	return &db.NodeDB{Filename: os.Getenv("NODES_FILENAME")}
}

func ioAddTx(tx bc.Transaction) (bc.Transaction, error) {
	bdb := getBlockchainDb()
	m := sync.Mutex{}

	m.Lock()
	defer m.Unlock()

	blockchain, err := bdb.LoadBlockchain()
	if err != nil {
		return tx, errors.New("blockchain currently not available")
	}

	tx, err = blockchain.AddTx(tx)
	if err != nil {
		return tx, err
	}

	if err := bdb.SaveBlockchain(*blockchain); err != nil {
		return tx, errors.New("couldn't update blockchain")
	}

	return tx, nil
}

func ioAddBlock(block bc.Block) (bc.Block, error) {
	bdb := getBlockchainDb()
	m := sync.Mutex{}

	m.Lock()
	defer m.Unlock()

	blockchain, err := bdb.LoadBlockchain()
	if err != nil {
		return block, errors.New("blockchain currently not available")
	}

	err = blockchain.AddBlock(block)
	if err != nil {
		return block, err
	}

	if err := bdb.SaveBlockchain(*blockchain); err != nil {
		return block, errors.New("couldn't update blockchain")
	}

	return block, nil
}

func ioNewBlockchain() {
	dbd := getBlockchainDb()

	blockchain := bc.NewBlockchain()

	dbd.SaveBlockchain(*blockchain)
}
