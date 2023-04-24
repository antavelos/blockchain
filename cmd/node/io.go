package main

import (
	"fmt"
	"sync"

	"github.com/antavelos/blockchain/pkg/common"
	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
	"github.com/antavelos/blockchain/pkg/models/wallet"
)

func ioAddTx(tx bc.Transaction) (bc.Transaction, error) {
	err := tx.Validate()
	if err != nil {
		return bc.Transaction{}, err
	}

	bdb := getBlockchainDb()
	m := sync.Mutex{}

	m.Lock()
	defer m.Unlock()

	blockchain, err := bdb.LoadBlockchain()
	if err != nil {
		return tx, common.GenericError{Msg: "blockchain currently not available"}
	}

	tx, err = blockchain.AddTx(tx)
	if err != nil {
		return tx, err
	}

	if err := bdb.SaveBlockchain(*blockchain); err != nil {
		return tx, common.GenericError{Msg: "couldn't update blockchain"}
	}

	return tx, nil
}

func ioAddBlock(block bc.Block) (bc.Block, error) {
	difficulty := getMiningDifficulty()

	if !block.IsValid(difficulty) {
		return bc.Block{}, common.GenericError{Msg: fmt.Sprintf("block does not start with %v '0'", difficulty)}
	}

	bdb := getBlockchainDb()
	m := sync.Mutex{}

	m.Lock()
	defer m.Unlock()

	blockchain, err := bdb.LoadBlockchain()
	if err != nil {
		return block, common.GenericError{Msg: "blockchain currently not available"}
	}

	if err := blockchain.AddBlock(block); err != nil {
		return block, err
	}

	if err := bdb.SaveBlockchain(*blockchain); err != nil {
		return block, common.GenericError{Msg: "couldn't update blockchain"}
	}

	return block, nil
}

func ioNewBlockchain() {
	dbd := getBlockchainDb()

	blockchain := bc.NewBlockchain()

	dbd.SaveBlockchain(*blockchain)
}

func ioGetWallet() (wallet.Wallet, error) {
	wdb := getWalletDb()
	wallets, err := wdb.LoadWallets()
	if err != nil {
		return wallet.Wallet{}, err
	}

	if len(wallets) == 0 {
		return wallet.Wallet{}, common.GenericError{Msg: "no wallets available"}
	}

	return wallets[0], nil
}
