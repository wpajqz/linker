package client

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/silenceper/pool"
	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/client/export"
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
			if c.options.onError != nil {
				c.options.onError(err)
			}
		}

		if c.options.network == linker.NetworkTCP {
			exportClient, err = export.NewClient(address, &ReadyStateCallback{Open: c.options.onOpen, Close: c.options.onClose, Error: func(err string) { c.options.onError(errors.New(err)) }})
		} else {
			exportClient, err = export.NewUDPClient(address, &ReadyStateCallback{Open: c.options.onOpen, Close: c.options.onClose, Error: func(err string) { c.options.onError(errors.New(err)) }})
		}

		if err != nil {
			return nil, fmt.Errorf("brpc error: %s\n", err.Error())
		}

		exportClient.SetUDPPayload(c.options.udpPayload)
		exportClient.SetContentType(c.options.contentType)
		exportClient.SetPluginForPacketSender(c.options.pluginForPacketSender...)
		exportClient.SetPluginForPacketReceiver(c.options.pluginForPacketReceiver...)

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
		InitialCap: c.options.initialCap,
		MaxCap:     c.options.maxCap,
		Factory:    factory,
		Close:      close,
		Ping:       ping,
		//连接最大空闲时间，超过该时间的连接 将会关闭，可避免空闲时连接EOF，自动失效的问题
		IdleTimeout: c.options.idleTimeout,
	}

	return pool.NewChannelPool(pc)
}
