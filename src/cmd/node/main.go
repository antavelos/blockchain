package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	bc "github.com/antavelos/blockchain/src/blockchain"
)

const dnsHost = "http://localhost:3000"

var Port *string

func main() {
	Port = flag.String("port", "8080", "The node's server port")
	serve := flag.Bool("serve", false, "Indicates whether it will run as server")
	mine := flag.Bool("mine", false, "Indicates whether it will run as miner")
	init := flag.Bool("init", false, "Initialises the blockchain. Existing blockchain will be overriden. Overrules other options.")

	flag.Parse()

	if *init {
		ioNewBlockchain()
	}

	if *serve && *mine {
		log.Fatal("Cannot do both serve and mine.")
	}

	if !*serve && !*mine {
		log.Fatal("No action was chosen. Possible actions: 1) serve, 2) mine. Exiting.")
	}

	nodes := getDNSNodes()
	pingNodes(nodes)
	resolveLongestBlockchain(nodes)

	if *serve {
		runServer()
	}

	if *mine {
		runMiningLoop()
	}
}

func runMiningLoop() {
	i := 0
	for {
		block, err := Mine()
		if err != nil {
			ErrorLogger.Printf("New block [FAIL]: %v", err.Error())
		} else {
			InfoLogger.Printf("New block [OK]: %v", block.Idx)
		}

		time.Sleep(5 * time.Second)
		i = i + 1
	}
}

func runServer() {
	router := initRouter()
	router.Run(fmt.Sprintf("localhost:%v", *Port))
}

func getDNSNodes() []bc.Node {
	nodes, err := pingDns()
	if err != nil {
		log.Fatalf("Couldn't retrieve nodes from DNS %v", err.Error())
	}

	if err := ioSaveNodes(nodes); err != nil {
		log.Printf("Couldn't save nodes received from DNS.")
	}

	return nodes
}

func pingNodes(nodes []bc.Node) {
	for _, node := range nodes {
		log.Printf("Pinging node %v", node.Host)
		if err := ping(node); err != nil {
			log.Printf("Couldn't ping node %v: %v", node.Host, err.Error())
		}
	}
}
