package main

import (
	"flag"
	"fmt"
	"log"
)

const dnsHost = "http://localhost:3000"

var port *string

func main() {
	port = flag.String("port", "8080", "the node's server port")
	initBlockchain := flag.Bool("init-blockchain", false, "determines whether to initialise the blockchain or not")
	// base := flag.Bool("base", false, "determines whether it's one of the base nodes of the blockchain")

	flag.Parse()

	if *initBlockchain {
		ioNewBlockchain()
	} else {
		nodes, err := pingDns()
		if err != nil {
			log.Fatal("Couldn't retrieve addesses from DNS %v", err.Error())
		}

		for _, node := range nodes {
			if err := ping(node); err != nil {
				log.Printf("Couldn't ping node %v: %v", node.Host, err.Error())
			}
		}
		resolveLongestBlockchain(nodes)
	}

	router := initRouter()
	router.Run(fmt.Sprintf("localhost:%v", *port))
}
