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

func (ha *httpAPI) Run(debug bool) error {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	app := gin.Default()

	app.POST("/rpc", func(ctx *gin.Context) {
		if ha.options.client == nil {
			ctx.JSON(http.StatusInternalServerError, "api client is nil")
			return
		}

		session, err := ha.options.client.Session()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		var req Request
		if err := ctx.Bind(&req); err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		var b []byte
		err = session.SyncSend(req.Method, req.Param, client.RequestStatusCallback{
			Success: func(header, body []byte) {
				b = body
			},
			Error: func(code int, message string) {
				err = errors.New(message)
			},
		})

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		}

		ctx.Data(http.StatusOK, session.GetContentType(), b)

		return
	})

	return app.Run(ha.options.address)
}

func NewAPI(opts ...Option) api.API {
	options := Options{
		address: "localhost:9090",
	}

	for _, o := range opts {
		o(&options)
	}

	return &httpAPI{options: options}
}
