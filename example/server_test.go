package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/wpajqz/brpc"
	"github.com/wpajqz/brpc/export"
)

func TestServer(t *testing.T) {
	client , err := brpc.NewClient("127.0.0.1", 8080, brpc.Options{
		InitialCap:  5,
		MaxCap:      15,
		IdleTimeout: 0,
		OnOpen: func() {
			fmt.Println("open connection")
		},
		OnClose: func() {
			fmt.Println("close connection")
		},
		OnError: func(e error) {
			fmt.Printf("connection error: %s", e.Error())
		},
	})
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
