package main

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:9090/rpc/websocket?topic=/v1/example/subscribe&topic=/v1/example/two", nil)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			time.Sleep(3 * time.Second)
			err := conn.WriteMessage(websocket.PingMessage, []byte("ok"))
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}()

	for {
		t, c, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) {
				fmt.Println("websocket closed")
				return
			}

			fmt.Println(err.Error())
			return
		}

		fmt.Println(t, string(c))
	}
}
