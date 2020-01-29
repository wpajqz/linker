package graphql

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wpajqz/linker/api"
	"github.com/wpajqz/linker/client"
)

var brpc *client.Client

const defaultDialTimeout = 30

type (
	graphqlAPI struct {
		options Options
	}
)

func (ja *graphqlAPI) Dial(network, address string) error {
	ja.options.dialOptions = append(ja.options.dialOptions, client.Network(network))

	var err error
	brpc, err = client.NewClient([]string{address}, ja.options.dialOptions...)
	if err != nil {
		return err
	}

	gin.SetMode(gin.ReleaseMode)

	app := gin.Default()
	app.Any("/graphql", ja.hf())

	go app.Run(ja.options.address)

	return nil
}

func NewAPI(address string, opts ...Option) api.API {
	options := Options{
		address: address,
		timeout: defaultDialTimeout * time.Second,
	}

	for _, o := range opts {
		o(&options)
	}

	return &graphqlAPI{options: options}
}
