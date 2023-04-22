package main

import (
	"fmt"
	"strings"

	node_client "github.com/antavelos/blockchain/pkg/clients/node"
	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/db"
	"github.com/antavelos/blockchain/pkg/lib/bus"
	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
)

const shareTxTopic = "SHARE_TRANSACTION"
const rewardTopic = "REWARD_TRANSACTION"

func shareTxHandler(event bus.DataEvent) {
	tx := event.Data.(bc.Transaction)

	ndb := db.GetNodeDb()
	nodes, _ := ndb.LoadNodes()
	responses := node_client.ShareTx(nodes, tx)
	if responses.ErrorsRatio() > 0 {
		common.ErrorLogger.Printf("Failed to share the transaction with some nodes: \n%v", responses.ErrorStrings())
	}
}

func rewardHandler(event bus.DataEvent) error {
	tx := event.Data.(bc.Transaction)

	tx, err := ioAddTx(tx)
	if err != nil {
		return common.GenericError{Msg: "failed to add reward transaction", Extra: err}
	}

	ndb := db.GetNodeDb()
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
	shareTxChan := bus.Subscribe(shareTxTopic)
	rewardChan := bus.Subscribe(rewardTopic)

	for {
		select {
		case shareTx := <-*shareTxChan:
			go shareTxHandler(shareTx)
		case rewardTx := <-*rewardChan:
			go rewardHandler(rewardTx)
		}
	}

}
