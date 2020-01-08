package redis

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/wpajqz/linker/broker"
)

type redisBroker struct {
	client *redis.Client
	pb     map[string]*redis.PubSub
}

func (rb *redisBroker) Publish(topic string, message interface{}) error {
	count, err := rb.client.Publish(topic, message).Result()
	if err != nil {
		return err
	}

	fmt.Println("Subscribe count", count)
	return nil
}

func (rb *redisBroker) Subscribe(nodeID, topic string, process func([]byte)) {
	ps := rb.client.Subscribe(topic)
	rb.pb[nodeID] = ps

	ch := rb.pb[nodeID].Channel()
	go func() {
		for msg := range ch {
			go process([]byte(msg.Payload))
		}
	}()
}

func (rb *redisBroker) UnSubscribe(nodeID string) error {
	return rb.pb[nodeID].Close()
}

func NewBroker(opts ...Option) broker.Broker {
	options := Options{
		Address: "127.0.0.1:6379",
	}

	for _, o := range opts {
		o(&options)
	}

	rc := redis.NewClient(&redis.Options{
		Addr:     options.Address,
		Password: options.Password,
		DB:       options.DB,
	})

	return &redisBroker{client: rc, pb: make(map[string]*redis.PubSub)}
}
