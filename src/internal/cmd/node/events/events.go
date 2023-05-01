package events

import bc "github.com/antavelos/blockchain/src/internal/pkg/models/blockchain"

type TransactionReceivedEvent struct {
	Tx bc.Transaction
}

func (e TransactionReceivedEvent) Name() string {
	return "TransactionReceivedEvent"
}

func (e TransactionReceivedEvent) Data() any {
	return e.Tx
}

type BlockMinedEvent struct{}

func (e BlockMinedEvent) Name() string {
	return "BlockMinedEvent"
}

func (e BlockMinedEvent) Data() any {
	return nil
}

type ConnectionRefusedEvent struct{}

func (e ConnectionRefusedEvent) Name() string {
	return "ConnectionRefusedEvent"
}

func (e ConnectionRefusedEvent) Data() any {
	return nil
}
