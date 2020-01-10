package http

import (
	"github.com/wpajqz/linker/client"
)

type Options struct {
	address string

	dialOptions []client.Option
}

type Option func(o *Options)

func DialOptions(opts ...client.Option) Option {
	return func(o *Options) {
		o.dialOptions = opts
	}
}
