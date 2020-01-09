package client

import (
	"context"
	"fmt"
	"time"

	"github.com/silenceper/pool"
	"github.com/wpajqz/linker/client/export"
	"github.com/wpajqz/linker/plugin"
)

var (
	defaultClient  *Client
	defaultNetwork = "tcp"
)

type (
	Client struct {
		network                 string
		dialTimeout             time.Duration
		address                 []string
		availableAddress        chan string
		maxPayload              int
		initialCap              int
		maxCap                  int
		idleTimeout             time.Duration
		clientPool              pool.Pool
		onOpen, onClose         func()
		onError                 func(error)
		pluginForPacketSender   []plugin.PacketPlugin
		pluginForPacketReceiver []plugin.PacketPlugin
	}
)

func NewClient(address []string, opts ...Option) (*Client, error) {
	options := options{
		network:     defaultNetwork,
		dialTimeout: 3 * time.Second,
		maxPayload:  10 * 1024 * 1024,
		initialCap:  10,
		maxCap:      30,
	}

	for _, o := range opts {
		o(&options)
	}

	defaultClient = &Client{
		network:          options.network,
		dialTimeout:      options.dialTimeout,
		address:          address,
		availableAddress: make(chan string, len(address)),
		maxPayload:       options.maxPayload,
		initialCap:       options.initialCap,
		maxCap:           options.maxCap,
		idleTimeout:      options.idleTimeout,
		onOpen:           options.onOpen,
		onClose:          options.onClose,
		onError:          options.onError,
	}

	defaultClient.fillAddress(address)

	p, err := defaultClient.newExportPool()
	if err != nil {
		return nil, err
	}

	defaultClient.clientPool = p

	return defaultClient, err
}

func Session() (*export.Client, error) {
	return defaultClient.Session()
}

func (c *Client) Session() (*export.Client, error) {
	v, err := c.clientPool.Get()
	if err != nil {
		return nil, err
	}
	defer c.clientPool.Put(v)

	return v.(*export.Client), nil
}

func (c *Client) fillAddress(address []string) {
	for _, v := range address {
		defaultClient.availableAddress <- v
	}
}

func (c *Client) getAddress() (string, error) {
	ctx, _ := context.WithTimeout(context.Background(), c.dialTimeout)
	for {
		select {
		case addr := <-c.availableAddress:
			return addr, nil
		case <-ctx.Done():
			return "", fmt.Errorf("brpc error: no address available %s\n", ctx.Err())
		}
	}
}
