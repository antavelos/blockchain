package eventbus

import "github.com/antavelos/blockchain/src/pkg/utils"

type Event interface {
	Name() string
	Data() any
}

// type DataEvent struct {
// 	Ev   Event
// 	Data any
// }

type EventHandlers map[string]func(Event)

type Bus struct {
	handlers EventHandlers
}

func NewBus() *Bus {
	return &Bus{handlers: make(EventHandlers)}
}

func (b *Bus) RegisterEventHandler(event string, handler func(Event)) {
	b.handlers[event] = handler
}

func (b *Bus) Handle(event Event) {
	handler, ok := b.handlers[event.Name()]
	if !ok {
		utils.LogError("event handler not available")
	}

	go handler(event)
}
