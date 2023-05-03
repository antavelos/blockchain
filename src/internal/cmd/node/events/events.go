package events

import (
	"github.com/antavelos/blockchain/src/pkg/eventbus"
)

const (
	InitNodeEvent            eventbus.Event = "InitNodeEvent"
	TransactionReceivedEvent eventbus.Event = "TransactionReceivedEvent"
	BlockMinedEvent          eventbus.Event = "BlockMinedEvent"
	BlockMiningFailedEvent   eventbus.Event = "BlockMiningFailedEvent"
	ConnectionRefusedEvent   eventbus.Event = "ConnectionRefusedEvent"
)
