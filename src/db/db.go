package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	bc "github.com/antavelos/blockchain/src/blockchain"
	"github.com/antavelos/blockchain/src/wallet"
)

type DB struct {
	Filename string
}

type BlockchainDB DB
type NodeDB DB
type WalletDB DB

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

func (db *BlockchainDB) LoadBlockchain() (*bc.Blockchain, error) {
	var blockchain bc.Blockchain

	file, err := read(db.Filename)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(file, &blockchain)

	return &blockchain, nil
}

func (db *NodeDB) SaveNodes(nodes []bc.Node) error {
	nodesBytes, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}

	return write(db.Filename, nodesBytes)
}

func (db *NodeDB) LoadNodes() ([]bc.Node, error) {
	var nodes []bc.Node

	file, err := read(db.Filename)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(file, &nodes)

	return nodes, nil
}

func (db *NodeDB) AddNode(node bc.Node) error {
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

func containsNode(nodes []bc.Node, node bc.Node) bool {
	for _, n := range nodes {
		if n.Host == node.Host {
			return true
		}
	}
	return false
}

func (db *WalletDB) SaveWallet(w wallet.Wallet) error {
	wallets, err := db.LoadWallets()
	if err != nil {
		return err
	}

	wallets = append(wallets, w)

	marshalled, err := json.MarshalIndent(wallets, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal wallets: %v", err.Error())
	}

	return write(db.Filename, marshalled)
}

func (db *WalletDB) LoadWallets() ([]wallet.Wallet, error) {
	var wallets []wallet.Wallet

	file, err := read(db.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load wallets from file: %v", err.Error())
	}

	json.Unmarshal(file, &wallets)

	return wallets, nil
}

func (db *WalletDB) CreateWallet() (*wallet.Wallet, error) {

	w, err := wallet.NewWallet()
	if err != nil {
		return &wallet.Wallet{}, fmt.Errorf("failed to create a new wallet: %v", err.Error())
	}

	err = db.SaveWallet(*w)
	if err != nil {
		return &wallet.Wallet{}, fmt.Errorf("failed to save new wallet: %v", err.Error())
	}

	return w, nil
}
