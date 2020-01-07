package redis

type (
	Options struct {
		Address  string
		Password string
		PoolSize int
	}

	Option func(o *Options)
)

func Address(address string) Option {
	return func(o *Options) {
		o.Address = address
	}
}

func Password(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

func PoolSize(size int) Option {
	return func(o *Options) {
		o.PoolSize = size
	}
}
