package linker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"runtime"
	"time"

	"github.com/wpajqz/linker/plugins"
	"github.com/wpajqz/linker/utils/convert"
)

func (s *Server) handleTcpConnection(ctx context.Context, conn net.Conn) error {
	receivePackets := make(chan Packet, 100)
	go s.handleTcpPacket(ctx, conn, receivePackets)

	var (
		bType         = make([]byte, 4)
		bSequence     = make([]byte, 8)
		bHeaderLength = make([]byte, 4)
		bBodyLength   = make([]byte, 4)
		sequence      int64
		headerLength  uint32
		bodyLength    uint32
	)

	for {
		conn.SetDeadline(time.Now().Add(s.timeout))

		if n, err := io.ReadFull(conn, bType); err != nil && n != 4 {
			return err
		}

		if n, err := io.ReadFull(conn, bSequence); err != nil && n != 8 {
			return err
		}

		if n, err := io.ReadFull(conn, bHeaderLength); err != nil && n != 4 {
			return err
		}

		if n, err := io.ReadFull(conn, bBodyLength); err != nil && n != 4 {
			return err
		}

		sequence = convert.BytesToInt64(bSequence)
		headerLength = convert.BytesToUint32(bHeaderLength)
		bodyLength = convert.BytesToUint32(bBodyLength)
		pacLen := headerLength + bodyLength + uint32(20)

		if pacLen > s.maxPayload {
			_, file, line, _ := runtime.Caller(1)
			return SystemError{time.Now(), file, line, "packet larger than MaxPayload"}
		}

		header := make([]byte, headerLength)
		if n, err := io.ReadFull(conn, header); err != nil && n != int(headerLength) {
			return err
		}

		body := make([]byte, bodyLength)
		if n, err := io.ReadFull(conn, body); err != nil && n != int(bodyLength) {
			return err
		}

		rp, err := NewReceivePack(convert.BytesToUint32(bType), sequence, header, body, []PacketPlugin{
			&plugins.Decryption{},
		})

		if err != nil {
			return err
		}

		receivePackets <- rp
	}
}

func (s *Server) handleTcpPacket(ctx context.Context, conn net.Conn, receivePackets <-chan Packet) {
	var c Context = &ContextTcp{Conn: conn}
	for {
		select {
		case p := <-receivePackets:
			c = NewContextTcp(conn, p.Operator, p.Sequence, s.contentType, p.Header, p.Body)
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
func (s *Server) RunTcp(name, address string) error {
	listener, err := net.Listen(name, address)
	if err != nil {
		return err
	}

	defer listener.Close()

	fmt.Printf("tcp server running on %s\n", address)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go func(conn net.Conn) {
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

				cancel()
				conn.Close()
			}()

			if s.constructHandler != nil {
				s.constructHandler.Handle(nil)
			}

			err := s.handleTcpConnection(ctx, conn)
			if err != nil {
				if s.errorHandler != nil {
					s.errorHandler(err)
				}
			}
		}(conn)
	}
}
