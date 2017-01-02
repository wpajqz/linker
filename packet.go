package linker

type Packet struct {
	Type         uint32
	HeaderLength uint32
	BodyLength   uint32
	bHeader      []byte
	bBody        []byte
}

func NewPack(operator uint32, header, body []byte) Packet {
	return Packet{
		Type:         operator,
		HeaderLength: uint32(len(header)),
		BodyLength:   uint32(len(body)),
		bHeader:      header,
		bBody:        body,
	}
}

// 得到序列化后的Packet
func (p Packet) Bytes() (buf []byte) {
	buf = append(buf, Uint32ToBytes(p.Type)...)
	buf = append(buf, Uint32ToBytes(p.HeaderLength)...)
	buf = append(buf, Uint32ToBytes(p.BodyLength)...)
	buf = append(buf, p.bHeader...)
	buf = append(buf, p.bBody...)

	return buf
}

func (p Packet) OperateType() uint32 {
	return p.Type
}

func (p Packet) Header() []byte {
	return p.bHeader
}

func (p Packet) Body() []byte {
	return p.bBody
}
