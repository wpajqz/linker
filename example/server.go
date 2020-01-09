package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wpajqz/linker"
)

const timeout = 60 * 6 * time.Second

var topic = "/v1/example/subscribe"

func main() {
	server := linker.NewServer(
		linker.Debug(),
		linker.Timeout(timeout),
		linker.WithOnError(linker.HandlerFunc(func(ctx linker.Context) {
			ie := ctx.InternalError()
			if ie != "" {
				ctx.Error(linker.StatusInternalServerError, ctx.InternalError())
			}
		})),
	)

	router := linker.NewRouter()
	router.NSRouter("/v1",
		router.NSRoute(
			"/healthy",
			linker.HandlerFunc(func(ctx linker.Context) {
				err := ctx.Publish(topic, map[string]interface{}{"subscribe": true})
				if err != nil {
					ctx.Error(linker.StatusInternalServerError, err.Error())
				}

				var param map[string]interface{}
				if err := ctx.ParseParam(&param); err != nil {
					ctx.Error(linker.StatusInternalServerError, err.Error())
				}

				ctx.Success(param)
			}),
		),
	)

	server.BindRouter(router)
	go func() {
		r := gin.Default()
		err := server.RunHTTP("127.0.0.1:8081", "/websocket", r)
		if err != nil {
			panic(err)
		}
	}()

	log.Fatal(server.RunTCP("tcp", "127.0.0.1:8080"))
}
