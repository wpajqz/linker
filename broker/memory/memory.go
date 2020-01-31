package memory

import (
	"sync"

	"github.com/wpajqz/linker/broker"
	"github.com/wpajqz/linker/broker/memory/pubsub"
)

type (
	memoryBroker struct {
		ps sync.Map
		pf sync.Map
	}

	topicMap map[string]func([]byte)
)

func (mb *memoryBroker) Publish(topic string, message interface{}) error {
	mb.ps.Range(func(key, value interface{}) bool {
		ps := value.(*pubsub.PubSub)
		err := ps.Publish(topic, message)
		if err != nil {
			return false
		}

		return true
	})

	return nil
}

func (mb *memoryBroker) Subscribe(nodeID, topic string, process func([]byte)) error {
	actual, _ := mb.ps.LoadOrStore(nodeID, pubsub.New())
	if av, ok := actual.(*pubsub.PubSub); ok {
		actual, _ := mb.pf.LoadOrStore(nodeID, make(topicMap))
		tm := actual.(topicMap)
		if tm[topic] == nil {
			tm[topic] = process
		}

		err := av.Subscribe(topic, func(i interface{}) {
			if msg, ok := i.([]byte); ok {
				tm[topic](msg)
			}
		})

		return err
	}

	return nil
}

func (mb *memoryBroker) UnSubscribe(nodeID, topic string) error {
	if v, ok := mb.ps.Load(nodeID); ok {
		if v != nil {
			v.(*pubsub.PubSub).UnSubscribe(topic)
		}
	}

	return nil
}

func (mb *memoryBroker) UnSubscribeAll(nodeID string) error {
	mb.ps.Delete(nodeID)

	return nil
}

func NewBroker() broker.Broker {
	return &memoryBroker{}
}
