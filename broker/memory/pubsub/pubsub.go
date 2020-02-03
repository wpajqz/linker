package pubsub

import "sync"

type (
	PubSub struct {
		sync.Map
	}

	topicChain chan interface{}
)

func New() *PubSub {
	return &PubSub{}
}

func (ps *PubSub) Subscribe(topic string, process func(interface{})) error {
	actual, _ := ps.LoadOrStore(topic, make(topicChain, 1000))
	if ch, ok := actual.(topicChain); ok {
		go func(chain topicChain) {
			for msg := range ch {
				go process(msg)
			}
		}(ch)
	}

	return nil
}

func (ps *PubSub) Publish(topic string, message interface{}) error {
	actual, _ := ps.LoadOrStore(topic, make(topicChain, 1000))
	if ch, ok := actual.(topicChain); ok {
		ch <- message
	}

	return nil
}

func (ps *PubSub) UnSubscribe(topic string) {
	ps.Delete(topic)
}
