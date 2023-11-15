package quicserver

import (
	"context"
	"net"

	"github.com/codeine-dev/go-gateway/pkg/gateway"
	"github.com/codeine-dev/go-gateway/pkg/interfaces"
	"github.com/codeine-dev/go-gateway/pkg/utils"
	"github.com/quic-go/quic-go"
	"github.com/sirupsen/logrus"
)

type QuicServer struct {
	listen_addr *net.UDPAddr
	listener    *quic.Listener
	stopping    chan bool
	running     bool
}

func MakeQuickServer(listen_addr *net.UDPAddr) QuicServer {
	return QuicServer{
		listen_addr: listen_addr,
		listener:    nil,
		stopping:    make(chan bool, 1),
		running:     false,
	}
}

func (s *QuicServer) Setup() error {
	udpConn, err := net.ListenUDP("udp4", s.listen_addr)
	if err != nil {
		return err
	}

	tr := quic.Transport{
		Conn: udpConn,
	}

	ln, err := tr.Listen(utils.GenerateTLSConfig(), nil)
	if err != nil {
		return err
	}
	s.listener = ln

	return nil
}

func (s *QuicServer) Status(ctx context.Context) interfaces.ServerStatus {
	if s.listener == nil {
		return interfaces.Created
	}

	if s.running {
		return interfaces.Running
	}
	return interfaces.Stopped
}

func (s *QuicServer) Stop(ctx context.Context) error {
	s.stopping <- true

	return nil
}

type accepted struct {
	conn quic.Connection
	err  error
}

func (s *QuicServer) Start(ctx context.Context, handler gateway.AgentHandler) error {
	if s.listener == nil {
		err := s.Setup()
		if err != nil {
			return err
		}
	}

	local_listener := s.listener
	s.stopping = make(chan bool)

	// at this point s.listener has to be != nil!!
	go func() {
		defer local_listener.Close()
		defer func() {
			s.running = false
		}()
		s.running = true
		for {
			c := make(chan accepted, 1)
			go func() {
				conn, err := local_listener.Accept(ctx)
				c <- accepted{conn, err}
			}()
			select {
			case <-s.stopping:
				return
			case a := <-c:
				if a.err != nil {
					logrus.Errorln("Failed to accept QUIC connection", a.err)
				} else {
					//logrus.Infoln("New QUIC connection: ", a.conn.ConnectionState(), a.conn.RemoteAddr())
					//a.conn
					agent := NewQuicConnectedAgent(a.conn)
					handler.HandleNewAgent(agent)
				}
			}
		}

	}()

	return nil
}
