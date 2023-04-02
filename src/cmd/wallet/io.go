package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/antavelos/blockchain/src/wallet"
)

func getWalletsFilename() string {
	return "./wallets.json"
}

func SaveWallet(w wallet.Wallet) error {
	wallets, err := LoadWallets()
	if err != nil {
		return err
	}

	wallets = append(wallets, w)

	marshalled, err := json.MarshalIndent(wallets, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal wallets: %v", err.Error())
	}

	return os.WriteFile(getWalletsFilename(), marshalled, os.ModePerm)
}

func LoadWallets() ([]wallet.Wallet, error) {
	var wallets []wallet.Wallet

	file, err := os.ReadFile(getWalletsFilename())
	if err != nil {
		return nil, fmt.Errorf("failed to load wallets from file: %v", err.Error())
	}

	json.Unmarshal(file, &wallets)

	return wallets, nil
}

func CreateWallet() (*wallet.Wallet, error) {

	w, err := wallet.NewWallet()
	if err != nil {
		return &wallet.Wallet{}, fmt.Errorf("failed to create a new wallet: %v", err.Error())
	}

	err = SaveWallet(*w)
	if err != nil {
		return &wallet.Wallet{}, fmt.Errorf("failed to save new wallet: %v", err.Error())
	}

	return w, nil
}
