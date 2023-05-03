package main

import (
	"fmt"

	cfg "github.com/antavelos/blockchain/src/pkg/config"
	"github.com/antavelos/blockchain/src/pkg/utils"
)

var config cfg.Config
var envVars []string = []string{
	"PORT",
	"DNS_PORT",
	"DNS_HOST",
}

func main() {
	var err error
	config, err = cfg.LoadConfig(envVars)
	if err != nil {
		utils.LogFatal("Configuration error", err.Error())
	}

	router := initRouter()
	router.Run(fmt.Sprintf(":%v", config["PORT"]))
}
