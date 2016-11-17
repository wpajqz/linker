package linker

import "net"

type response struct {
	net.Conn
	Method uint32
	Params Packet
	Header map[string]string
}

func (r *response) SetResponseProperty(key, value string) {
	r.Header[key] = value
}

func (r *response) GetResponseProperty(key string) string {
	return r.Header[key]
}

func (r *response) DeleteResponseProperty(key string) {
	delete(r.Header, key)
}
