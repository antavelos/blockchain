package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"strconv"
	"time"

	dns_client "github.com/antavelos/blockchain/pkg/clients/dns"
	node_client "github.com/antavelos/blockchain/pkg/clients/node"
	"github.com/antavelos/blockchain/pkg/common"
	"github.com/antavelos/blockchain/pkg/db"
	cfg "github.com/antavelos/blockchain/pkg/lib/config"
	bc "github.com/antavelos/blockchain/pkg/models/blockchain"
	w "github.com/antavelos/blockchain/pkg/models/wallet"
)

var config cfg.Config
var envVars []string = []string{
	"PORT",
	"WALLET_CREATION_INTERVAL_IN_SEC",
	"TRANSACTION_CREATION_INTERVAL_IN_SEC",
	"DNS_HOST",
	"DNS_PORT",
}
var _walletsDB *db.WalletDB

func getWalletDb() *db.WalletDB {
	if _walletsDB == nil {
		_walletsDB = db.GetWalletDb(config["WALLETS_FILENAME"])
	}
	return _walletsDB
}

func main() {
	simulate := flag.Bool("simulate", false, "simulates new wallets' and transactions'")

	flag.Parse()

	var err error
	config, err = cfg.LoadConfig(envVars)
	if err != nil {
		common.LogFatal("Configuration error", err.Error())
	}

	if *simulate {
		go runSimulation()
	}

	runServer()
}

func runServer() {
	router := InitRouter()
	router.Run(fmt.Sprintf(":%v", config["PORT"]))
}

func runSimulation() {
	wdb := getWalletDb()
	walletCreationIntervalInSec, _ := strconv.Atoi(config["WALLET_CREATION_INTERVAL_IN_SEC"])
	txCreationIntervalInSec, _ := strconv.Atoi(config["TRANSACTION_CREATION_INTERVAL_IN_SEC"])

	i := 0
	for {
		if i%walletCreationIntervalInSec == 0 {
			w, err := wdb.CreateWallet()
			if err != nil {
				common.LogError("New wallet [FAIL]", err.Error())
			} else {
				common.LogInfo("New wallet [OK]", hex.EncodeToString(w.Address))
			}
		}

		if i%txCreationIntervalInSec == 0 {
			tx, err := createTransaction()
			if err != nil {
				common.LogError("Failed to create new transaction", err.Error())
			}

			tx, err = sendTransaction(tx)
			msg := fmt.Sprintf("Transaction from %v to %v", tx.Body.Sender, tx.Body.Recipient)
			if err != nil {
				common.LogError(msg, "[FAIL]", err.Error())
			} else {
				common.LogError(msg, "[OK]", tx.Id)
			}
		}

		time.Sleep(1 * time.Second)
		i = i + 1
	}
}

func getRandomWallets() ([]w.Wallet, error) {
	wdb := getWalletDb()

	wallets, err := wdb.LoadWallets()
	if err != nil {
		return nil, common.GenericError{Msg: "failed to load wallets"}
	}

	lenWallets := len(wallets)

	if len(wallets) == 0 {
		return nil, common.GenericError{Msg: "no wallet yet"}
	}

	randomWallet1 := wallets[common.GetRandomInt(lenWallets-1)]

	var randomWallet2 w.Wallet
	for {
		randomWallet2 = wallets[common.GetRandomInt(lenWallets-1)]

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

	return bc.NewTransaction(senderWallet, recipientWallet, common.GetRandomFloat(0.001, 0.1))
}

func getDnsHost() string {
	return fmt.Sprintf("http://%v:%v", config["DNS_HOST"], config["DNS_PORT"])
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

	randomNode := nodes[common.GetRandomInt(len(nodes)-1)]

	return node_client.SendTransaction(randomNode, tx)
}
