package linker

import "net"

type request struct {
	net.Conn
	Method uint32
	Params Packet
	Header map[string]string
}

func (r *request) SetRequestProperty(key, value string) {
	r.Header[key] = value
}

func (r *request) GetRequestProperty(key string) string {
	return r.Header[key]
}

func (r *request) DeleteRequestProperty(key string) {
	delete(r.Header, key)
}
