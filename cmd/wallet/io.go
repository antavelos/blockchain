package main

import (
	"os"

	"github.com/antavelos/blockchain/pkg/db"
)

func getWalletDb() *db.WalletDB {
	return &db.WalletDB{Filename: os.Getenv("WALLETS_FILENAME")}
}
