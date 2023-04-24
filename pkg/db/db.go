package db

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/antavelos/blockchain/pkg/common"
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

func GetBlockchainDb(filename string) *BlockchainDB {
	return &BlockchainDB{Filename: filename}
}

func GetNodeDb(filename string) *NodeDB {
	return &NodeDB{Filename: filename}
}

func GetWalletDb(filename string) *WalletDB {
	return &WalletDB{Filename: filename}
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

func (db *BlockchainDB) UpdateBlockchain(newBlockchain *bc.Blockchain) error {

	m := sync.Mutex{}

	m.Lock()
	defer m.Unlock()

	blockchain, err := db.LoadBlockchain()
	if err != nil {
		return err
	}

	blockchain = bc.UpdateBlockchain(blockchain, newBlockchain)

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
		return common.GenericError{Msg: "nodes not available"}
	}

	if !containsNode(nodes, node) {
		nodes = append(nodes, node)
	}

	err = db.SaveNodes(nodes)
	if err != nil {
		return common.GenericError{Msg: "couldn't update nodes"}
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
		return common.GenericError{Msg: "failed to marshal wallets"}
	}

	return write(db.Filename, marshalled)
}

func (db *WalletDB) LoadWallets() ([]w.Wallet, error) {
	var wallets []w.Wallet

	file, err := read(db.Filename)
	if err != nil {
		return nil, common.GenericError{Msg: "failed to load wallets from file"}
	}

	json.Unmarshal(file, &wallets)

	return wallets, nil
}

func (db *WalletDB) CreateWallet() (*w.Wallet, error) {

	wallet, err := w.NewWallet()
	if err != nil {
		return nil, common.GenericError{Msg: "failed to create a new wallet", Extra: err}
	}

	err = db.SaveWallet(*wallet)
	if err != nil {
		return nil, common.GenericError{Msg: "failed to save new wallet"}
	}

	return wallet, nil
}

func (db *WalletDB) IsEmpty() bool {

	wallets, _ := db.LoadWallets()

	return len(wallets) == 0
}
