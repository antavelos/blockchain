package clientwallet

import (
	"github.com/antavelos/blockchain/pkg/lib/rest"
	"github.com/antavelos/blockchain/pkg/models/wallet"
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
