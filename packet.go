package linker

type Packet interface {
	New(length, operator uint32, data []byte) Packet
	OperateType() uint32
	Pack(operator uint32, data interface{}) (Packet, error)
	UnPack(data interface{}) error
	Bytes() []byte
}
