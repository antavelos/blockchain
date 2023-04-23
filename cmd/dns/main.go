package main

import (
	"fmt"

	"github.com/antavelos/blockchain/pkg/common"
	cfg "github.com/antavelos/blockchain/pkg/lib/config"
)

var config cfg.Config
var envVars []string = []string{
	"PORT",
}

func main() {

	var err error
	config, err = cfg.LoadConfig(envVars)
	if err != nil {
		common.LogFatal("Configuration error", err.Error())
	}

	router := InitRouter()

	router.Run(fmt.Sprintf(":%v", config["PORT"]))
}
