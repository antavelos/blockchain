package main

import (
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
		return common.GenericError{Msg: "failed to share new block"}
	}

	responses := node_client.ShareBlock(nodes, block)
	if responses.ErrorsRatio() > 0 {
		msg := fmt.Sprintf("new block was not accepted by some nodes: \n%v", strings.Join(responses.ErrorStrings(), "\n"))
		common.LogError(msg)

		if responses.ErrorsRatio() > 0.5 {
			return common.GenericError{Msg: msg}
		}
	}

	return nil
}

func Mine() (bc.Block, error) {
	bdb := db.GetBlockchainDb()

	blockchain, err := bdb.LoadBlockchain()

	if err != nil {
		return bc.Block{}, common.GenericError{Msg: "blockchain currently not available"}
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
		return bc.Block{}, common.GenericError{Msg: "failed to update blockchain"}
	}

	return block, nil
}
