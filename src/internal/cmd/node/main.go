package main

import (
	"flag"
	"fmt"

	"github.com/antavelos/blockchain/src/internal/cmd/node/api"
	cfg "github.com/antavelos/blockchain/src/internal/cmd/node/config"
	"github.com/antavelos/blockchain/src/internal/cmd/node/eventhandlers"
	"github.com/antavelos/blockchain/src/internal/cmd/node/events"
	"github.com/antavelos/blockchain/src/internal/cmd/node/miner"
	rep "github.com/antavelos/blockchain/src/internal/pkg/repos"
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

	repos := rep.InitRepos(rep.DBFilenames{
		BlockchainFilename: config.Get("BLOCKCHAIN_FILENAME"),
		NodeFilename:       config.Get("NODES_FILENAME"),
		WalletFilename:     config.Get("WALLETS_FILENAME"),
	})

	bus := eventhandlers.NewEventBus(config, repos)

	bus.Handle(eventbus.DataEvent{Ev: events.InitNodeEvent})

	if *mine {
		mineHandler := miner.NewMineHandler(bus, config, repos)
		go mineHandler.RunLoop()
	}

	// TODO: add a periodic longest blockchain resolve

	apiHandler := api.NewRouteHandler(bus, repos)
	router := apiHandler.InitRouter()
	router.Run(fmt.Sprintf(":%v", config.Get("PORT")))
}
