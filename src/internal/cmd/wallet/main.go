package main

import (
	"bytes"
	"flag"
	"fmt"
	"time"

	dns_client "github.com/antavelos/blockchain/src/internal/pkg/clients/dns"
	node_client "github.com/antavelos/blockchain/src/internal/pkg/clients/node"
	bc "github.com/antavelos/blockchain/src/internal/pkg/models/blockchain"
	w "github.com/antavelos/blockchain/src/internal/pkg/models/wallet"
	repo "github.com/antavelos/blockchain/src/internal/pkg/repos/wallet"
	cfg "github.com/antavelos/blockchain/src/pkg/config"
	"github.com/antavelos/blockchain/src/pkg/db"
	"github.com/antavelos/blockchain/src/pkg/utils"
)

const defaultWalletCreationIntervalInSec = 300
const defaultTransactionCreationIntervalInSec = 4

var config cfg.Config
var envVars []string = []string{
	"PORT",
	"WALLET_CREATION_INTERVAL_IN_SEC",
	"TRANSACTION_CREATION_INTERVAL_IN_SEC",
	"WALLETS_FILENAME",
	"DNS_HOST",
	"DNS_PORT",
}

var _db *db.DB
var _walletRepo *repo.WalletRepo

func getDB() *db.DB {
	if _db == nil {
		_db = db.NewDB(config["WALLETS_FILENAME"])
	}
	return _db
}

func getWalletRepo() *repo.WalletRepo {
	return repo.NewWalletRepo(getDB())
}

func getWallerCreationIntervalInSec() int {
	return config.GetInteger("WALLET_CREATION_INTERVAL_IN_SEC", defaultWalletCreationIntervalInSec)
}

func getTransactionCreationIntervalInSec() int {
	return config.GetInteger("TRANSACTION_CREATION_INTERVAL_IN_SEC", defaultTransactionCreationIntervalInSec)
}

func main() {
	simulate := flag.Bool("simulate", false, "simulates new wallets' and transactions'")

	flag.Parse()

	var err error
	config, err = cfg.LoadConfig(envVars)
	if err != nil {
		utils.LogFatal("Configuration error", err.Error())
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
	wrepo := getWalletRepo()
	walletCreationIntervalInSec := getWallerCreationIntervalInSec()
	txCreationIntervalInSec := getTransactionCreationIntervalInSec()

	i := 0
	for {
		time.Sleep(1 * time.Second)
		fmt.Print(i)
		if i%walletCreationIntervalInSec == 0 {
			w, err := wrepo.CreateWallet()
			if err != nil {
				utils.LogError("New wallet [FAIL]", err.Error())
			} else {
				utils.LogInfo("New wallet [OK]", w.AddressString())
			}
		}

		if i%txCreationIntervalInSec == 0 {
			tx, err := createTransaction()
			if err != nil {
				utils.LogError("Failed to create new transaction", err.Error())
				continue
			}

			sentTx, err := sendTransaction(tx)
			msg := fmt.Sprintf("Transaction from %v to %v", tx.Body.Sender, tx.Body.Recipient)
			if err != nil {
				utils.LogError(msg, "[FAIL]", err.Error())
			} else {
				utils.LogInfo(msg, "[OK]", sentTx.Id)
			}
		}

		i = i + 1
	}
}

func getRandomWallets() ([]w.Wallet, error) {
	wrepo := getWalletRepo()

	wallets, err := wrepo.GetWallets()
	if err != nil {
		return nil, utils.GenericError{Msg: "failed to load wallets", Extra: err}
	}

	lenWallets := len(wallets)

	if len(wallets) == 0 {
		return nil, utils.GenericError{Msg: "no wallet yet"}
	}

	randomWallet1 := wallets[utils.GetRandomInt(lenWallets-1)]

	var randomWallet2 w.Wallet
	for {
		randomWallet2 = wallets[utils.GetRandomInt(lenWallets-1)]

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

	return bc.NewTransaction(senderWallet, recipientWallet, utils.GetRandomFloat(0.001, 0.1))
}

func getDNSHost() string {
	return fmt.Sprintf("http://%v:%v", config["DNS_HOST"], config["DNS_PORT"])
}

func sendTransaction(tx bc.Transaction) (bc.Transaction, error) {
	dnsHost := getDNSHost()

	nodes, err := dns_client.GetDNSNodes(dnsHost)
	if err != nil {
		return tx, utils.GenericError{Msg: "failed to retrieve DNS nodes"}
	}

	if len(nodes) == 0 {
		return tx, utils.GenericError{Msg: "nodes not available"}
	}

	randomNode := nodes[utils.GetRandomInt(len(nodes)-1)]

	return node_client.SendTransaction(randomNode, tx)
}
