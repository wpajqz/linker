package linker

import (
	"net"
	"strings"
)

type request struct {
	net.Conn
	OperateType  uint32
	Sequence     int64
	Header, Body []byte
}

func (r *request) SetRequestProperty(key, value string) {
	v := r.GetRequestProperty(key)
	if v != "" {
		r.Header = []byte(strings.Trim(string(r.Header), key+"="+value+";"))
	}

	r.Header = append(r.Header, []byte(key+"="+value+";")...)
}

func (r *request) GetRequestProperty(key string) string {
	values := strings.Split(string(r.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}
