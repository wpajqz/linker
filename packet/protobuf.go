package packet

import (
	"fmt"

	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/utils"

	"github.com/golang/protobuf/proto"
)

type ProtoPacket struct {
	Type         uint32
	HeaderLength uint32
	BodyLength   uint32
	Header       []byte
	Body         []byte
}

// 得到序列化后的Packet
func (p ProtoPacket) Bytes() (buf []byte) {
	buf = append(buf, utils.Uint32ToBytes(p.Type)...)
	buf = append(buf, utils.Uint32ToBytes(p.HeaderLength)...)
	buf = append(buf, utils.Uint32ToBytes(p.BodyLength)...)
	buf = append(buf, p.Header...)
	buf = append(buf, p.Body...)

	return buf
}

// 将数据包类型和pb数据结构一起打包成Packet，并加密Packet.Data
func (p ProtoPacket) Pack(operator uint32, header []byte, body interface{}) (linker.Packet, error) {
	p.Type = operator
	pbData, err := proto.Marshal(body.(proto.Message))
	if err != nil {
		return ProtoPacket{}, fmt.Errorf("Pack error: %v", err.Error())
	}

	p.HeaderLength = uint32(len(header))
	p.Header = header

	// 对Data进行AES加密
	p.Body, err = utils.Encrypt(pbData)
	if err != nil {
		return ProtoPacket{}, fmt.Errorf("Pack error: %v", err.Error())
	}

	p.BodyLength = uint32(len(p.Body))

	return p, nil
}

func (p ProtoPacket) UnPack(pb interface{}) error {
	decryptData, err := utils.Decrypt(p.Body)
	if err != nil {
		return fmt.Errorf("Unpack error: %v", err.Error())
	}

	err = proto.Unmarshal(decryptData, pb.(proto.Message))
	if err != nil {
		return fmt.Errorf("Unpack error: %v", err.Error())
	}

	return nil
}

func (p ProtoPacket) New(operator uint32, header, body []byte) linker.Packet {
	return ProtoPacket{
		Type:         operator,
		HeaderLength: uint32(len(header)),
		BodyLength:   uint32(len(body)),
		Header:       header,
		Body:         body,
	}
}

func (p ProtoPacket) OperateType() uint32 {
	return p.Type
}
