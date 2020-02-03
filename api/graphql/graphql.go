package graphql

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/wpajqz/linker/api"
	"github.com/wpajqz/linker/client"
	"github.com/wpajqz/linker/client/export"
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
	app.Any("/rpc", func(ctx *gin.Context) {
		ctx.Set("ctx", ctx)
		ctx.Set("timeout", ja.options.timeout)

		h.ContextHandler(ctx, ctx.Writer, ctx.Request)
	})

	app.GET("/rpc/websocket", func(ctx *gin.Context) {
		var upgrade = websocket.Upgrader{
			HandshakeTimeout:  ja.options.timeout,
			EnableCompression: true,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		conn, err := upgrade.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}

		brpc, err := client.NewClient([]string{address}, ja.options.dialOptions...)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}

		session, err := brpc.Session()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}

		topics := ctx.QueryArray("topic")
		defer func() {
			for _, topic := range topics {
				err := session.RemoveMessageListener(topic)
				if err != nil {
					fmt.Printf("[api] remove message listener error: %s", err.Error())
				}
			}

			_ = conn.Close()
		}()

		for _, topic := range topics {
			err = session.AddMessageListener(topic, export.HandlerFunc(func(header, body []byte) {
				err := conn.WriteMessage(websocket.TextMessage, body)
				if err != nil {
					fmt.Printf("[api] write message error: %s", err.Error())
				}
			}))

			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
				return
			}
		}

		for {
			mt, p, err := conn.ReadMessage()
			if err != nil {
				return
			}

			switch mt {
			case websocket.PingMessage:
				err := conn.WriteMessage(websocket.PongMessage, p)
				if err != nil {
					return
				}
			case websocket.CloseMessage:
				return
			}
		}
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
