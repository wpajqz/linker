package linker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/wpajqz/linker/utils/convert"
)

func (s *Server) handleUdpConnection(ctx context.Context, conn *net.UDPConn) error {
	receivePackets := make(chan Packet, 100)
	go s.handleTcpPacket(ctx, conn, receivePackets)

	for {
		conn.SetDeadline(time.Now().Add(s.config.Timeout))

		data := make([]byte, MaxPayload)
		n, _, err := conn.ReadFromUDP(data)
		if err != nil {
			return err
		}

		bType := data[0:4]
		bSequence := data[4:12]
		bHeaderLength := data[12:16]

		sequence := convert.BytesToInt64(bSequence)
		headerLength := convert.BytesToUint32(bHeaderLength)

		header := data[20 : 20+headerLength]
		body := data[20+headerLength : n]

		rp, err := NewPacket(convert.BytesToUint32(bType), sequence, header, body, s.config.PluginForPacketReceiver)
		if err != nil {
			return err
		}

		receivePackets <- rp
	}
}

func (s *Server) handleUdpPacket(ctx context.Context, conn net.Conn, receivePackets <-chan Packet) {
	var c Context = &ContextTcp{Conn: conn}
	for {
		select {
		case p := <-receivePackets:
			c = NewContextTcp(conn, p.Operator, p.Sequence, p.Header, p.Body, s.config)
			if p.Operator == OPERATOR_HEARTBEAT && s.pingHandler != nil {
				go func() {
					s.pingHandler.Handle(c)
					c.Success(nil)
				}()

				continue
			}

			handler, ok := s.router.handlerContainer[p.Operator]
			if !ok {
				continue
			}

			go func(c Context, handler Handler) {
				if rm, ok := s.router.routerMiddleware[p.Operator]; ok {
					for _, v := range rm {
						c = v.Handle(c)
					}
				}

				for _, v := range s.router.middleware {
					c = v.Handle(c)
					if tm, ok := v.(TerminateMiddleware); ok {
						tm.Terminate(c)
					}
				}

				handler.Handle(c)
				c.Success(nil) // If it don't call the function of Success or Error, deal it by default
			}(c, handler)
		case <-ctx.Done():
			// 执行链接退出以后回收操作
			if s.destructHandler != nil {
				s.destructHandler.Handle(c)
			}

			return
		}
	}
}

// 开始运行Tcp服务
func (s *Server) RunUdp(name, address string) error {
	udpAddr, err := net.ResolveUDPAddr(name, address)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP(name, udpAddr)
	if err != nil {
		return err
	}

	fmt.Printf("udp server running on %s\n", address)

	if s.config.ReadBufferSize > 0 {
		conn.SetReadBuffer(s.config.ReadBufferSize)
	}

	if s.config.WriteBufferSize > 0 {
		conn.SetWriteBuffer(s.config.WriteBufferSize)
	}

	if s.constructHandler != nil {
		s.constructHandler.Handle(nil)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if r := recover(); r != nil {
			if s.errorHandler != nil {
				switch v := r.(type) {
				case error:
					s.errorHandler(v)
				case string:
					s.errorHandler(errors.New(v))
				}
			}
		}

		conn.Close()
		cancel()
	}()

	err = s.handleUdpConnection(ctx, conn)
	if err != nil {
		if s.errorHandler != nil {
			s.errorHandler(err)
		}
	}

	return err
}
