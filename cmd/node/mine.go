package main

import (
	"fmt"
	"strings"

	node_client "github.com/antavelos/blockchain/pkg/clients/node"
	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/lib/bus"
	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
)

func shareBlock(block bc.Block) error {
	ndb := getNodeDb()

	nodes, err := ndb.LoadNodes()
	if err != nil {
		return common.GenericError{Msg: "failed to share new block"}
	}

	responses := node_client.ShareBlock(nodes, block)

	if responses.HasConnectionRefused() {
		common.LogInfo("Refresing DNS nodes")
		bus.Publish(RefreshDnsNodes, nil)
	}

	if responses.ErrorsRatio() > 0 {
		msg := fmt.Sprintf("new block was not accepted by some nodes: \n%v", strings.Join(responses.ErrorStrings(), "\n"))
		common.LogError(msg)

		if responses.ErrorsRatio() > 0.5 {
			return common.GenericError{Msg: msg}
		}
	}

	return nil
}

func mine(block *bc.Block, difficulty int) {

	common.LogInfo("Mining...")
	for !block.IsValid(difficulty) {
		block.Nonce += 1
	}
	common.LogInfo("New block mined with nonce", block.Nonce)
}

func Mine() (bc.Block, error) {
	bdb := getBlockchainDb()

	blockchain, err := bdb.LoadBlockchain()

	if err != nil {
		return bc.Block{}, common.GenericError{Msg: "blockchain currently not available"}
	}

	if !blockchain.HasPendingTxs() {
		return bc.Block{}, common.GenericError{Msg: "no pending transactions found"}
	}

	txsPerBlock := getTxsNumPerBlock()
	block, err := blockchain.NewBlock(txsPerBlock)
	if err != nil {
		return bc.Block{}, err
	}

	difficulty := getMiningDifficulty()

	mine(&block, difficulty)

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
