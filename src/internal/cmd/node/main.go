package main

import (
	"flag"
	"fmt"

	"github.com/antavelos/blockchain/src/internal/cmd/node/api"
	cfg "github.com/antavelos/blockchain/src/internal/cmd/node/config"
	"github.com/antavelos/blockchain/src/internal/cmd/node/eventhandlers"
	"github.com/antavelos/blockchain/src/internal/cmd/node/events"
	"github.com/antavelos/blockchain/src/internal/cmd/node/miner"
	bc_repo "github.com/antavelos/blockchain/src/internal/pkg/repos/blockchain"
	node_repo "github.com/antavelos/blockchain/src/internal/pkg/repos/node"
	wallet_repo "github.com/antavelos/blockchain/src/internal/pkg/repos/wallet"
	"github.com/antavelos/blockchain/src/pkg/db"
	"github.com/antavelos/blockchain/src/pkg/eventbus"
	"github.com/antavelos/blockchain/src/pkg/utils"
)

func main() {
	mine := flag.Bool("mine", false, "Indicates whether it will run as well as miner")
	flag.Parse()

	config, err := cfg.NewConfig()
	if err != nil {
		utils.LogFatal("Configuration error", err.Error())
	}

	blockchainRepo := bc_repo.NewBlockchainRepo(db.NewDB(config.Get("BLOCKCHAIN_FILENAME")))
	nodeRepo := node_repo.NewNodeRepo(db.NewDB(config.Get("NODES_FILENAME")))
	walletRepo := wallet_repo.NewWalletRepo(db.NewDB(config.Get("WALLETS_FILENAME")))

	bus := eventhandlers.NewEventBus(config, blockchainRepo, nodeRepo, walletRepo)

	bus.Handle(eventbus.DataEvent{Ev: events.InitNodeEvent})

	if *mine {
		mineHandler := miner.NewMineHandler(bus, config, blockchainRepo, nodeRepo, walletRepo)
		go mineHandler.RunLoop()
	}

	// TODO: add a periodic longest blockchain resolve

	apiHandler := api.NewRouteHandler(bus, blockchainRepo, nodeRepo)
	router := apiHandler.InitRouter()
	router.Run(fmt.Sprintf(":%v", config.Get("PORT")))
}
