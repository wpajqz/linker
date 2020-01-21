package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/wpajqz/linker/api"
	"github.com/wpajqz/linker/client"
	"github.com/wpajqz/linker/client/export"
)

const defaultDialTimeout = 30

type (
	request struct {
		Method string `binding:"required"`
		Param  interface{}
	}

	httpAPI struct {
		options Options
	}
)

func (ha *httpAPI) Dial(network, address string) error {
	ha.options.dialOptions = append(ha.options.dialOptions, client.Network(network))

	brpc, err := client.NewClient([]string{address}, ha.options.dialOptions...)
	if err != nil {
		return err
	}

	gin.SetMode(gin.ReleaseMode)

	app := gin.Default()

	app.POST("/rpc", func(ctx *gin.Context) {
		session, err := brpc.Session()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}

		var req request
		if err := ctx.Bind(&req); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}

		var (
			b           []byte
			errCallback error
		)

		for k, v := range ctx.Request.Header {
			session.SetRequestProperty(k, strings.Join(v, ","))
		}

		to, _ := context.WithTimeout(context.Background(), ha.options.timeout)
		err = session.SyncSendWithTimeout(to, req.Method, req.Param, client.RequestStatusCallback{
			Success: func(header, body []byte) {
				for _, v := range strings.Split(string(header), ";") {
					if len(v) > 0 {
						ss := strings.Split(v, "=")
						if len(ss) > 1 {
							ctx.Writer.Header().Set(ss[0], ss[1])
						}
					}
				}

				b = body
			},
			Error: func(code int, message string) {
				errCallback = errors.New(message)
			},
		})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
			return
		}

		if errCallback != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": errCallback.Error()})
			return
		}

		ctx.Data(http.StatusOK, session.GetContentType(), b)

		return
	})

	app.GET("/rpc/websocket", func(ctx *gin.Context) {
		var upgrade = websocket.Upgrader{
			HandshakeTimeout:  ha.options.timeout,
			EnableCompression: true,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		var quit = make(chan bool)

		conn, err := upgrade.Upgrade(ctx.Writer, ctx.Request, nil)
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
			_ = conn.Close()
			for _, topic := range topics {
				_ = session.RemoveMessageListener(topic)
			}

		}()

		for _, topic := range topics {
			err = session.AddMessageListener(topic, export.HandlerFunc(func(header, body []byte) {
				err := conn.WriteMessage(websocket.TextMessage, body)
				if err != nil {
					quit <- true
				}
			}))

			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
				return
			}
		}

		<-quit
	})

	go app.Run(ha.options.address)

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

	return &httpAPI{options: options}
}
