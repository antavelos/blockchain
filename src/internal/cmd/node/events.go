package main

import (
	"fmt"

	node_client "github.com/antavelos/blockchain/src/internal/pkg/clients/node"
	bc "github.com/antavelos/blockchain/src/internal/pkg/models/blockchain"
	"github.com/antavelos/blockchain/src/pkg/bus"
	"github.com/antavelos/blockchain/src/pkg/utils"
)

const (
	ShareTransactionTopic bus.Topic = iota
	RewardTransactionTopic
	RefreshDNSNodesTopic
)

func shareTxHandler(event bus.DataEvent) {
	tx := event.Data.(bc.Transaction)

	nrepo := getNodeRepo()
	nodes, _ := nrepo.GetNodes()
	responses := node_client.ShareTx(nodes, tx)
	if responses.ErrorsRatio() > 0 {
		utils.LogError("Failed to share the transaction with some nodes", responses.Errors())
	}
}

func rewardTxHandler(event bus.DataEvent) {
	tx := event.Data.(bc.Transaction)

	err := reward(tx)
	if err != nil {
		utils.LogError(err.Error())
	}
}

func refreshDNSNodesHandler() {
	// TODO: to be called from dedicated module
	err := retrieveDNSNodes()
	if err != nil {
		utils.LogError("failed to refresh DNS nodes", err.Error())
	}
}

func reward(tx bc.Transaction) error {

	brepo := getBlockchainRepo()
	tx, err := brepo.AddTx(tx)
	if err != nil {
		return utils.GenericError{Msg: "failed to add reward transaction", Extra: err}
	}

	nrepo := getNodeRepo()
	nodes, _ := nrepo.GetNodes()
	if err != nil {
		return utils.GenericError{Msg: "failed to load nodes", Extra: err}
	}

	responses := node_client.ShareTx(nodes, tx)
	if responses.ErrorsRatio() > 0 {
		return utils.GenericError{
			Msg: fmt.Sprintf("failed to share the transaction with other nodes: %v", responses.Errors()),
		}
	}

	return nil
}

func startEventLoop() {
	shareTxChan := bus.Subscribe(ShareTransactionTopic)
	rewardChan := bus.Subscribe(RewardTransactionTopic)
	refreshDNSNodesChan := bus.Subscribe(RefreshDNSNodesTopic)

	for {
		select {
		case shareTx := <-*shareTxChan:
			go shareTxHandler(shareTx)
		case rewardTx := <-*rewardChan:
			go rewardTxHandler(rewardTx)
		case <-*refreshDNSNodesChan:
			go refreshDNSNodesHandler()
		}
	}
}
