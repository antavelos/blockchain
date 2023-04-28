package main

import (
	"fmt"

	node_client "github.com/antavelos/blockchain/pkg/clients/node"
	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/lib/bus"
	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
)

const (
	ShareTransactionTopic bus.Topic = iota
	RewardTransactionTopic
	RefreshDnsNodesTopic
)

func shareTxHandler(event bus.DataEvent) {
	tx := event.Data.(bc.Transaction)

	nrepo := getNodeRepo()
	nodes, _ := nrepo.GetNodes()
	responses := node_client.ShareTx(nodes, tx)
	if responses.ErrorsRatio() > 0 {
		common.LogError("Failed to share the transaction with some nodes", responses.Errors())
	}
}

func rewardTxHandler(event bus.DataEvent) {
	tx := event.Data.(bc.Transaction)

	err := reward(tx)
	if err != nil {
		common.LogError(err.Error())
	}
}

func refreshDnsNodesHandler() {
	// TODO: to be called from dedicated module
	err := retrieveDnsNodes()
	if err != nil {
		common.LogError("failed to refresh DNS nodes", err.Error())
	}
}

func reward(tx bc.Transaction) error {

	brepo := getBlockchainRepo()
	tx, err := brepo.AddTx(tx)
	if err != nil {
		return common.GenericError{Msg: "failed to add reward transaction", Extra: err}
	}

	nrepo := getNodeRepo()
	nodes, _ := nrepo.GetNodes()
	if err != nil {
		return common.GenericError{Msg: "failed to load nodes", Extra: err}
	}

	responses := node_client.ShareTx(nodes, tx)
	if responses.ErrorsRatio() > 0 {
		return common.GenericError{
			Msg: fmt.Sprintf("failed to share the transaction with other nodes: %v", responses.Errors()),
		}
	}

	return nil
}

func startEventLoop() {
	shareTxChan := bus.Subscribe(ShareTransactionTopic)
	rewardChan := bus.Subscribe(RewardTransactionTopic)
	refreshDnsNodesChan := bus.Subscribe(RefreshDnsNodesTopic)

	for {
		select {
		case shareTx := <-*shareTxChan:
			go shareTxHandler(shareTx)
		case rewardTx := <-*rewardChan:
			go rewardTxHandler(rewardTx)
		case <-*refreshDnsNodesChan:
			go refreshDnsNodesHandler()
		}
	}
}
