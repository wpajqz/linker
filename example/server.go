package main

import (
	"fmt"
	"log"
	"time"

	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/api/graphql"
	"github.com/wpajqz/linker/broker/redis"
	"github.com/wpajqz/linker/client"
)

const timeout = 60 * 6 * time.Second

var topic = "/v1/example/subscribe"

func main() {
	server := linker.NewServer(
		linker.API(
			graphql.NewAPI(
				"127.0.0.1:9090",
				graphql.GraphQL(),
				graphql.Pretty(),
				graphql.DialOptions(
					client.InitialCapacity(1),
					client.MaxCapacity(1),
					client.WithOnError(func(err error) {
						fmt.Println(err.Error())
					}),
				),
			),
		),
		linker.Broker(redis.NewBroker(redis.Address("121.41.20.11:6379"), redis.Password("links471155401"))),
		linker.Timeout(timeout),
		linker.WithHTTPEndpoint(linker.Endpoint{Address: "localhost:8081", WSRoute: "/websocket"}),
		linker.WithUDPEndpoint(linker.Endpoint{Address: "localhost:8082"}),
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

	log.Fatal(server.Run())
}
