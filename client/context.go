package client

import (
	"net"
	"strings"

	"github.com/wpajqz/linker/coder"
)

type Context struct {
	Request  *Request
	Response Response
}

func (c *Context) ParseParam(param interface{}) error {
	t := c.Request.GetRequestProperty("Content-Type")
	r, err := coder.NewCoder(t)
	if err != nil {
		return err
	}

	err = r.Decoder(c.Response.Body, param)
	if err != nil {
		return err
	}

	return nil
}

type Request struct {
	net.Conn
	OperateType  uint32
	Sequence     int64
	Header, Body []byte
}

func (r *Request) SetRequestProperty(key, value string) {
	r.Header = append(r.Header, []byte(key+"="+value+";")...)
}

func (r *Request) GetRequestProperty(key string) string {
	values := strings.Split(string(r.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}

type Response struct {
	net.Conn
	OperateType  uint32
	Sequence     int64
	Header, Body []byte
}

func (r Response) GetResponseProperty(key string) string {
	values := strings.Split(string(r.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}
