package linker

type Packet interface {
	New(length, operator int32, data []byte) Packet
	OperateType() int32
	Pack(operator int32, data interface{}) (Packet, error)
	UnPack(data interface{}) error
	Bytes() []byte
}
