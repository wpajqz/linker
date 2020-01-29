package graphql

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/wpajqz/linker/api"
	"github.com/wpajqz/linker/client"
)

const defaultDialTimeout = 30

var brpc *client.Client

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

	schema, err := graphql.NewSchema(graphql.SchemaConfig{Query: queryType})
	if err != nil {
		return err
	}

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   ja.options.pretty,
		GraphiQL: ja.options.graphQL,
	})

	gin.SetMode(gin.ReleaseMode)

	app := gin.Default()
	app.Any("/graphql", func(ctx *gin.Context) {
		ctx.Set("ctx", ctx)
		ctx.Set("timeout", ja.options.timeout)

		h.ContextHandler(ctx, ctx.Writer, ctx.Request)
	})

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
