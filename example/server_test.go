package main

import (
	"testing"

	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/plugins"
)
import (
	"fmt"

	"github.com/wpajqz/go-sdk/export"
)

func TestServer(t *testing.T) {
	client := export.NewClient("127.0.0.1", 8080, &ReadyStateCallback{
		Open: func() {
			fmt.Println("open connection")
		},
		Close: func() {
			fmt.Println("close connection")
		},
		Error: func(err string) {
			fmt.Println(err)
		},
	})

	client.SetPluginForPacketSender([]linker.PacketPlugin{
		&plugins.Encryption{},
	})

	client.SetPluginForPacketReceiver([]linker.PacketPlugin{
		&plugins.Decryption{},
	})

	for {
		if err := client.Ping(nil, RequestStatusCallback{}); err == nil {
			break
		}
	}

	client.SetRequestProperty("sid", "go")

	err := client.SyncSend("/v1/healthy", nil, RequestStatusCallback{
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
}
