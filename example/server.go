package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/plugins"
)

const timeout = 60 * 6 * time.Second

func main() {
	server := linker.NewServer(
		linker.Debug(),
		linker.Timeout(timeout),
		linker.PluginForPacketSender([]linker.PacketPlugin{
			&plugins.Encryption{},
			&plugins.Debug{Sender: true},
		}),
		linker.PluginForPacketReceiver([]linker.PacketPlugin{
			&plugins.Decryption{},
			&plugins.Debug{Sender: false},
		}),
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
				fmt.Println(ctx.GetRequestProperty("sid"))
				ctx.Success(map[string]interface{}{"keepalive": true})
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
