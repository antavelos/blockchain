package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

const filename = "./blockchain.json"

func dbSaveBlockchain(blockchain Blockchain) error {
	jsonBlockchain, err := json.MarshalIndent(blockchain, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, jsonBlockchain, os.ModePerm)
}

func dbLoadBlockchain() (*Blockchain, error) {
	var blockchain Blockchain

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(file, &blockchain)

	return &blockchain, nil
}

func dbBlockchainExists() bool {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
