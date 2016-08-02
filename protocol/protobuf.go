package protocol

import (
	"fmt"

	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/utils"

	"github.com/golang/protobuf/proto"
)

type ProtoPacket struct {
	Length int32
	Type   int32
	Data   []byte
}

// 得到序列化后的Packet
func (p ProtoPacket) Bytes() (buf []byte) {
	buf = append(buf, utils.Int32ToBytes(p.Length)...)
	buf = append(buf, utils.Int32ToBytes(p.Type)...)
	buf = append(buf, p.Data...)

	return buf
}

// 将数据包类型和pb数据结构一起打包成Packet，并加密Packet.Data
func (p ProtoPacket) Pack(operator int32, pb interface{}) (linker.Packet, error) {
	pbData, err := proto.Marshal(pb.(proto.Message))
	if err != nil {
		return ProtoPacket{}, fmt.Errorf("Pack error: %v", err.Error())
	}

	p.Type = operator

	// 对Data进行AES加密
	p.Data, err = utils.Encrypt(pbData)
	if err != nil {
		return ProtoPacket{}, fmt.Errorf("Pack error: %v", err.Error())
	}

	p.Length = int32(8 + len(p.Data))

	return p, nil
}

func (p ProtoPacket) UnPack(pb interface{}) error {
	decryptData, err := utils.Decrypt(p.Data)
	if err != nil {
		return fmt.Errorf("Unpack error: %v", err.Error())
	}

	err = proto.Unmarshal(decryptData, pb.(proto.Message))
	if err != nil {
		return fmt.Errorf("Unpack error: %v", err.Error())
	}

	return nil
}

func (p ProtoPacket) New(length, operator int32, data []byte) linker.Packet {
	return ProtoPacket{
		Length: length,
		Type:   operator,
		Data:   data,
	}
}

func (p ProtoPacket) OperateType() int32 {
	return p.Type
}
