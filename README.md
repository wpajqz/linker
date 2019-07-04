### 包结构 ### 
```go
	Packet struct {
		Operator     uint32 // 帧类型: 4个字节
		Sequence     int64  // 帧序列, 一般用纳秒时间戳标记包发送时间: 8个字节
		HeaderLength uint32 // 头部长度: 4个字节
		BodyLength   uint32 // 内容部分长度: 4个字节
		Header       []byte // 头部
		Body         []byte // 内容
    }
```