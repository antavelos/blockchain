package eventhandlers

import (
	"fmt"

	"github.com/antavelos/blockchain/src/internal/cmd/node/common"
	"github.com/antavelos/blockchain/src/internal/cmd/node/events"
	node_client "github.com/antavelos/blockchain/src/internal/pkg/clients/node"
	bc "github.com/antavelos/blockchain/src/internal/pkg/models/blockchain"
	bc_repo "github.com/antavelos/blockchain/src/internal/pkg/repos/blockchain"
	node_repo "github.com/antavelos/blockchain/src/internal/pkg/repos/node"
	"github.com/antavelos/blockchain/src/pkg/eventbus"
	"github.com/antavelos/blockchain/src/pkg/utils"
)

type EventHandler struct {
	CommonHandler  *common.CommonHandler
	BlockchainRepo *bc_repo.BlockchainRepo
	NodeRepo       *node_repo.NodeRepo
}

func (h EventHandler) HandleShareTx(event eventbus.DataEvent) {
	tx := event.Data.(bc.Transaction)

	nodes, _ := h.NodeRepo.GetNodes()
	responses := node_client.ShareTx(nodes, tx)
	if responses.ErrorsRatio() > 0 {
		utils.LogError("Failed to share the transaction with some nodes", responses.Errors())
	}
}

func (h EventHandler) HandleRewardTx(event eventbus.DataEvent) {
	tx := event.Data.(bc.Transaction)

	err := h.reward(tx)
	if err != nil {
		utils.LogError(err.Error())
	}
}

func (h EventHandler) HandleRefreshDNSNodes(event eventbus.DataEvent) {
	err := h.CommonHandler.RefreshDNSNodes()
	if err != nil {
		utils.LogError("failed to refresh DNS nodes", err.Error())
	}
}

func (h EventHandler) reward(tx bc.Transaction) error {

	tx, err := h.BlockchainRepo.AddTx(tx)
	if err != nil {
		return utils.GenericError{Msg: "failed to add reward transaction", Extra: err}
	}

	nodes, _ := h.NodeRepo.GetNodes()
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

func NewEventBus(com *common.CommonHandler, br *bc_repo.BlockchainRepo, nr *node_repo.NodeRepo) *eventbus.Bus {
	eh := EventHandler{CommonHandler: com, BlockchainRepo: br, NodeRepo: nr}

	bus := eventbus.NewBus()
	bus.RegisterEventHandler(events.TransactionReceivedEvent, eh.HandleShareTx)
	bus.RegisterEventHandler(events.BlockMinedEvent, eh.HandleRewardTx)
	bus.RegisterEventHandler(events.ConnectionRefusedEvent, eh.HandleRefreshDNSNodes)

	return bus
}
