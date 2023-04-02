package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	bc "github.com/antavelos/blockchain/src/blockchain"
	"github.com/antavelos/blockchain/src/crypto"
	"github.com/antavelos/blockchain/src/wallet"
)

const serverPort = "4000"

func main() {
	simulate := flag.Bool("simulate", false, "simulates new wallets' and transactions'")
	serve := flag.Bool("serve", false, "runs as server by serving new wallets' requests")

	flag.Parse()

	if *simulate && *serve {
		ErrorLogger.Fatal("Cannot do both simulate and server.")
	}

	if !*simulate && !*serve {
		ErrorLogger.Fatal("No action was chosen. Exiting.")
	}

	if *simulate {
		runLoop()
	}

	if *serve {
		runServer()
	}
}

func runServer() {
	router := InitRouter()
	router.Run(fmt.Sprintf("localhost:%v", serverPort))
}

func runLoop() {
	i := 0
	for {
		if i%10 == 0 {
			w, err := CreateWallet()
			if err != nil {
				ErrorLogger.Printf("New wallet [FAIL]: %v", err.Error())
			} else {
				InfoLogger.Printf("New wallet [OK]: %v", hex.EncodeToString(w.Address))
			}
		}

		if i%2 == 0 {
			tx, err := createTransaction()
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
	wallets, err := LoadWallets()
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

	txb := bc.TransactionBody{
		Sender:    hex.EncodeToString(senderWallet.Address),
		Recipient: hex.EncodeToString(recipientWallet.Address),
		Amount:    getRandomFloat(0.001, 0.1),
	}

	txbBytes, err := json.Marshal(txb)
	if err != nil {
		return bc.Transaction{}, fmt.Errorf("failed to marshal transaction body: %v", err)
	}

	signature, err := senderWallet.Sign(crypto.HashData(txbBytes))
	if err != nil {
		return bc.Transaction{}, fmt.Errorf("failed to sign transaction body: %v", err)
	}

	tx := bc.Transaction{
		Body:      txb,
		Signature: signature,
	}

	return sendTransaction(tx)
}

func sendTransaction(tx bc.Transaction) (bc.Transaction, error) {

	host := "http://localhost:3001"
	url := host + "/transactions"

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

	if resp.StatusCode != 201 {
		return tx, fmt.Errorf("%v", string(body))
	}

	var newTx bc.Transaction
	if err := json.Unmarshal(body, &newTx); err != nil {
		return bc.Transaction{}, err
	}

	return newTx, nil
}
