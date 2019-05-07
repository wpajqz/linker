package main

import (
	"fmt"
	"log"
	"time"

	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/plugins"
)

const timeout = 60 * 6 * time.Second

func main() {
	server := linker.NewServer(
		linker.Config{
			Timeout: timeout,
			PluginForPacketSender: []linker.PacketPlugin{
				&plugins.Encryption{},
				&plugins.Debug{Sender: true},
			},
			PluginForPacketReceiver: []linker.PacketPlugin{
				&plugins.Decryption{},
				&plugins.Debug{Sender: false},
			},
		})

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
		err := server.RunWebSocket("127.0.0.1:8081")
		if err != nil {
			panic(err)
		}
	}()

	log.Fatal(server.RunTCP("tcp", "127.0.0.1:8080"))
}
