package linker

import (
	"context"
	"fmt"
	"net"

	uuid "github.com/satori/go.uuid"
	"github.com/wpajqz/linker/utils/convert"
)

func (s *Server) handleUDPData(conn *net.UDPConn, remote *net.UDPAddr, data []byte, length int) {
	bType := data[0:4]
	bSequence := data[4:12]
	bHeaderLength := data[12:16]

	sequence := convert.BytesToInt64(bSequence)
	headerLength := convert.BytesToUint32(bHeaderLength)

	header := data[20 : 20+headerLength]
	body := data[20+headerLength : length]

	rp, err := NewPacket(convert.BytesToUint32(bType), sequence, header, body, s.options.pluginForPacketReceiver)
	if err != nil {
		return
	}

	var ctx Context = NewContextUdp(context.Background(), conn, remote, rp.Operator, rp.Sequence, rp.Header, rp.Body, s.options)

	ctx.Set(nodeID, uuid.NewV4().String())

	defer func() {
		if r := recover(); r != nil {
			var errMsg string

			switch v := r.(type) {
			case string:
				errMsg = v
			case error:
				errMsg = v.Error()
			default:
				errMsg = StatusText(StatusInternalServerError)
			}

			ctx.Set(errorTag, errMsg)

			if s.options.errorHandler != nil {
				s.options.errorHandler.Handle(ctx)
			}

			ctx.Error(StatusInternalServerError, errMsg)
		}

		if err := ctx.UnSubscribeAll(); err != nil {
			ctx.Error(StatusInternalServerError, err.Error())
		}
	}()

	if rp.Operator == OperatorHeartbeat {
		if s.options.pingHandler != nil {
			s.options.pingHandler.Handle(ctx)
		}

		ctx.Success(nil)
	}

	handler, ok := s.router.handlerContainer[rp.Operator]
	if !ok {
		ctx.Error(StatusInternalServerError, "server don't register your request.")
	}

	if rm, ok := s.router.routerMiddleware[rp.Operator]; ok {
		for _, v := range rm {
			ctx = v.Handle(ctx)
		}
	}

	for _, v := range s.router.middleware {
		ctx = v.Handle(ctx)
		if tm, ok := v.(TerminateMiddleware); ok {
			tm.Terminate(ctx)
		}
	}

	handler.Handle(ctx)
	ctx.Success(nil) // If it don't call the function of Success or Error, deal it by default
}

// 开始运行Tcp服务
func (s *Server) runUDP(address string) error {
	udpAddr, err := net.ResolveUDPAddr(NetworkUDP, address)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP(NetworkUDP, udpAddr)
	if err != nil {
		return err
	}

	defer conn.Close()

	fmt.Printf("Listening and serving UDP on %s\n", address)

	if s.options.readBufferSize > 0 {
		err := conn.SetReadBuffer(s.options.readBufferSize)
		if err != nil {
			return err
		}
	}

	if s.options.writeBufferSize > 0 {
		err := conn.SetWriteBuffer(s.options.writeBufferSize)
		if err != nil {
			return err
		}
	}

	if s.options.constructHandler != nil {
		s.options.constructHandler.Handle(nil)
	}

	for {
		var data = make([]byte, s.options.udpPayload)
		n, remote, err := conn.ReadFromUDP(data)
		if err != nil {
			continue
		}

		go s.handleUDPData(conn, remote, data, n)
	}
}
