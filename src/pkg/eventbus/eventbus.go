package eventbus

import (
	"fmt"

	"github.com/antavelos/blockchain/src/pkg/utils"
)

type Event string

type DataEvent struct {
	Ev   Event
	Data any
}

type EventHandlers map[Event]func(DataEvent)

type Bus struct {
	handlers EventHandlers
}

func NewBus() *Bus {
	return &Bus{handlers: make(EventHandlers)}
}

func (b *Bus) RegisterEventHandler(ev Event, handler func(DataEvent)) {
	b.handlers[ev] = handler
}

func (b *Bus) Handle(de DataEvent) {
	handler, ok := b.handlers[de.Ev]
	if !ok {
		utils.LogError("event handler not available")
	}
	utils.LogInfo(fmt.Sprintf("Handling '%v' event", de.Ev))
	go handler(de)
}
