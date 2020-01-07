package redis

type (
	Options struct {
		Address  string
		Password string
		DB       int
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

func DB(db int) Option {
	return func(o *Options) {
		o.DB = db
	}
}
