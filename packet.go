package linker

type Packet interface {
	New(operator uint32, header, body []byte) Packet
	OperateType() uint32
	Pack(operator uint32, header []byte, body interface{}) (Packet, error)
	UnPack(body interface{}) error
	Header() []byte
	Bytes() []byte
}
