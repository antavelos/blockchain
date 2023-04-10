package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	bc "github.com/antavelos/blockchain/src/blockchain"
)

func getDnsHost() string {
	return fmt.Sprintf("http://%v:%v", os.Getenv("DNS_HOST"), os.Getenv("DNS_PORT"))
}

func getSelfHost() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Printf("Cannot get IP: " + err.Error() + "\n")
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

func getSelfPort() string {
	return os.Getenv("PORT")
}

func main() {
	mine := flag.Bool("mine", false, "Indicates whether it will run as miner")
	init := flag.Bool("init", false, "Initialises the blockchain. Existing blockchain will be overriden. Overrules other options.")

	flag.Parse()

	if *init {
		ioNewBlockchain()
	}

	initNode()

	if *mine {
		go runMiningLoop()
	}

	router := initRouter()
	router.Run(fmt.Sprintf(":%v", getSelfPort()))
}

func initNode() {
	IntroduceToDns()

	getDNSNodes()

	Ping()

	ResolveLongestBlockchain()
}

func runMiningLoop() {
	i := 0
	for {
		block, err := Mine()
		if err != nil {
			ErrorLogger.Printf("New block [FAIL]: %v", err.Error())

			InfoLogger.Println("Resolving longest blockchain")
			err := ResolveLongestBlockchain()
			if err != nil {
				ErrorLogger.Printf("Failed to resolve longest blockchain: %v", err.Error())
			}

		} else {
			InfoLogger.Printf("New block [OK]: %v", block.Idx)
		}

		time.Sleep(5 * time.Second)
		i = i + 1
	}
}

func getDNSNodes() []bc.Node {
	ndb := getNodeDb()

	nodes, err := GetDnsNodes()
	if err != nil {
		log.Fatalf("Couldn't retrieve nodes from DNS %v", err.Error())
	}

	nodes = Filter(nodes, func(n bc.Node) bool {
		return n.GetPort() != getSelfPort()
	})

	if err := ndb.SaveNodes(nodes); err != nil {
		log.Printf("Couldn't save nodes received from DNS.")
	}

	return nodes
}
