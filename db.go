package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const filename = "./blockchain.json"

func dbSaveBlockchain(blockchain Blockchain) error {
	jsonBlockchain, err := json.Marshal(blockchain)
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
