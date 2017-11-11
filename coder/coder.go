package coder

import "errors"

const (
	JSON     = "text/json"
	PROTOBUF = "text/protobuf"
)

var coderContainer = make(map[string]Coder)

type Coder interface {
	Encoder(data interface{}) ([]byte, error)
	Decoder(data []byte, v interface{}) error
}

func NewCoder(name string) (Coder, error) {
	if v, ok := coderContainer[name]; ok {
		return v, nil
	}

	return nil, errors.New("unsupported data type")
}

func register(name string, coder Coder) {
	if _, ok := coderContainer["name"]; !ok {
		coderContainer[name] = coder
	}
}
