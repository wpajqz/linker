package crypt

import "github.com/wpajqz/linker/utils/encrypt"

type Decrypt struct{}

func (d *Decrypt) Handle(header, body []byte) (h, b []byte) {
	h, err := encrypt.Decrypt(header)
	if err != nil {
		return
	}

	b, err = encrypt.Decrypt(body)
	if err != nil {
		return
	}

	return
}

func NewDecryptPlugin() *Decrypt {
	return &Decrypt{}
}
