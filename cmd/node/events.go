package main

import (
	"fmt"
	"strings"

	node_client "github.com/antavelos/blockchain/pkg/clients/node"
	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/lib/bus"
	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
)

const (
	ShareTransaction bus.Topic = iota
	RewardTransaction
)

func shareTxHandler(event bus.DataEvent) {
	tx := event.Data.(bc.Transaction)

	ndb := getNodeDb()
	nodes, _ := ndb.LoadNodes()
	responses := node_client.ShareTx(nodes, tx)
	if responses.ErrorsRatio() > 0 {
		common.LogError("Failed to share the transaction with some nodes", responses.ErrorStrings())
	}
}

func rewardTxHandler(event bus.DataEvent) {
	tx := event.Data.(bc.Transaction)

	err := reward(tx)
	if err != nil {
		common.LogError(err.Error())
	}
}

func reward(tx bc.Transaction) error {

	tx, err := ioAddTx(tx)
	if err != nil {
		return common.GenericError{Msg: "failed to add reward transaction", Extra: err}
	}

	ndb := getNodeDb()
	nodes, err := ndb.LoadNodes()
	if err != nil {
		return common.GenericError{Msg: "failed to load nodes", Extra: err}
	}

	responses := node_client.ShareTx(nodes, tx)
	if responses.ErrorsRatio() > 0 {
		return common.GenericError{
			Msg: fmt.Sprintf("failed to share the transaction with other nodes\n %v", strings.Join(responses.ErrorStrings(), "\n")),
		}
	}

	return nil
}

func startEventLoop() {
	shareTxChan := bus.Subscribe(ShareTransaction)
	rewardChan := bus.Subscribe(RewardTransaction)

	for {
		select {
		case shareTx := <-*shareTxChan:
			go shareTxHandler(shareTx)
		case rewardTx := <-*rewardChan:
			go rewardTxHandler(rewardTx)
		}
	}

}
