package main

import (
	"fmt"

	node_client "github.com/antavelos/blockchain/src/internal/pkg/clients/node"
	bc "github.com/antavelos/blockchain/src/internal/pkg/models/blockchain"
	"github.com/antavelos/blockchain/src/pkg/bus"
	"github.com/antavelos/blockchain/src/pkg/utils"
)

func shareBlock(block bc.Block) error {
	nrepo := getNodeRepo()

	nodes, err := nrepo.GetNodes()
	if err != nil {
		return utils.GenericError{Msg: "failed to share new block"}
	}

	responses := node_client.ShareBlock(nodes, block)

	if responses.HasConnectionRefused() {
		utils.LogInfo("Refresing DNS nodes")
		bus.Publish(RefreshDNSNodesTopic, nil)
	}

	if responses.ErrorsRatio() > 0 {
		msg := fmt.Sprintf("new block was not accepted by some nodes: %v", responses.Errors())
		utils.LogError(msg)

		if responses.ErrorsRatio() > 0.49 {
			return utils.GenericError{Msg: msg}
		}
	}

	return nil
}

func mine(block *bc.Block, difficulty int) {

	utils.LogInfo("Mining...")
	for !block.IsValid(difficulty) {
		block.Nonce += 1
	}
	utils.LogInfo("New block mined with nonce", block.Nonce)
}

func Mine() (bc.Block, error) {
	brepo := getBlockchainRepo()

	blockchain, err := brepo.GetBlockchain()

	if err != nil {
		return bc.Block{}, utils.GenericError{Msg: "blockchain currently not available"}
	}

	if !blockchain.HasPendingTxs() {
		return bc.Block{}, utils.GenericError{Msg: "no pending transactions found"}
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

	err = brepo.ReplaceBlockchain(*blockchain)
	if err != nil {
		return bc.Block{}, utils.GenericError{Msg: "failed to update blockchain"}
	}

	return block, nil
}
