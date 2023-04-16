package clientwallet

import (
	"github.com/antavelos/blockchain/pkg/lib/rest"
	"github.com/antavelos/blockchain/pkg/models/wallet"
)

const newWalletEndpoint = "/wallets/new"

func GetNewWallet(host string) (wallet.Wallet, error) {
	requester := rest.GetRequester{
		URL: host + newWalletEndpoint,
		M:   wallet.WalletMarshaller{},
	}

	response := requester.Request()

	return response.Body.(wallet.Wallet), response.Err
}
