package miner

import (
	"fmt"
	"time"

	cfg "github.com/antavelos/blockchain/src/internal/cmd/node/config"
	"github.com/antavelos/blockchain/src/internal/cmd/node/events"
	node_client "github.com/antavelos/blockchain/src/internal/pkg/clients/node"
	bc "github.com/antavelos/blockchain/src/internal/pkg/models/blockchain"
	rep "github.com/antavelos/blockchain/src/internal/pkg/repos"
	"github.com/antavelos/blockchain/src/pkg/eventbus"
	"github.com/antavelos/blockchain/src/pkg/utils"
)

type Miner struct {
	Bus    *eventbus.Bus
	Config *cfg.Config
	Repos  *rep.Repos
}

func NewMiner(bus *eventbus.Bus, config *cfg.Config, repos *rep.Repos) *Miner {
	return &Miner{Bus: bus, Config: config, Repos: repos}
}

func (m *Miner) shareBlock(block bc.Block) error {
	nodes, err := m.Repos.NodeRepo.GetNodes()
	if err != nil {
		return utils.GenericError{Msg: "failed to share new block"}
	}

	responses := node_client.ShareBlock(nodes, block)

	if responses.HasConnectionRefused() {
		m.Bus.Handle(eventbus.DataEvent{Ev: events.ConnectionRefusedEvent})
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

func (m *Miner) mine() (bc.Block, error) {
	blockchain, err := m.Repos.BlockchainRepo.GetBlockchain()

	if err != nil {
		return bc.Block{}, utils.GenericError{Msg: "blockchain currently not available"}
	}

	if !blockchain.HasPendingTxs() {
		return bc.Block{}, utils.GenericError{Msg: "no pending transactions found"}
	}

	block, err := blockchain.NewBlock(m.Config.DefaultTxsPerBlock)
	if err != nil {
		return bc.Block{}, err
	}

	utils.LogInfo("Mining...")
	for !block.IsValid(m.Config.DefaultMiningDifficulty) {
		block.Nonce += 1
	}
	utils.LogInfo("New block mined with nonce", block.Nonce)

	err = m.shareBlock(block)
	if err != nil {
		return bc.Block{}, err
	}

	err = blockchain.AddBlock(block)
	if err != nil {
		return bc.Block{}, err
	}

	err = m.Repos.BlockchainRepo.ReplaceBlockchain(*blockchain)
	if err != nil {
		return bc.Block{}, utils.GenericError{Msg: "failed to update blockchain"}
	}

	return block, nil
}

func (m *Miner) Run() {
	for {
		block, err := m.mine()
		if err != nil {
			utils.LogError("New block [FAIL]", err.Error())
			m.Bus.Handle(eventbus.DataEvent{Ev: events.BlockMiningFailedEvent})
		} else {
			utils.LogInfo("New block [OK]", block.Idx)
			m.Bus.Handle(eventbus.DataEvent{Ev: events.BlockMinedEvent})
		}

		time.Sleep(5 * time.Second)
	}
}
