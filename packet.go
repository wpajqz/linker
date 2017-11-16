package linker

import (
	"github.com/wpajqz/linker/utils/convert"
	"github.com/wpajqz/linker/utils/encrypt"
)

type Packet struct {
	nType         uint32
	nSequence     int64
	nHeaderLength uint32
	nBodyLength   uint32
	bHeader       []byte
	bBody         []byte
}

func NewPack(operator uint32, sequence int64, header, body []byte) Packet {
	var err error

	header, err = encrypt.Encrypt(header)
	if err != nil {
		panic(err)
	}

	body, err = encrypt.Encrypt(body)
	if err != nil {
		panic(err)
	}

	return Packet{
		nType:         operator,
		nSequence:     sequence,
		nHeaderLength: uint32(len(header)),
		nBodyLength:   uint32(len(body)),
		bHeader:       header,
		bBody:         body,
	}
}

// 得到序列化后的Packet
func (p Packet) Bytes() (buf []byte) {
	buf = append(buf, convert.Uint32ToBytes(p.nType)...)
	buf = append(buf, convert.Int64ToBytes(p.nSequence)...)
	buf = append(buf, convert.Uint32ToBytes(p.nHeaderLength)...)
	buf = append(buf, convert.Uint32ToBytes(p.nBodyLength)...)
	buf = append(buf, p.bHeader...)
	buf = append(buf, p.bBody...)

	return buf
}

func (p Packet) OperateType() uint32 {
	return p.nType
}

func (p Packet) Sequence() int64 {
	return p.nSequence
}

func (p Packet) Header() []byte {
	b, err := encrypt.Decrypt(p.bHeader)
	if err != nil {
		panic(err)
	}

	return b
}

func (p Packet) Body() []byte {
	b, err := encrypt.Decrypt(p.bBody)
	if err != nil {
		panic(err)
	}

	return b
}
