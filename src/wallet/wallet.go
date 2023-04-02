package wallet

import (
	"encoding/hex"

	"github.com/antavelos/blockchain/src/crypto"
)

type Wallet struct {
	Address    []byte `json:"address"`
	PrivateKey []byte `json:"privateKey"`
	PublicKey  []byte `json:"publicKey"`
}

func NewWallet() (*Wallet, error) {
	privateKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	publicKey, err := crypto.PublicKeyFromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		Address:    crypto.AddressFromPublicKey(publicKey),
		PrivateKey: crypto.MarshalPrivateKey(privateKey),
		PublicKey:  crypto.MarshalPublicKey(publicKey),
	}, nil
}

func (w Wallet) Sign(message []byte) (string, error) {
	signature, err := crypto.SignData(message, w.PrivateKey)

	return hex.EncodeToString(signature), err
}
