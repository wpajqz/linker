package client

import (
	"fmt"
	"net"
	"strings"

	"github.com/golang/protobuf/proto"
)

type Context struct {
	Request  *request
	Response response
}

func (c *Context) ParseParam(data interface{}) error {
	err := proto.Unmarshal(c.Response.Body, data.(proto.Message))
	if err != nil {
		return fmt.Errorf("Unpack error: %v", err.Error())
	}

	return nil
}

type request struct {
	net.Conn
	OperateType  uint32
	Header, Body []byte
}

func (r *request) SetRequestProperty(key, value string) {
	r.Header = append(r.Header, []byte(key+"="+value+";")...)
}

type response struct {
	net.Conn
	OperateType  uint32
	Header, Body []byte
}

func (r response) GetResponseProperty(key string) string {
	values := strings.Split(string(r.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}
