package packet

import "github.com/wpajqz/linker"

type ProtoPacket struct {
	Type         uint32
	HeaderLength uint32
	BodyLength   uint32
	bHeader      []byte
	bBody        []byte
}

func (p ProtoPacket) Pack(operator uint32, header, body []byte) linker.Packet {
	return ProtoPacket{
		Type:         operator,
		HeaderLength: uint32(len(header)),
		BodyLength:   uint32(len(body)),
		bHeader:      header,
		bBody:        body,
	}
}

// 得到序列化后的Packet
func (p ProtoPacket) UnPack() (buf []byte) {
	buf = append(buf, linker.Uint32ToBytes(p.Type)...)
	buf = append(buf, linker.Uint32ToBytes(p.HeaderLength)...)
	buf = append(buf, linker.Uint32ToBytes(p.BodyLength)...)
	buf = append(buf, p.bHeader...)
	buf = append(buf, p.bBody...)

	return buf
}

func (p ProtoPacket) OperateType() uint32 {
	return p.Type
}

func (p ProtoPacket) Header() []byte {
	return p.bHeader
}

func (p ProtoPacket) Body() []byte {
	return p.bBody
}
