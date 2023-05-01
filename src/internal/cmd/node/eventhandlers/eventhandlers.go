package eventhandlers

import (
	"fmt"

	"github.com/antavelos/blockchain/src/internal/cmd/node/common"
	cfg "github.com/antavelos/blockchain/src/internal/cmd/node/config"
	"github.com/antavelos/blockchain/src/internal/cmd/node/events"
	node_client "github.com/antavelos/blockchain/src/internal/pkg/clients/node"
	bc "github.com/antavelos/blockchain/src/internal/pkg/models/blockchain"
	bc_repo "github.com/antavelos/blockchain/src/internal/pkg/repos/blockchain"
	node_repo "github.com/antavelos/blockchain/src/internal/pkg/repos/node"
	wallet_repo "github.com/antavelos/blockchain/src/internal/pkg/repos/wallet"
	"github.com/antavelos/blockchain/src/pkg/eventbus"
	"github.com/antavelos/blockchain/src/pkg/utils"
)

type EventHandler struct {
	Config         *cfg.Config
	CommonHandler  *common.CommonHandler
	BlockchainRepo *bc_repo.BlockchainRepo
	NodeRepo       *node_repo.NodeRepo
	WalletRepo     *wallet_repo.WalletRepo
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

	rewardTx, err := h.makeRewardTx()
	if err != nil {
		utils.LogError("failed to create reward transaction", err.Error())
		return
	}

	err = h.reward(rewardTx)
	if err != nil {
		utils.LogError("failed to reward self", err.Error())
		return
	}
	utils.LogInfo("rewarded self with", h.Config.DefaultRewardAmount)
}

func (h EventHandler) HandleRefreshDNSNodes(event eventbus.DataEvent) {
	err := h.CommonHandler.RefreshDNSNodes()
	if err != nil {
		utils.LogError("failed to refresh DNS nodes", err.Error())
	}
}

func (h EventHandler) makeRewardTx() (bc.Transaction, error) {
	if h.WalletRepo.IsEmpty() {
		return bc.Transaction{}, utils.GenericError{Msg: "node has no wallet"}
	}

	wallets, err := h.WalletRepo.GetWallets()
	if err != nil {
		return bc.Transaction{}, utils.GenericError{Msg: "failed to get wallets", Extra: err}
	}

	return bc.Transaction{
		Body: bc.TransactionBody{
			Sender:    h.Config.CoinBaseSenderAddress,
			Recipient: wallets[0].AddressString(),
			Amount:    h.Config.DefaultRewardAmount,
		},
	}, nil
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

func NewEventBus(config *cfg.Config, com *common.CommonHandler, br *bc_repo.BlockchainRepo, nr *node_repo.NodeRepo, wr *wallet_repo.WalletRepo) *eventbus.Bus {
	eh := EventHandler{Config: config, CommonHandler: com, BlockchainRepo: br, NodeRepo: nr, WalletRepo: wr}

	bus := eventbus.NewBus()
	bus.RegisterEventHandler(events.TransactionReceivedEvent, eh.HandleShareTx)
	bus.RegisterEventHandler(events.BlockMinedEvent, eh.HandleRewardTx)
	bus.RegisterEventHandler(events.ConnectionRefusedEvent, eh.HandleRefreshDNSNodes)

	return bus
}
