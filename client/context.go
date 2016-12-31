package client

import (
	"net"
	"strings"

	"github.com/wpajqz/linker"
)

type (
	request struct {
		net.Conn
		linker.Packet
	}

	response struct {
		net.Conn
		linker.Packet
	}

	Context struct {
		Request  *request
		Response response
	}
)

func (c *Context) ParseParam(data interface{}) error {
	return c.Response.UnPack(data)
}

func (r *request) SetRequestProperty(key, value string) {
	r.Packet = r.New(r.OperateType(), append(r.Header(), []byte(key+"="+value+";")...), r.Body())
}

func (r response) GetResponseProperty(key string) string {
	values := strings.Split(string(r.Header()), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}

func (r response) UnPack(data interface{}) error {
	return r.Packet.UnPack(data)
}
