package client

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/silenceper/pool"
	"github.com/wpajqz/linker/client/export"
	"github.com/wpajqz/linker/plugin/crypt"
)

var interval int64 = 60

func (c *Client) newExportPool() (pool.Pool, error) {
	// ping请求的回调，出错的时候调用
	cb := RequestStatusCallback{
		Error: func(code int, msg string) {
			log.Printf("brpc error: %s\n", msg)
		},
	}

	// factory 创建连接的方法
	factory := func() (interface{}, error) {
		var (
			exportClient *export.Client
			err          error
		)

		// 出现错误的时候不退出，继续等待服务端恢复以后进行重连
		address, err := c.getAddress()
		if err != nil {
			c.fillAddress(c.address)
			if c.onError != nil {
				c.onError(err)
			}
		}

		if c.network == "tcp" {
			exportClient, err = export.NewClient(address, &ReadyStateCallback{Open: c.onOpen, Close: c.onClose, Error: func(err string) { c.onError(errors.New(err)) }})
		} else {
			exportClient, err = export.NewUDPClient(address, &ReadyStateCallback{Open: c.onOpen, Close: c.onClose, Error: func(err string) { c.onError(errors.New(err)) }})
		}

		if err != nil {
			return nil, fmt.Errorf("brpc error: %s\n", err.Error())
		}

		exportClient.SetMaxPayload(c.maxPayload)
		exportClient.SetPluginForPacketSender(crypt.NewEncryptPlugin())
		exportClient.SetPluginForPacketReceiver(crypt.NewDecryptPlugin())

		go func(ec *export.Client) {
			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			for {
				select {
				case <-ticker.C:
					err := ec.Ping(nil, cb)
					if err != nil {
						return
					}
				}
			}
		}(exportClient)

		c.availableAddress <- address

		return exportClient, nil
	}

	close := func(v interface{}) error {
		return v.(*export.Client).Close()
	}

	ping := func(v interface{}) error {
		return v.(*export.Client).Ping(nil, cb)
	}

	pc := &pool.Config{
		InitialCap: c.initialCap,
		MaxCap:     c.maxCap,
		Factory:    factory,
		Close:      close,
		Ping:       ping,
		//连接最大空闲时间，超过该时间的连接 将会关闭，可避免空闲时连接EOF，自动失效的问题
		IdleTimeout: c.idleTimeout,
	}

	return pool.NewChannelPool(pc)
}
