package client

import (
	"fmt"
	"net"
	"strings"

	"github.com/golang/protobuf/proto"
)

type (
	request struct {
		net.Conn
		OperateType  uint32
		Header, Body []byte
	}

	response struct {
		net.Conn
		OperateType  uint32
		Header, Body []byte
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
	r.Header = append(r.Header, []byte(key+"="+value+";")...)
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

func (r response) UnPack(data interface{}) error {
	err := proto.Unmarshal(r.Body, data.(proto.Message))
	if err != nil {
		return fmt.Errorf("Unpack error: %v", err.Error())
	}
	return nil
}
