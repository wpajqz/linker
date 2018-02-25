package linker

import (
	"github.com/wpajqz/linker/utils/convert"
	"github.com/wpajqz/linker/utils/encrypt"
)

type Packet struct {
	Operator     uint32
	Sequence     int64
	HeaderLength uint32
	BodyLength   uint32
	Header       []byte
	Body         []byte
}

func NewSendPack(operator uint32, sequence int64, header, body []byte) (Packet, error) {
	header, err := encrypt.Encrypt(header)
	if err != nil {
		return Packet{}, err
	}

	body, err = encrypt.Encrypt(body)
	if err != nil {
		return Packet{}, err
	}

	return Packet{
		Operator:     operator,
		Sequence:     sequence,
		HeaderLength: uint32(len(header)),
		BodyLength:   uint32(len(body)),
		Header:       header,
		Body:         body,
	}, nil
}

func NewReceivePack(operator uint32, sequence int64, header, body []byte) (Packet, error) {
	header, err := encrypt.Decrypt(header)
	if err != nil {
		return Packet{}, nil
	}

	body, err = encrypt.Decrypt(body)
	if err != nil {
		return Packet{}, nil
	}

	return Packet{
		Operator:     operator,
		Sequence:     sequence,
		HeaderLength: uint32(len(header)),
		BodyLength:   uint32(len(body)),
		Header:       header,
		Body:         body,
	}, nil
}

// 得到序列化后的Packet
func (p Packet) Bytes() (buf []byte) {
	buf = append(buf, convert.Uint32ToBytes(p.Operator)...)
	buf = append(buf, convert.Int64ToBytes(p.Sequence)...)
	buf = append(buf, convert.Uint32ToBytes(p.HeaderLength)...)
	buf = append(buf, convert.Uint32ToBytes(p.BodyLength)...)
	buf = append(buf, p.Header...)
	buf = append(buf, p.Body...)

	return buf
}
