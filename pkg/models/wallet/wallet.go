package wallet

import (
	"encoding/hex"
	"encoding/json"

	"github.com/antavelos/blockchain/pkg/lib/crypto"
	"github.com/antavelos/blockchain/pkg/lib/rest"
)

type Wallet struct {
	Address    []byte `json:"address"`
	PrivateKey []byte `json:"privateKey"`
	PublicKey  []byte `json:"publicKey"`
}

type WalletMarshaller rest.ObjectMarshaller

func (nm WalletMarshaller) Unmarshal(data []byte) (any, error) {
	var target any
	if nm.Many {
		target = make([]Wallet, 0)
	} else {
		target = Wallet{}
	}

	err := json.Unmarshal(data, &target)

	return target, err
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
