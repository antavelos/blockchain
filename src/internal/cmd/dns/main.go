package main

import (
	"fmt"

	repo "github.com/antavelos/blockchain/src/internal/pkg/repos/node"
	cfg "github.com/antavelos/blockchain/src/pkg/config"
	"github.com/antavelos/blockchain/src/pkg/db"
	"github.com/antavelos/blockchain/src/pkg/utils"
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
		utils.LogFatal("Configuration error", err.Error())
	}

	router := InitRouter()

	router.Run(fmt.Sprintf(":%v", config["PORT"]))
}