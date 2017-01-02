package linker

import (
	"net"
	"strings"
)

type response struct {
	net.Conn
	Packet
}

func (r *response) SetResponseProperty(key, value string) {
	r.Packet = NewPack(r.OperateType(), append(r.Header(), []byte(key+"="+value+";")...), r.Body())
}

func (r *response) GetResponseProperty(key string) string {
	values := strings.Split(string(r.Header()), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}
