package crypt

import "github.com/wpajqz/linker/utils/encrypt"

type Encrypt struct{}

func (e *Encrypt) Handle(header, body []byte) (h, b []byte) {
	h, err := encrypt.Encrypt(header)
	if err != nil {
		return
	}

	b, err = encrypt.Encrypt(body)
	if err != nil {
		return
	}

	return
}

func NewEncryptPlugin() *Encrypt {
	return &Encrypt{}
}
