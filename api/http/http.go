package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wpajqz/linker/api"
	"github.com/wpajqz/linker/client"
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
