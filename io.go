package blockchain

import (
	"encoding/json"
	"errors"
)

// const blockchainFilename = "./blockchain.json"
// const nodesFilename = "./nodes.json"

var jsonBlockchain []byte
var jsonNodes []byte

func DbSaveBlockchain(blockchain Blockchain) error {
	marshalled, err := json.MarshalIndent(blockchain, "", "  ")
	if err != nil {
		return err
	}

	jsonBlockchain = marshalled

	return nil
	// return ioutil.WriteFile(blockchainFilename, jsonBlockchain, os.ModePerm)
}

func DbLoadBlockchain() (*Blockchain, error) {
	var blockchain Blockchain

	// file, err := ioutil.ReadFile(blockchainFilename)
	// if err != nil {
	// 	return nil, err
	// }

	// json.Unmarshal(file, &blockchain)
	json.Unmarshal(jsonBlockchain, &blockchain)

	return &blockchain, nil
}

func DbBlockchainExists() bool {
	// _, err := os.Stat(blockchainFilename)

	// return !errors.Is(err, os.ErrNotExist)
	return false
}

func DbSaveNodes(nodes []Node) error {
	marshalled, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return err
	}

	jsonNodes = marshalled
	return nil
	// return ioutil.WriteFile(nodesFilename, jsonNodes, os.ModePerm)

}

func DbLoadNodes() ([]Node, error) {
	var nodes []Node

	// file, err := ioutil.ReadFile(nodesFilename)
	// if err != nil {
	// 	return nil, err
	// }

	// json.Unmarshal(file, &nodes)
	json.Unmarshal(jsonNodes, &nodes)

	return nodes, nil
}

func DbAddNode(node Node) error {
	nodes, err := DbLoadNodes()
	if err != nil {
		return errors.New("nodes list not available")
	}

	found := false
	for _, n := range nodes {
		if n.Host == node.Host {
			found = true
			break
		}
	}

	if !found {
		nodes = append(nodes, node)
	}

	err = DbSaveNodes(nodes)
	if err != nil {
		return errors.New("couldn't update nodes' list")
	}

	return nil
}
