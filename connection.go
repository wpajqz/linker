package linker

import (
	"context"
	"fmt"
	"io"
	"net"
	"runtime"
	"time"

	"github.com/wpajqz/linker/utils/convert"
	"github.com/wpajqz/linker/utils/encrypt"
)

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) error {
	receivePackets := make(chan Packet, 100)
	go s.handlePacket(ctx, conn, receivePackets)

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
			if err == io.EOF {
				return err
			}

			_, file, line, _ := runtime.Caller(1)
			return SystemError{time.Now(), file, line, fmt.Sprintf("Read packetLength failed: %v", err)}
		}

		if n, err := io.ReadFull(conn, bSequence); err != nil && n != 8 {
			if err == io.EOF {
				return err
			}

			_, file, line, _ := runtime.Caller(1)
			return SystemError{time.Now(), file, line, fmt.Sprintf("Read packetLength failed: %v", err)}
		}

		if n, err := io.ReadFull(conn, bHeaderLength); err != nil && n != 4 {
			if err == io.EOF {
				return err
			}

			_, file, line, _ := runtime.Caller(1)
			return SystemError{time.Now(), file, line, fmt.Sprintf("Read packetLength failed: %v", err)}
		}

		if n, err := io.ReadFull(conn, bBodyLength); err != nil && n != 4 {
			if err == io.EOF {
				return err
			}

			_, file, line, _ := runtime.Caller(1)
			return SystemError{time.Now(), file, line, fmt.Sprintf("Read packetLength failed: %v", err)}
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

		header, err := encrypt.Decrypt(header)
		if err != nil {
			return err
		}

		body, err = encrypt.Decrypt(body)
		if err != nil {
			return err
		}

		receivePackets <- NewPack(convert.BytesToUint32(bType), sequence, header, body)
	}
}

func (s *Server) handlePacket(ctx context.Context, conn net.Conn, receivePackets <-chan Packet) {
	var c = &Context{}
	for {
		select {
		case p := <-receivePackets:
			handler, ok := s.handlerContainer[p.OperateType()]
			if !ok {
				continue
			}

			c = NewContext(conn, p.OperateType(), p.Sequence(), s.contentType, p.Header(), p.Body())
			go func(handler Handler) {
				defer func() {
					if err := recover(); err != nil {
						s.errorHandler(err.(error))
					}
				}()

				if rm, ok := s.routerMiddleware[p.OperateType()]; ok {
					for _, v := range rm {
						c = v.Handle(c)
					}
				}

				for _, v := range s.middleware {
					c = v.Handle(c)
					if tm, ok := v.(TerminateMiddleware); ok {
						tm.Terminate(c)
					}
				}

				handler(c)
				c.Success(nil) // If it don't call the function of Success or Error, deal it by default
			}(handler)
		case <-ctx.Done():
			// 执行链接退出以后回收操作
			if s.destructHandler != nil {
				s.destructHandler(c)
			}

			return
		}
	}
}
