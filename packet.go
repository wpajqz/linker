package linker

import (
	"fmt"

	"github.com/wpajqz/linker/plugin"
	"github.com/wpajqz/linker/utils/convert"
)

type (
	Packet struct {
		Operator     uint32
		Sequence     int64
		HeaderLength uint32
		BodyLength   uint32
		Header       []byte
		Body         []byte
	}
)

func NewPacket(operator uint32, sequence int64, header, body []byte, plugins []plugin.PacketPlugin) (p Packet, err error) {
	defer func() {
		if r := recover(); r != nil {
			p = Packet{}
			err = fmt.Errorf("[packet error] operator:%d sequence:%d header:%s body:%s detail:%#v", operator, sequence, string(header), string(body), r)
		}
	}()

	for _, plugin := range plugins {
		header, body = plugin.Handle(header, body)
	}

	p = Packet{
		Operator:     operator,
		Sequence:     sequence,
		HeaderLength: uint32(len(header)),
		BodyLength:   uint32(len(body)),
		Header:       header,
		Body:         body,
	}

	return p, err
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
