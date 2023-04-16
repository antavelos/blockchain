package main

import (
	"errors"
	"fmt"
	"strings"

	node_client "github.com/antavelos/blockchain/pkg/clients/node"
	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/db"
	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
)

func shareBlock(block bc.Block) error {
	ndb := db.GetNodeDb()

	nodes, err := ndb.LoadNodes()
	if err != nil {
		return fmt.Errorf("failed to share new block: %v", err.Error())
	}

	responses := node_client.ShareBlock(nodes, block)
	if responses.ErrorsRatio() > 0 {
		errorStrings := strings.Join(responses.ErrorStrings(), "\n")
		common.ErrorLogger.Printf("new block was not accepted by some nodes: \n%v", errorStrings)

		if responses.ErrorsRatio() > 0.5 {
			return fmt.Errorf("new block was rejected by other nodes: \n%v", errorStrings)
		}
	}

	return nil
}

func Mine() (bc.Block, error) {
	bdb := db.GetBlockchainDb()

	blockchain, err := bdb.LoadBlockchain()

	if err != nil {
		return bc.Block{}, errors.New("blockchain currently not available")
	}

	block, err := blockchain.NewBlock()
	if err != nil {
		return bc.Block{}, err
	}

	err = shareBlock(block)
	if err != nil {
		return bc.Block{}, err
	}

	err = blockchain.AddBlock(block)
	if err != nil {
		return bc.Block{}, err
	}

	err = bdb.SaveBlockchain(*blockchain)
	if err != nil {
		return bc.Block{}, errors.New("failed to update blockchain")
	}

	return block, nil
}
