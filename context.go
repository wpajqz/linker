package linker

import (
	"bytes"
	"context"
	"strings"

	"github.com/wpajqz/linker/codec"
)

type (
	Context interface {
		WithValue(key interface{}, value interface{})
		Value(key interface{}) interface{}
		ParseParam(data interface{}) error
		Success(body interface{})
		Error(code int, message string)
		Write(operator string, body interface{}) (int, error)
		SetRequestProperty(key, value string)
		GetRequestProperty(key string) string
		SetResponseProperty(key, value string)
		GetResponseProperty(key string) string
		LocalAddr() string
		RemoteAddr() string
	}

	common struct {
		config            Config
		operateType       uint32
		sequence          int64
		body              []byte
		Context           context.Context
		Request, Response struct {
			Header, Body []byte
		}
	}
)

func (dc *common) WithValue(key interface{}, value interface{}) {
	dc.Context = context.WithValue(dc.Context, key, value)
}

func (dc *common) Value(key interface{}) interface{} {
	return dc.Context.Value(key)
}

func (dc *common) ParseParam(data interface{}) error {
	r, err := codec.NewCoder(dc.config.ContentType)
	if err != nil {
		return err
	}

	return r.Decoder(dc.body, data)
}

func (dc *common) SetRequestProperty(key, value string) {
	v := dc.GetRequestProperty(key)
	if v != "" {
		dc.Request.Header = bytes.Trim(dc.Request.Header, key+"="+value+";")
	}

	dc.Request.Header = append(dc.Request.Header, []byte(key+"="+value+";")...)
}

func (dc *common) GetRequestProperty(key string) string {
	values := strings.Split(string(dc.Request.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}

func (dc *common) SetResponseProperty(key, value string) {
	v := dc.GetResponseProperty(key)
	if v != "" {
		dc.Response.Header = bytes.Trim(dc.Response.Header, key+"="+value+";")
	}

	dc.Response.Header = append(dc.Response.Header, []byte(key+"="+value+";")...)
}

func (dc *common) GetResponseProperty(key string) string {
	values := strings.Split(string(dc.Response.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}
