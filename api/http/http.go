package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wpajqz/linker/api"
	"github.com/wpajqz/linker/client"
)

type (
	Request struct {
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
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		var req Request
		if err := ctx.Bind(&req); err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		var (
			b           []byte
			errCallback error
		)
		err = session.SyncSend(req.Method, req.Param, client.RequestStatusCallback{
			Success: func(header, body []byte) {
				b = body
			},
			Error: func(code int, message string) {
				errCallback = errors.New(message)
			},
		})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
		}

		if errCallback != nil {
			ctx.JSON(http.StatusInternalServerError, errCallback.Error())
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
	}

	for _, o := range opts {
		o(&options)
	}

	return &httpAPI{options: options}
}
