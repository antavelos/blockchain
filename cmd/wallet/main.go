package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	dns_client "github.com/antavelos/blockchain/pkg/clients/dns"
	node_client "github.com/antavelos/blockchain/pkg/clients/node"
	"github.com/antavelos/blockchain/pkg/common"
	com "github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/db"
	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
	w "github.com/antavelos/blockchain/pkg/models/wallet"
)

var port string = os.Getenv("PORT")

func main() {
	simulate := flag.Bool("simulate", false, "simulates new wallets' and transactions'")

	flag.Parse()

	if *simulate {
		go runSimulation()
	}

	runServer()
}

func runServer() {
	router := InitRouter()
	router.Run(fmt.Sprintf(":%v", port))
}

func runSimulation() {
	wdb := db.GetWalletDb()
	walletCreationIntervalInSec, _ := strconv.Atoi(os.Getenv("WALLET_CREATION_INTERVAL_IN_SEC"))
	txCreationIntervalInSec, _ := strconv.Atoi(os.Getenv("TRANSACTION_CREATION_INTERVAL_IN_SEC"))

	i := 0
	for {
		if i%walletCreationIntervalInSec == 0 {
			w, err := wdb.CreateWallet()
			if err != nil {
				com.ErrorLogger.Printf("New wallet [FAIL]"}
			} else {
				com.InfoLogger.Printf("New wallet [OK]: %v", hex.EncodeToString(w.Address))
			}
		}

		if i%txCreationIntervalInSec == 0 {
			tx, err := createTransaction()
			if err != nil {
				com.ErrorLogger.Printf("Failed to create new transaction"}
			}

			tx, err = sendTransaction(tx)
			if err != nil {
				com.ErrorLogger.Printf("Transaction from %v to %v [FAIL]: %v", tx.Body.Sender, tx.Body.Recipient, err.Error())
			} else {
				com.InfoLogger.Printf("Transaction from %v to %v [OK]: %v", tx.Body.Sender, tx.Body.Recipient, tx.Id)
			}
		}

		time.Sleep(1 * time.Second)
		i = i + 1
	}
}

func getRandomWallets() ([]w.Wallet, error) {
	wdb := db.GetWalletDb()

	wallets, err := wdb.LoadWallets()
	if err != nil {
		return nil, common.GenericError{Msg: "failed to load wallets"}
	}

	lenWallets := len(wallets)

	if len(wallets) == 0 {
		return nil, common.GenericError{Msg: "no wallet yet")
	}

	randomWallet1 := wallets[com.GetRandomInt(lenWallets-1)]

	var randomWallet2 w.Wallet
	for {
		randomWallet2 = wallets[com.GetRandomInt(lenWallets-1)]

		if !bytes.Equal(randomWallet2.Address, randomWallet1.Address) {
			break
		}
	}

	return []w.Wallet{randomWallet1, randomWallet2}, nil
}

func createTransaction() (bc.Transaction, error) {
	randomWallets, err := getRandomWallets()
	if err != nil {
		return bc.Transaction{}, err
	}
	senderWallet := randomWallets[0]
	recipientWallet := randomWallets[1]

	return bc.NewTransaction(senderWallet, recipientWallet, com.GetRandomFloat(0.001, 0.1))
}

func getDnsHost() string {
	return fmt.Sprintf("http://%v:%v", os.Getenv("DNS_HOST"), os.Getenv("DNS_PORT"))
}

func sendTransaction(tx bc.Transaction) (bc.Transaction, error) {
	dnsHost := getDnsHost()

	nodes, err := dns_client.GetDnsNodes(dnsHost)
	if err != nil {
		return tx, common.GenericError{Msg: "failed to retrieve DNS nodes"}
	}

	if len(nodes) == 0 {
		return tx, common.GenericError{Msg: "nodes not available"}
	}

	randomNode := nodes[com.GetRandomInt(len(nodes)-1)]

	response := node_client.SendTransaction(randomNode, tx)
	if response.Err != nil {
		return bc.Transaction{}, response.Err
	}

	return response.Body.(bc.Transaction), nil
}
