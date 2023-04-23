package bus

import (
	"sync"
)

type Topic int64

type DataEvent struct {
	topic Topic
	Data  any
}

type DataChannel chan DataEvent

type DataChannels []DataChannel

type EventBus struct {
	subscribers map[Topic]DataChannels
	mx          sync.RWMutex
}

var _eb = &EventBus{
	subscribers: map[Topic]DataChannels{},
}

func Publish(topic Topic, data any) {
	_eb.mx.RLock()
	if chans, found := _eb.subscribers[topic]; found {
		channels := append(DataChannels{}, chans...)

		go func(data DataEvent, dataChannelSlices DataChannels) {
			for _, ch := range dataChannelSlices {
				ch <- data
			}
		}(DataEvent{Data: data, topic: topic}, channels)
	}
	_eb.mx.RUnlock()
}

func Subscribe(topic Topic) *DataChannel {
	ch := make(DataChannel)

	_eb.mx.Lock()
	if prev, found := _eb.subscribers[topic]; found {
		_eb.subscribers[topic] = append(prev, ch)
	} else {
		_eb.subscribers[topic] = append([]DataChannel{}, ch)
	}
	_eb.mx.Unlock()

	return &ch
}
