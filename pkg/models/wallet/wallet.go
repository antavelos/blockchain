package wallet

import (
	"encoding/hex"
	"encoding/json"

	"github.com/antavelos/blockchain/pkg/lib/crypto"
)

type Wallet struct {
	Address    []byte `json:"address"`
	PrivateKey []byte `json:"privateKey"`
	PublicKey  []byte `json:"publicKey"`
}

func Unmarshal(data []byte) (wallet Wallet, err error) {
	err = json.Unmarshal(data, &wallet)
	return
}

func UnmarshalMany(data []byte) (wallets []Wallet, err error) {
	err = json.Unmarshal(data, &wallets)
	return
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

func (w Wallet) AddressString() string {
	return hex.EncodeToString(w.Address)
}

func (w Wallet) VerifySignature(data []byte, signature []byte) bool {
	return crypto.VerifySignature(data, w.PublicKey, signature)
}
