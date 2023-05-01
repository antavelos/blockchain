package miner

import (
	"fmt"
	"time"

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

type MineHandler struct {
	Bus            *eventbus.Bus
	Config         *cfg.Config
	CommonHandler  *common.CommonHandler
	BlockchainRepo *bc_repo.BlockchainRepo
	NodeRepo       *node_repo.NodeRepo
	WalletRepo     *wallet_repo.WalletRepo
}

func NewMineHandler(bus *eventbus.Bus, config *cfg.Config, com *common.CommonHandler, br *bc_repo.BlockchainRepo, nr *node_repo.NodeRepo, wr *wallet_repo.WalletRepo) *MineHandler {
	return &MineHandler{Bus: bus, Config: config, CommonHandler: com, BlockchainRepo: br, NodeRepo: nr, WalletRepo: wr}
}

func (h *MineHandler) shareBlock(block bc.Block) error {
	nodes, err := h.NodeRepo.GetNodes()
	if err != nil {
		return utils.GenericError{Msg: "failed to share new block"}
	}

	responses := node_client.ShareBlock(nodes, block)

	if responses.HasConnectionRefused() {
		utils.LogInfo("Refresing DNS nodes")
		h.Bus.Handle(eventbus.DataEvent{Ev: events.ConnectionRefusedEvent})
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

func (h *MineHandler) mine() (bc.Block, error) {
	blockchain, err := h.BlockchainRepo.GetBlockchain()

	if err != nil {
		return bc.Block{}, utils.GenericError{Msg: "blockchain currently not available"}
	}

	if !blockchain.HasPendingTxs() {
		return bc.Block{}, utils.GenericError{Msg: "no pending transactions found"}
	}

	block, err := blockchain.NewBlock(h.Config.DefaultTxsPerBlock)
	if err != nil {
		return bc.Block{}, err
	}

	utils.LogInfo("Mining...")
	for !block.IsValid(h.Config.DefaultMiningDifficulty) {
		block.Nonce += 1
	}
	utils.LogInfo("New block mined with nonce", block.Nonce)

	err = h.shareBlock(block)
	if err != nil {
		return bc.Block{}, err
	}

	err = blockchain.AddBlock(block)
	if err != nil {
		return bc.Block{}, err
	}

	err = h.BlockchainRepo.ReplaceBlockchain(*blockchain)
	if err != nil {
		return bc.Block{}, utils.GenericError{Msg: "failed to update blockchain"}
	}

	return block, nil
}

func (h *MineHandler) RunLoop() {
	for {
		block, err := h.mine()
		if err != nil {
			utils.LogError("New block [FAIL]", err.Error())

			utils.LogInfo("Resolving longest blockchain")
			err := h.CommonHandler.ResolveLongestBlockchain()
			if err != nil {
				utils.LogError("Failed to resolve longest blockchain", err.Error())
			}

		} else {
			utils.LogInfo("New block [OK]", block.Idx)
			h.Bus.Handle(eventbus.DataEvent{Ev: events.BlockMinedEvent})
		}

		time.Sleep(5 * time.Second)
	}
}
