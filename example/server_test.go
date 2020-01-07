package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/wpajqz/brpc"
	"github.com/wpajqz/brpc/export"
)

func TestServer(t *testing.T) {
	var address = []string{"127.0.0.1:8080"}

	client, err := brpc.NewClient(
		address,
		brpc.WithOnOpen(func() { fmt.Println("open connection") }),
		brpc.WithOnClose(func() { fmt.Println("close connection") }),
		brpc.WithOnError(func(err error) { fmt.Printf("connection error: %s", err.Error()) }),
	)

	if err != nil {
		t.Fatal(err)
	}

	for {
		session, err := client.Session()
		if err != nil {
			fmt.Printf(err.Error())
			continue
		}

		time.Sleep(1 * time.Second)
		err = session.AddMessageListener("/v1/my/message", export.HandlerFunc(func(header, body []byte) {
			fmt.Println(string(header), string(body))
		}))

		if err != nil {
			fmt.Printf(err.Error())
			continue
		}

		go func(session *export.Client) {
			session.SetRequestProperty("sid", "go")
			err = session.SyncSend("/v1/healthy", nil, brpc.RequestStatusCallback{
				Start: func() {
					fmt.Println("start request")
				},
				End: func() {
					fmt.Println("end request")
				},
				Success: func(header, body []byte) {
					fmt.Println(string(body))
				},
				Error: func(code int, message string) {
					fmt.Println(code, message)
				},
			})
			if err != nil {
				t.Error(err)
			}
		}(session)
	}
}
