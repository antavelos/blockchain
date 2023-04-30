package events

import (
	"github.com/antavelos/blockchain/src/pkg/eventbus"
)

const (
	TransactionReceivedEvent eventbus.Event = iota
	BlockMinedEvent
	ConnectionRefusedEvent
)
