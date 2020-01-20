package client

import (
	"context"
	"fmt"
	"time"

	"github.com/silenceper/pool"
	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/client/export"
	"github.com/wpajqz/linker/codec"
)

var (
	defaultClient  *Client
	defaultNetwork = linker.NetworkTCP
)

type (
	Client struct {
		options          options
		clientPool       pool.Pool
		address          []string
		availableAddress chan string
	}
)

func NewClient(address []string, opts ...Option) (*Client, error) {
	options := options{
		network:     defaultNetwork,
		contentType: codec.JSON,
		udpPayload:  4096,
		dialTimeout: 3 * time.Second,
		initialCap:  10,
		maxCap:      30,
	}

	for _, o := range opts {
		o(&options)
	}

	defaultClient = &Client{
		options:          options,
		address:          address,
		availableAddress: make(chan string, len(address)),
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
	ctx, _ := context.WithTimeout(context.Background(), c.options.dialTimeout)
	for {
		select {
		case addr := <-c.availableAddress:
			return addr, nil
		case <-ctx.Done():
			return "", fmt.Errorf("brpc error: no address available %s\n", ctx.Err())
		}
	}
}
