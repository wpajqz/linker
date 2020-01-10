package http

import (
	"github.com/wpajqz/linker/client"
)

type Options struct {
	address string
	client  *client.Client
}

type Option func(o *Options)

func Address(address string) Option {
	return func(o *Options) {
		o.address = address
	}
}

func Client(client *client.Client) Option {
	return func(o *Options) {
		o.client = client
	}
}
