package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	bc "github.com/antavelos/blockchain/pkg/blockchain"
)

func getDnsHost() string {
	return fmt.Sprintf("http://%v:%v", os.Getenv("DNS_HOST"), os.Getenv("DNS_PORT"))
}

func getWalletsHost() string {
	return fmt.Sprintf("http://%v:%v", os.Getenv("WALLETS_HOST"), os.Getenv("WALLETS_PORT"))
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

	GetDnsNodes()

	PingNodes()

	ResolveLongestBlockchain()

	GetNewWallet()
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

			err := reward()
			if err != nil {
				ErrorLogger.Printf("Failed to reward node: %v", err.Error())
			}

		}

		time.Sleep(5 * time.Second)
		i = i + 1
	}
}

func reward() error {
	wallet, err := ioGetWallet()
	if err != nil {
		return fmt.Errorf("node wallet not available: %v", err)
	}

	tx := bc.Transaction{
		Body: bc.TransactionBody{
			Sender:    "0",
			Recipient: hex.EncodeToString(wallet.Address),
			Amount:    1.0,
		},
	}

	tx, err = ioAddTx(tx)
	if err != nil {
		return fmt.Errorf("failed to add reward transaction: %v", err.Error())
	}

	nodeErrors := ShareTx(tx)
	errorStrings := ErrorsToStrings(nodeErrors)
	if len(errorStrings) > 0 {
		return fmt.Errorf("failed to share the transaction with other nodes: \n%v", strings.Join(errorStrings, "\n"))
	}

	return nil
}
