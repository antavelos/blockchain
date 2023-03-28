package main

import (
	"flag"
	"fmt"
	"log"
)

const dnsHost = "http://localhost:3000"

var Port *string

func main() {
	Port = flag.String("port", "8080", "the node's server port")
	initBlockchain := flag.Bool("init-blockchain", false, "determines whether to initialise the blockchain or not")
	// base := flag.Bool("base", false, "determines whether it's one of the base nodes of the blockchain")

	flag.Parse()

	nodes, err := pingDns()
	if err != nil {
		log.Fatal("Couldn't retrieve nodes from DNS %v", err.Error())
	}

	if err := ioSaveNodes(nodes); err != nil {
		log.Printf("Couldn't save nodes received from DNS.")
	}

	if *initBlockchain {
		ioNewBlockchain()
	} else {

		for _, node := range nodes {
			log.Printf("Pinging node %v", node.Host)
			if err := ping(node); err != nil {
				log.Printf("Couldn't ping node %v: %v", node.Host, err.Error())
			}
		}
		resolveLongestBlockchain(nodes)
	}

	router := initRouter()
	router.Run(fmt.Sprintf("localhost:%v", *Port))
}
