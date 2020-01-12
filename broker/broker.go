package broker

type Broker interface {
	Publish(topic string, message interface{}) error
	Subscribe(nodeID, topic string, process func([]byte)) error
	UnSubscribe(nodeID, topic string) error
	UnSubscribeAll(nodeID string) error
}
