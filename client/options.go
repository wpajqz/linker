package client

import "time"

type (
	options struct {
		network         string
		dialTimeout     time.Duration
		maxPayload      int
		initialCap      int
		maxCap          int
		idleTimeout     time.Duration
		onOpen, onClose func()
		onError         func(error)
	}

	Option func(*options)
)

func WithNetwork(n string) Option {
	return Option(func(o *options) {
		o.network = n
	})
}

func WithDialTimeout(n time.Duration) Option {
	return Option(func(o *options) {
		o.dialTimeout = n
	})
}

func WithMaxPayload(n int) Option {
	return Option(func(o *options) {
		o.maxPayload = n
	})
}

func WithInitialCapacity(n int) Option {
	return Option(func(o *options) {
		o.initialCap = n
	})
}

func WithMaxCapacity(n int) Option {
	return Option(func(o *options) {
		o.maxCap = n
	})
}

func WithIdleTimeout(timeout time.Duration) Option {
	return Option(func(o *options) {
		o.idleTimeout = timeout
	})
}

func WithOnOpen(fn func()) Option {
	return Option(func(o *options) {
		o.onOpen = fn
	})
}

func WithOnClose(fn func()) Option {
	return Option(func(o *options) {
		o.onClose = fn
	})
}

func WithOnError(fn func(err error)) Option {
	return Option(func(o *options) {
		o.onError = fn
	})
}
