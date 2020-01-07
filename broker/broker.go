package broker

type Broker interface {
	Publish(topic string, message interface{}) error
	Subscribe(topic string, process func([]byte))
}
