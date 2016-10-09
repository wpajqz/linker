package linker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/wpajqz/linker/utils"
)

func (s *Server) handleConnection(conn net.Conn) {
	quit := make(chan bool)

	defer func() {
		if err := recover(); err != nil {
			s.errorHandler(err.(SystemError))
		}

		conn.Close()
		quit <- true
	}()

	receivePackets := make(chan Packet, 100)
	go s.handlePacket(conn, receivePackets, quit)

	var (
		bLen   []byte = make([]byte, 4)
		bType  []byte = make([]byte, 4)
		pacLen uint32
	)

	for {
		conn.SetDeadline(time.Now().Add(s.timeout))

		if n, err := io.ReadFull(conn, bLen); err != nil && n != 4 {
			panic(SystemError{"error", fmt.Errorf("Read packetLength failed: %v", err)})
		}

		if n, err := io.ReadFull(conn, bType); err != nil && n != 4 {
			panic(SystemError{"error", fmt.Errorf("Read packetLength failed: %v", err)})
		}

		if pacLen = utils.BytesToUint32(bLen); pacLen > s.MaxPayload {
			panic(SystemError{"error", errors.New("packet larger than MaxPayload")})
		}

		dataLength := pacLen - 8
		data := make([]byte, dataLength)
		if n, err := io.ReadFull(conn, data); err != nil && n != int(dataLength) {
			panic(SystemError{"error", fmt.Errorf("Read packetLength failed: %v", err)})
		}

		receivePackets <- s.protocolPacket.New(pacLen, utils.BytesToUint32(bType), data)
	}
}

func (s *Server) handlePacket(conn net.Conn, receivePackets <-chan Packet, quit <-chan bool) {
	defer func() {
		if err := recover(); err != nil {
			s.errorHandler(err.(SystemError))
		}
	}()

	for {
		select {
		case p := <-receivePackets:
			handler, ok := s.handlerContainer[p.OperateType()]
			if !ok {
				continue
			}

			go func(handler Handler) {
				defer func() {
					if err := recover(); err != nil {
						s.errorHandler(err.(SystemError))
					}
				}()

				req := &Request{conn, p.OperateType(), p}
				res := Response{conn, p.OperateType(), p}
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
			fmt.Println("stop server running.")
		}
	}
}
