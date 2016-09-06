package packet

import (
	"fmt"

	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/utils"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type MsgPacket struct {
	Length uint32
	Type   uint32
	Data   []byte
}

// 得到序列化后的Packet
func (p MsgPacket) Bytes() (buf []byte) {
	buf = append(buf, utils.Uint32ToBytes(p.Length)...)
	buf = append(buf, utils.Uint32ToBytes(p.Type)...)
	buf = append(buf, p.Data...)

	return buf
}

// 将数据包类型和pb数据结构一起打包成Packet，并加密Packet.Data
func (p MsgPacket) Pack(dataType uint32, data interface{}) (linker.Packet, error) {
	pbData, err := msgpack.Marshal(data)
	if err != nil {
		return MsgPacket{}, fmt.Errorf("Pack error: %v", err.Error())
	}

	p.Type = dataType

	// 对Data进行AES加密
	p.Data, err = utils.Encrypt(pbData)
	if err != nil {
		return MsgPacket{}, fmt.Errorf("Pack error: %v", err.Error())
	}

	p.Length = uint32(8 + len(p.Data))

	return p, nil
}

func (p MsgPacket) UnPack(pb interface{}) error {
	decryptData, err := utils.Decrypt(p.Data)
	if err != nil {
		return fmt.Errorf("Unpack error: %v", err.Error())
	}

	err = msgpack.Unmarshal(decryptData, pb)
	if err != nil {
		return fmt.Errorf("Unpack error: %v", err.Error())
	}

	return nil
}

func (p MsgPacket) New(length, operator uint32, data []byte) linker.Packet {
	return MsgPacket{
		Length: length,
		Type:   operator,
		Data:   data,
	}
}

func (p MsgPacket) OperateType() uint32 {
	return p.Type
}
