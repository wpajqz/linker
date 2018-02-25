package plugins

import (
	"fmt"
	"github.com/wpajqz/linker/utils/encrypt"
)

type Debug struct {
	Sender   bool
	Operator uint32
}

func (d *Debug) Handle(header, body []byte) (h, b []byte) {
	if d.Sender {
		th, _ := encrypt.Decrypt(header)
		tb, _ := encrypt.Decrypt(body)

		fmt.Println("[send packet]", "operator:", d.Operator, "header:", string(th), "body:", string(tb))
	} else {
		fmt.Println("[receive packet]", "operator:", d.Operator, "header:", string(header), "body:", string(body))
	}

	return header, body
}
