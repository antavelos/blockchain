package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
	nd "github.com/antavelos/blockchain/pkg/models/node"
	w "github.com/antavelos/blockchain/pkg/models/wallet"
)

type DB struct {
	Filename string
}

type BlockchainDB DB
type NodeDB DB
type WalletDB DB

func GetBlockchainDb() *BlockchainDB {
	return &BlockchainDB{Filename: os.Getenv("BLOCKCHAIN_FILENAME")}
}

func GetNodeDb() *NodeDB {
	return &NodeDB{Filename: os.Getenv("NODES_FILENAME")}
}

func GetWalletDb() *WalletDB {
	return &WalletDB{Filename: os.Getenv("WALLETS_FILENAME")}
}

func createIfNotExists(filename string) error {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		file, err := os.OpenFile(filename, os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		file.Close()
	}

	return nil
}

func read(filename string) ([]byte, error) {
	err := createIfNotExists(filename)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(filename)
}

func write(filename string, data []byte) error {
	err := createIfNotExists(filename)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, os.ModePerm)
}

func (db *BlockchainDB) SaveBlockchain(blockchain bc.Blockchain) error {
	blockchainBytes, err := json.MarshalIndent(blockchain, "", "  ")
	if err != nil {
		return err
	}

	return write(db.Filename, blockchainBytes)
}

func updateBlockchain(oldBlockchain *bc.Blockchain, newBlockchain *bc.Blockchain) *bc.Blockchain {
	if oldBlockchain == nil {
		return newBlockchain
	}

	// TODO: append the blocks diff
	oldBlockchain.Blocks = newBlockchain.Blocks

	// TODO: to refactor
	for i := len(oldBlockchain.Blocks) - 1; i > 0; i-- {
		for _, tx := range oldBlockchain.TxPool {
			if oldBlockchain.Blocks[i].HasTx(tx) {
				oldBlockchain.RemoveTx(tx)
			}
		}
	}

	return oldBlockchain
}

func (db *BlockchainDB) UpdateBlockchain(newBlockchain *bc.Blockchain) error {

	m := sync.Mutex{}

	m.Lock()
	defer m.Unlock()

	blockchain, err := db.LoadBlockchain()
	if err != nil {
		return err
	}

	blockchain = updateBlockchain(blockchain, newBlockchain)

	return db.SaveBlockchain(*blockchain)
}

func (db *BlockchainDB) LoadBlockchain() (*bc.Blockchain, error) {
	var blockchain bc.Blockchain

	file, err := read(db.Filename)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(file, &blockchain)

	return &blockchain, nil
}

func (db *NodeDB) SaveNodes(nodes []nd.Node) error {
	nodesBytes, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}

	return write(db.Filename, nodesBytes)
}

func (db *NodeDB) LoadNodes() ([]nd.Node, error) {
	var nodes []nd.Node

	file, err := read(db.Filename)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(file, &nodes)

	return nodes, nil
}

func (db *NodeDB) AddNode(node nd.Node) error {
	nodes, err := db.LoadNodes()
	if err != nil {
		return errors.New("nodes not available")
	}

	if !containsNode(nodes, node) {
		nodes = append(nodes, node)
	}

	err = db.SaveNodes(nodes)
	if err != nil {
		return errors.New("couldn't update nodes")
	}

	return nil
}

func containsNode(nodes []nd.Node, node nd.Node) bool {
	for _, n := range nodes {
		if n.GetHost() == node.GetHost() {
			return true
		}
	}
	return false
}

func (db *WalletDB) SaveWallet(wallet w.Wallet) error {
	wallets, err := db.LoadWallets()
	if err != nil {
		return err
	}

	wallets = append(wallets, wallet)

	marshalled, err := json.MarshalIndent(wallets, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal wallets: %v", err.Error())
	}

	return write(db.Filename, marshalled)
}

func (db *WalletDB) LoadWallets() ([]w.Wallet, error) {
	var wallets []w.Wallet

	file, err := read(db.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load wallets from file: %v", err.Error())
	}

	json.Unmarshal(file, &wallets)

	return wallets, nil
}

func (db *WalletDB) CreateWallet() (*w.Wallet, error) {

	wallet, err := w.NewWallet()
	if err != nil {
		return nil, fmt.Errorf("failed to create a new wallet: %v", err.Error())
	}

	err = db.SaveWallet(*wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to save new wallet: %v", err.Error())
	}

	return wallet, nil
}
