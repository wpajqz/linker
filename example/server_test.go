package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/wpajqz/linker/client"
	"github.com/wpajqz/linker/client/export"
)

func TestServer(t *testing.T) {
	var (
		address = []string{"127.0.0.1:8080"}
		cc      *client.Client
		err     error
	)

	cc, err = client.NewClient(
		address,
		client.WithOnOpen(func() {
			session, err := client.Session()
			if err != nil {
				panic(err)
			}

			err = session.AddMessageListener(topic, export.HandlerFunc(func(header, body []byte) {
				fmt.Println("topic: ", string(body))
			}))

			if err != nil {
				fmt.Println("topic", err.Error())
			}
		}),
		client.WithInitialCapacity(1),
		client.WithMaxCapacity(1),
		client.WithOnClose(func() { fmt.Println("close connection") }),
		client.WithOnError(func(err error) { fmt.Printf("connection error: %s", err.Error()) }),
	)

	if err != nil {
		t.Fatal(err)
	}

	for {
		session, err := cc.Session()
		if err != nil {
			fmt.Printf(err.Error())
			continue
		}

		time.Sleep(2 * time.Second)

		session.SetRequestProperty("sid", "go")
		err = session.SyncSend("/v1/healthy", nil, client.RequestStatusCallback{
			Success: func(header, body []byte) {
				fmt.Println("operator", string(body))
			},
			Error: func(code int, message string) {
				fmt.Println("operator", code, message)
			},
		})
		if err != nil {
			t.Error(err)
		}
	}
}
