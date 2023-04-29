package main

import (
	"fmt"

	cnf "github.com/antavelos/blockchain/src/pkg/config"
)

var config cnf.Config
var envVars []string = []string{
	"PORT",
	"DNS_PORT",
	"DNS_HOST",
}

func main() {
	router := initRouter()
	router.Run(fmt.Sprintf(":%v", config["PORT"]))
}
