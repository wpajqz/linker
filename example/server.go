package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/api/http"
	"github.com/wpajqz/linker/broker/redis"
	"github.com/wpajqz/linker/client"
)

const timeout = 60 * 6 * time.Second

var topic = "/v1/example/subscribe"

func main() {
	server := linker.NewServer(
		linker.Debug(),
		linker.API(
			http.NewAPI(
				"127.0.0.1:9090",
				http.DialOptions(
					client.InitialCapacity(1),
					client.MaxCapacity(1),
				),
			),
		),
		linker.Broker(redis.NewBroker(redis.Address("121.41.20.11:6379"), redis.Password("links471155401"))),
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
