package ingress

import (
	"context"
	"net"

	"github.com/codeine-dev/go-gateway/pkg/gateway"
	"github.com/codeine-dev/go-gateway/pkg/interfaces"
	"github.com/sirupsen/logrus"
)

type TCPIngressServer struct {
	listen_addr *net.TCPAddr
	running     bool
	stopping    chan bool
}

func NewTCPIngressServer(listen_addr *net.TCPAddr) *TCPIngressServer {
	return &TCPIngressServer{
		listen_addr: listen_addr,
		stopping:    make(chan bool, 1),
		running:     false,
	}
}

func (s *TCPIngressServer) Setup() (*net.TCPListener, error) {
	tcpServer, err := net.ListenTCP("tcp4", s.listen_addr)
	if err != nil {
		return nil, err
	}

	return tcpServer, nil
}

func (s *TCPIngressServer) Status(ctx context.Context) interfaces.ServerStatus {
	if s.running {
		return interfaces.Running
	}
	return interfaces.Stopped
}

func (s *TCPIngressServer) Stop(ctx context.Context) error {
	s.stopping <- true

	return nil
}

type accepted struct {
	conn net.Conn
	err  error
}

func (s *TCPIngressServer) Start(ctx context.Context, handler gateway.AgentHandler) error {
	local_listener, err := s.Setup()
	if err != nil {
		return err
	}

	s.stopping = make(chan bool)

	go func() {
		defer local_listener.Close()
		defer func() {
			s.running = false
		}()
		s.running = true
		for {
			c := make(chan accepted, 1)
			go func() {
				conn, err := local_listener.Accept()
				c <- accepted{conn, err}
			}()
			select {
			case <-s.stopping:
				return
			case a := <-c:
				if a.err != nil {
					logrus.Errorln("Failed to accept QUIC connection", a.err)
				} else {
					handler.NewIngressConnection(a.conn)
				}
			}
		}

	}()

	return nil
}
