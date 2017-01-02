package linker

import (
	"net"
	"strings"
)

type request struct {
	net.Conn
	Packet
}

func (r *request) SetRequestProperty(key, value string) {
	r.Packet = NewPack(r.OperateType(), append(r.Header(), []byte(key+"="+value+";")...), r.Body())
}

func (r *request) GetRequestProperty(key string) string {
	values := strings.Split(string(r.Header()), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}
