package main

import (
	"fmt"

	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/db"
	cfg "github.com/antavelos/blockchain/pkg/lib/config"
	repo "github.com/antavelos/blockchain/pkg/repo/node"
)

var config cfg.Config
var envVars []string = []string{
	"PORT",
	"NODES_FILENAME",
}

var _db *db.DB

func getDB() *db.DB {
	if _db == nil {
		_db = db.NewDB(config["NODES_FILENAME"])
	}
	return _db
}

func getNodeRepo() *repo.NodeRepo {
	return repo.NewNodeRepo(getDB())
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
