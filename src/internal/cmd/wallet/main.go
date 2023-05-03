package main

import (
	"flag"
	"fmt"

	api "github.com/antavelos/blockchain/src/internal/cmd/wallet/api"
	cfg "github.com/antavelos/blockchain/src/internal/cmd/wallet/config"
	"github.com/antavelos/blockchain/src/internal/cmd/wallet/simulator"
	"github.com/antavelos/blockchain/src/internal/pkg/repos"
	"github.com/antavelos/blockchain/src/pkg/db"
	"github.com/antavelos/blockchain/src/pkg/utils"
)

func main() {
	simulate := flag.Bool("simulate", false, "simulates new wallets' and transactions'")
	flag.Parse()

	config, err := cfg.NewConfig()
	if err != nil {
		utils.LogFatal("Configuration error", err.Error())
	}

	walletRepo := repos.NewWalletRepo(db.NewDB(config.Get("WALLETS_FILENAME")))

	if *simulate {
		simulator := simulator.NewSimulator(config, walletRepo)
		go simulator.Run()
	}

	routeHandler := api.NewRouteHandler(walletRepo)
	router := routeHandler.InitRouter()
	router.Run(fmt.Sprintf(":%v", config.Get("PORT")))
}
