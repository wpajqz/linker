package linker

import (
	"net"
	"strings"
)

type response struct {
	net.Conn
	OperateType  uint32
	Sequence     int64
	Header, Body []byte
}

func (r *response) SetResponseProperty(key, value string) {
	r.Header = append(r.Header, []byte(key+"="+value+";")...)
}

func (r *response) GetResponseProperty(key string) string {
	values := strings.Split(string(r.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}
