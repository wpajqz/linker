package broker

type Broker interface {
	Publish(topic string, message interface{}) error
	Subscribe(nodeID, topic string, process func([]byte))
	UnSubscribe(nodeID string) error
}
