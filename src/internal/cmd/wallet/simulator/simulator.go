package simulator

import (
	"bytes"
	"fmt"
	"time"

	cfg "github.com/antavelos/blockchain/src/internal/cmd/wallet/config"

	dns_client "github.com/antavelos/blockchain/src/internal/pkg/clients/dns"
	node_client "github.com/antavelos/blockchain/src/internal/pkg/clients/node"
	bc "github.com/antavelos/blockchain/src/internal/pkg/models/blockchain"
	w "github.com/antavelos/blockchain/src/internal/pkg/models/wallet"
	"github.com/antavelos/blockchain/src/internal/pkg/repos"
	"github.com/antavelos/blockchain/src/pkg/utils"
)

type Simulator struct {
	Config     *cfg.Config
	WalletRepo *repos.WalletRepo
}

func NewSimulator(config *cfg.Config, walletRepo *repos.WalletRepo) *Simulator {
	return &Simulator{Config: config, WalletRepo: walletRepo}
}

func (s Simulator) Run() {
	i := 0
	for {
		time.Sleep(1 * time.Second)
		fmt.Print(i)
		if i%s.Config.WalletCreationIntervalInSec == 0 {
			w, err := s.WalletRepo.CreateWallet()
			if err != nil {
				utils.LogError("New wallet [FAIL]", err.Error())
			} else {
				utils.LogInfo("New wallet [OK]", w.AddressString())
			}
		}

		if i%s.Config.TransactionCreationIntervalInSec == 0 {
			tx, err := s.createTransaction()
			if err != nil {
				utils.LogError("Failed to create new transaction", err.Error())
				continue
			}

			sentTx, err := s.sendTransaction(tx)
			msg := fmt.Sprintf("Transaction from %v to %v", tx.Body.Sender, tx.Body.Recipient)
			if err != nil {
				utils.LogError(msg, "[FAIL]", err.Error())
			} else {
				utils.LogInfo(msg, "[OK]", sentTx.Id)
			}
		}

		i = i + 1
	}
}

func (s Simulator) getRandomWallets() ([]w.Wallet, error) {
	wallets, err := s.WalletRepo.GetWallets()
	if err != nil {
		return nil, utils.GenericError{Msg: "failed to load wallets", Extra: err}
	}

	lenWallets := len(wallets)

	if len(wallets) == 0 {
		return nil, utils.GenericError{Msg: "no wallet yet"}
	}

	randomWallet1 := wallets[utils.GetRandomInt(lenWallets-1)]

	var randomWallet2 w.Wallet
	for {
		randomWallet2 = wallets[utils.GetRandomInt(lenWallets-1)]

		if !bytes.Equal(randomWallet2.Address, randomWallet1.Address) {
			break
		}
	}

	return []w.Wallet{randomWallet1, randomWallet2}, nil
}

func (s Simulator) createTransaction() (bc.Transaction, error) {
	randomWallets, err := s.getRandomWallets()
	if err != nil {
		return bc.Transaction{}, err
	}
	senderWallet := randomWallets[0]
	recipientWallet := randomWallets[1]

	return bc.NewTransaction(senderWallet, recipientWallet, utils.GetRandomFloat(0.001, 0.1))
}

func (s Simulator) getDNSHost() string {
	return fmt.Sprintf("http://%v:%v", s.Config.Get("DNS_HOST"), s.Config.Get("DNS_PORT"))
}

func (s Simulator) sendTransaction(tx bc.Transaction) (bc.Transaction, error) {
	dnsHost := s.getDNSHost()

	nodes, err := dns_client.GetDNSNodes(dnsHost)
	if err != nil {
		return tx, utils.GenericError{Msg: "failed to retrieve DNS nodes"}
	}

	if len(nodes) == 0 {
		return tx, utils.GenericError{Msg: "nodes not available"}
	}

	randomNode := nodes[utils.GetRandomInt(len(nodes)-1)]

	return node_client.SendTransaction(randomNode, tx)
}
