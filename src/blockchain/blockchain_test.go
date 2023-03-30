package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/antavelos/blockchain/src/crypto"
)

func TestNewWallet(t *testing.T) {
	wallet1, _ := crypto.NewWallet()
	wallet2, _ := crypto.NewWallet()
	txb := NewTransactionBody(hex.EncodeToString(wallet1.Address), hex.EncodeToString(wallet2.Address), 0.001)
	t.Error(txb)
	txbBytes, _ := json.Marshal(txb)
	hash := crypto.HashData(txbBytes)
	sig, _ := wallet1.Sign(hash)

	t.Errorf("\nAddress 1: %v\nAddress 2: %v\nSignature: %v", hex.EncodeToString(wallet1.Address), hex.EncodeToString(wallet2.Address), sig)
}
