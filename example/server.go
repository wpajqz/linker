package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/broker/redis"
	"github.com/wpajqz/linker/plugin/crypt"
)

const timeout = 60 * 6 * time.Second

var topic = "/v1/example/subscribe"

func main() {
	server := linker.NewServer(
		linker.Debug(),
		linker.Broker(redis.NewBroker(redis.Address("127.0.0.1:6379"))),
		linker.Timeout(timeout),
		linker.PluginForPacketSender(crypt.NewEncryptPlugin()),
		linker.PluginForPacketReceiver(crypt.NewDecryptPlugin()),
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
				err := ctx.Publish(topic, map[string]interface{}{"keepalive": true})
				if err != nil {
					ctx.Error(linker.StatusInternalServerError, err.Error())
				}

				ctx.Success(map[string]interface{}{"operator": true})
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
