package graphql

import (
	"time"

	"github.com/wpajqz/linker/client"
)

type Options struct {
	address     string
	timeout     time.Duration
	graphQL     bool
	pretty      bool
	dialOptions []client.Option
}

type Option func(o *Options)

func Timeout(d time.Duration) Option {
	return func(o *Options) {
		o.timeout = d
	}
}

func GraphQL() Option {
	return func(o *Options) {
		o.graphQL = true
	}
}

func Pretty() Option {
	return func(o *Options) {
		o.pretty = true
	}
}

func DialOptions(opts ...client.Option) Option {
	return func(o *Options) {
		o.dialOptions = opts
	}
}
