package client

import "github.com/wpajqz/linker"

type (
	request struct {
		Method uint32
		Params linker.Packet
		Header map[string]string
	}

	response struct {
		Method uint32
		Params linker.Packet
		Header map[string]string
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
	r.Header[key] = value
}

func (r response) GetResponseProperty(key string) string {
	return r.Header[key]
}

func (r response) UnPack(data interface{}) error {
	return r.Params.UnPack(data)
}
