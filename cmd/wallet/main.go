package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	bc "github.com/antavelos/blockchain/pkg/blockchain"
	"github.com/antavelos/blockchain/pkg/wallet"
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
	wdb := getWalletDb()
	walletCreationIntervalInSec, _ := strconv.Atoi(os.Getenv("WALLET_CREATION_INTERVAL_IN_SEC"))
	txCreationIntervalInSec, _ := strconv.Atoi(os.Getenv("TRANSACTION_CREATION_INTERVAL_IN_SEC"))

	i := 0
	for {
		if i%walletCreationIntervalInSec == 0 {
			w, err := wdb.CreateWallet()
			if err != nil {
				ErrorLogger.Printf("New wallet [FAIL]: %v", err.Error())
			} else {
				InfoLogger.Printf("New wallet [OK]: %v", hex.EncodeToString(w.Address))
			}
		}

		if i%txCreationIntervalInSec == 0 {
			tx, err := createTransaction()
			if err != nil {
				ErrorLogger.Printf("Failed to create new transaction: %v", err.Error())
			}

			tx, err = sendTransaction(tx)
			if err != nil {
				ErrorLogger.Printf("Transaction from %v to %v [FAIL]: %v", tx.Body.Sender, tx.Body.Recipient, err.Error())
			} else {
				InfoLogger.Printf("Transaction from %v to %v [OK]: %v", tx.Body.Sender, tx.Body.Recipient, tx.Id)
			}
		}

		time.Sleep(1 * time.Second)
		i = i + 1
	}
}

func getRandomInt(numRange int) int {
	if numRange == 0 {
		return 0
	}

	rand.Seed(time.Now().UnixNano())

	return rand.Intn(numRange)
}

func getRandomFloat(min, max float64) float64 {
	rand.Seed(time.Now().UnixNano())

	return min + rand.Float64()*(max-min)
}

func getRandomWallets() ([]wallet.Wallet, error) {
	wdb := getWalletDb()

	wallets, err := wdb.LoadWallets()
	if err != nil {
		return nil, fmt.Errorf("failed to load wallets: %v", err.Error())
	}

	lenWallets := len(wallets)

	if len(wallets) == 0 {
		return nil, fmt.Errorf("no wallet yet")
	}

	randomWallet1 := wallets[getRandomInt(lenWallets-1)]

	var randomWallet2 wallet.Wallet
	for {
		randomWallet2 = wallets[getRandomInt(lenWallets-1)]

		if !bytes.Equal(randomWallet2.Address, randomWallet1.Address) {
			break
		}
	}

	return []wallet.Wallet{randomWallet1, randomWallet2}, nil
}

func createTransaction() (bc.Transaction, error) {
	randomWallets, err := getRandomWallets()
	if err != nil {
		return bc.Transaction{}, err
	}
	senderWallet := randomWallets[0]
	recipientWallet := randomWallets[1]

	return bc.NewTransaction(senderWallet, recipientWallet, getRandomFloat(0.001, 0.1))
}

func getDnsHost() string {
	return fmt.Sprintf("http://%v:%v", os.Getenv("DNS_HOST"), os.Getenv("DNS_PORT"))
}

func getDnsNodes() ([]bc.Node, error) {
	var nodes []bc.Node
	url := getDnsHost() + "/nodes"

	resp, err := http.Get(url)
	if err != nil {
		return nodes, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nodes, err
	}

	if err := json.Unmarshal(body, &nodes); err != nil {
		return nodes, err
	}

	return nodes, nil
}

func sendTransaction(tx bc.Transaction) (bc.Transaction, error) {

	nodes, err := getDnsNodes()
	if err != nil {
		return tx, fmt.Errorf("failed to retrieve DNS nodes: %v", err.Error())
	}

	if len(nodes) == 0 {
		return tx, errors.New("DNS nodes not available")
	}

	randomNode := nodes[getRandomInt(len(nodes)-1)]

	url := randomNode.Host + "/transactions"

	jsonTx, err := json.Marshal(tx)
	if err != nil {
		return tx, fmt.Errorf("failed to marshal transaction: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonTx))
	if err != nil {
		return tx, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return tx, err
	}

	if resp.StatusCode != http.StatusCreated {
		return tx, fmt.Errorf("%v", string(body))
	}

	var newTx bc.Transaction
	if err := json.Unmarshal(body, &newTx); err != nil {
		return bc.Transaction{}, err
	}

	return newTx, nil
}
