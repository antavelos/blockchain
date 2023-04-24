package main

import (
	"fmt"

	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/db"
	cfg "github.com/antavelos/blockchain/pkg/lib/config"
)

var config cfg.Config
var envVars []string = []string{
	"PORT",
	"NODES_FILENAME",
}

var _nodeDB *db.NodeDB

func getNodeDb() *db.NodeDB {
	if _nodeDB == nil {
		_nodeDB = db.GetNodeDb(config["NODES_FILENAME"])
	}
	return _nodeDB
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
