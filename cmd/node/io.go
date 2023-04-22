package main

import (
	"sync"

	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/db"
	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
	"github.com/antavelos/blockchain/pkg/models/wallet"
)

func ioAddTx(tx bc.Transaction) (bc.Transaction, error) {
	bdb := db.GetBlockchainDb()
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
	bdb := db.GetBlockchainDb()
	m := sync.Mutex{}

	m.Lock()
	defer m.Unlock()

	blockchain, err := bdb.LoadBlockchain()
	if err != nil {
		return block, common.GenericError{Msg: "blockchain currently not available"}
	}

	err = blockchain.AddBlock(block)
	if err != nil {
		return block, err
	}

	if err := bdb.SaveBlockchain(*blockchain); err != nil {
		return block, common.GenericError{Msg: "couldn't update blockchain"}
	}

	return block, nil
}

func ioNewBlockchain() {
	dbd := db.GetBlockchainDb()

	blockchain := bc.NewBlockchain()

	dbd.SaveBlockchain(*blockchain)
}

func ioGetWallet() (wallet.Wallet, error) {
	wdb := db.GetWalletDb()
	wallets, err := wdb.LoadWallets()
	if err != nil {
		return wallet.Wallet{}, err
	}

	if len(wallets) == 0 {
		return wallet.Wallet{}, common.GenericError{Msg: "no wallets available"}
	}

	return wallets[0], nil
}
