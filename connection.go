package linker

import (
	"context"
	"hash/crc32"
	"io"
	"log"
	"net"
	"time"

	"github.com/wpajqz/linker/utils"
)

func (s *Server) handleConnection(conn net.Conn) {
	quit := make(chan bool)

	defer func() {
		conn.Close()
		quit <- true
	}()

	receivePackets := make(chan Packet, 100)
	go s.handlePacket(conn, receivePackets, quit)

	// 只有设置有超时的情况下才进行心跳检测进行长连接
	heartbeatPackets := make(chan Packet, 100)
	if s.timeout > 0 {
		go s.checkHeartbeat(conn, heartbeatPackets, quit)
	}

	var (
		bLen   []byte = make([]byte, 4)
		bType  []byte = make([]byte, 4)
		pacLen uint32
	)

	for {
		if n, err := io.ReadFull(conn, bLen); err != nil && n != 4 {
			log.Printf("Read packetLength failed: %v", err)
			return
		}

		if n, err := io.ReadFull(conn, bType); err != nil && n != 4 {
			log.Printf("Read packetType failed: %v", err)
			return
		}

		if pacLen = utils.BytesToUint32(bLen); pacLen > s.MaxPayload {
			log.Printf("packet larger than MaxPayload")
			return
		}

		dataLength := pacLen - 8
		data := make([]byte, dataLength)
		if n, err := io.ReadFull(conn, data); err != nil && n != int(dataLength) {
			log.Printf("Read packetData failed: %v", err)
		}

		// 0号包预留为心跳包使用,其他handler不能够使用
		operator := utils.BytesToUint32(bType)
		if operator == crc32.ChecksumIEEE([]byte("heartbeat")) {
			// 只有设置有超时的情况下才进行心跳检测进行长连接
			if s.timeout > 0 {
				heartbeatPackets <- s.protocolPacket.New(pacLen, utils.BytesToUint32(bType), data)
			}

		} else {
			receivePackets <- s.protocolPacket.New(pacLen, utils.BytesToUint32(bType), data)
		}
	}
}

func (s *Server) handlePacket(conn net.Conn, receivePackets <-chan Packet, quit <-chan bool) {
	for {
		select {
		case p := <-receivePackets:
			handler, ok := s.handlerContainer[p.OperateType()]
			if !ok {
				continue
			}

			go func(handler Handler) {
				req := &Request{conn, p.OperateType(), p}
				res := Response{conn, 200, ""}
				ctx := NewContext(context.Background(), req, res)

				for _, v := range s.middleware {
					ctx = v.Handle(ctx)
				}

				if rm, ok := s.int32Middleware[p.OperateType()]; ok {
					for _, v := range rm {
						ctx = v.Handle(ctx)
					}
				}

				handler(ctx)
			}(handler)

		case <-quit:
			log.Println("Stop handle receivePackets.")
			return
		}
	}
}

func (s *Server) checkHeartbeat(conn net.Conn, heartbeatPackets <-chan Packet, quit <-chan bool) {
	for {
		select {
		case <-heartbeatPackets:
			if s.timeout != 0 {
				conn.SetDeadline(time.Now().Add(s.timeout))
			}
		case <-time.After(s.timeout):
			// todo:添加心跳断开以后的处理逻辑
			conn.Close()
		case <-quit:
			return
		}
	}
}
