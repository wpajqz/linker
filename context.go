package linker

import (
	"bytes"
	"context"
	"strings"

	"github.com/wpajqz/linker/codec"
)

type (
	Context interface {
		WithValue(key interface{}, value interface{}) Context
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

	defaultContext struct {
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

func (dc *defaultContext) WithValue(key interface{}, value interface{}) Context {
	dc.Context = context.WithValue(dc.Context, key, value)
	return dc
}

func (dc *defaultContext) Value(key interface{}) interface{} {
	return dc.Context.Value(key)
}

func (dc *defaultContext) ParseParam(data interface{}) error {
	r, err := codec.NewCoder(dc.config.ContentType)
	if err != nil {
		return err
	}

	return r.Decoder(dc.body, data)
}

func (dc *defaultContext) SetRequestProperty(key, value string) {
	v := dc.GetRequestProperty(key)
	if v != "" {
		dc.Request.Header = bytes.Trim(dc.Request.Header, key+"="+value+";")
	}

	dc.Request.Header = append(dc.Request.Header, []byte(key+"="+value+";")...)
}

func (dc *defaultContext) GetRequestProperty(key string) string {
	values := strings.Split(string(dc.Request.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}

func (dc *defaultContext) SetResponseProperty(key, value string) {
	v := dc.GetResponseProperty(key)
	if v != "" {
		dc.Response.Header = bytes.Trim(dc.Response.Header, key+"="+value+";")
	}

	dc.Response.Header = append(dc.Response.Header, []byte(key+"="+value+";")...)
}

func (dc *defaultContext) GetResponseProperty(key string) string {
	values := strings.Split(string(dc.Response.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}

func (dc *defaultContext) Success(body interface{}) {}

func (dc *defaultContext) Error(code int, message string) {}

func (dc *defaultContext) Write(operator string, body interface{}) (int, error) { return 0, nil }

func (dc *defaultContext) LocalAddr() string { return "" }

func (dc *defaultContext) RemoteAddr() string { return "" }
