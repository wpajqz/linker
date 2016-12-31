package linker

type Packet interface {
	New(operator uint32, header, body []byte) Packet
	Pack(operator uint32, header []byte, body interface{}) (Packet, error)
	UnPack(body interface{}) error
	OperateType() uint32
	Header() []byte
	Body() []byte
	Bytes() []byte
}
