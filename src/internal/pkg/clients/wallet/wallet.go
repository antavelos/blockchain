package clientwallet

import (
	"github.com/antavelos/blockchain/src/internal/pkg/models/wallet"
	"github.com/antavelos/blockchain/src/pkg/rest"
)

const newWalletEndpoint = "/wallets/new"

func GetNewWallet(host string) (wallet.Wallet, error) {
	requester := rest.GetRequester{
		URL: host + newWalletEndpoint,
	}

	response := requester.Request()

	if response.Err != nil {
		return wallet.Wallet{}, response.Err
	}

	return wallet.Unmarshal(response.Body)
}
