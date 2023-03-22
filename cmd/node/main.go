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
	flag.Parse()

	nodes, err := pingDns()
	if err != nil {
		log.Fatal("Couldn't retrieve addesses from DNS %v", err.Error())
	}

	for _, node := range nodes {
		err := ping(node)
		if err != nil {
			log.Printf("Couldn't ping node %v: %v", node.Host, err.Error())
		}
	}

	getMaxBlockchain(nodes)

	router := initRouter()
	router.Run(fmt.Sprintf("localhost:%v", *port))
}
