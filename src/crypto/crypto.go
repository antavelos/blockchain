package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"

	eth "github.com/ethereum/go-ethereum/crypto"
)

type Wallet struct {
	Address    []byte
	PrivateKey []byte
	PublicKey  []byte
}

func NewWallet() (*Wallet, error) {
	privateKey, err := GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	publicKey, err := PublicKeyFromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		Address:    AddressFromPublicKey(publicKey),
		PrivateKey: MarshalPrivateKey(privateKey),
		PublicKey:  MarshalPublicKey(publicKey),
	}, nil
}

func (w Wallet) Sign(message []byte) (string, error) {
	signature, err := SignData(message, w.PrivateKey)

	return hex.EncodeToString(signature), err
}

func GeneratePrivateKey() (*ecdsa.PrivateKey, error) {
	return eth.GenerateKey()
}

func MarshalPrivateKey(privateKey *ecdsa.PrivateKey) []byte {
	return eth.FromECDSA(privateKey)
}

func UnmarshalPrivateKey(privateKey []byte) (*ecdsa.PrivateKey, error) {
	return eth.ToECDSA(privateKey)
}

func MarshalPublicKey(publicKey *ecdsa.PublicKey) []byte {
	return eth.FromECDSAPub(publicKey)
}

func UnmarshalPublicKey(publicKey []byte) (*ecdsa.PublicKey, error) {
	return eth.UnmarshalPubkey(publicKey)
}

func PublicKeyFromPrivateKey(privateKey *ecdsa.PrivateKey) (*ecdsa.PublicKey, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	return publicKeyECDSA, nil
}

func HashData(data []byte) []byte {
	hash := eth.Keccak256Hash(data)

	return hash.Bytes()
}

func SignData(data []byte, privateKey []byte) ([]byte, error) {
	privateKeyECDSA, err := UnmarshalPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return eth.Sign(data, privateKeyECDSA)
}

func PublicKeyFromSignature(data []byte, signature []byte) ([]byte, error) {
	return eth.Ecrecover(data, signature)
}

func VerifySignature(data []byte, publicKey []byte, signature []byte) bool {
	signatureNoRecoverID := signature[:len(signature)-1] // remove recovery id
	return eth.VerifySignature(publicKey, data, signatureNoRecoverID)
}

func AddressFromPublicKey(publicKey *ecdsa.PublicKey) []byte {
	address := eth.PubkeyToAddress(*publicKey)
	return address.Bytes()
}
