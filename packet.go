package linker

type Packet interface {
	Pack(operator uint32, header, body []byte) Packet
	UnPack() []byte
	OperateType() uint32
	Header() []byte
	Body() []byte
}
