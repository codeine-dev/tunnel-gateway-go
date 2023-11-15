package quicclient

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"

	"github.com/codeine-dev/go-gateway/pkg/configuration"
	"github.com/codeine-dev/go-gateway/pkg/utils"
	"github.com/quic-go/quic-go"
	"github.com/sirupsen/logrus"
)

type QuicClient struct {
	server_addr          *net.UDPAddr
	connected            bool
	stopping             chan bool
	configurationService configuration.ConfigurationAgentService
	conn                 quic.Connection
}

func MakeQuicClient(server_addr *net.UDPAddr, configurationService configuration.ConfigurationAgentService) QuicClient {
	return QuicClient{
		server_addr:          server_addr,
		connected:            false,
		stopping:             make(chan bool, 1),
		configurationService: configurationService,
		conn:                 nil,
	}
}

func (s *QuicClient) Connect(ctx context.Context) error {
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: 1235})

	if err != nil {
		return err
	}

	tr := quic.Transport{
		Conn: udpConn,
	}

	s.conn, err = tr.Dial(ctx, s.server_addr, &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}, nil)

	if err != nil {
		return err
	}

	logrus.Info("Connected to QUIC server", s.conn.RemoteAddr())

	control_stream, err := s.conn.OpenStream()

	if err != nil {
		return err
	}

	err = utils.Write(control_stream, 10, []byte("1234"))
	if err != nil {
		return err
	}
	control_stream.Close()

	s.AcceptUpstreams(ctx)

	return nil
}

type accepted struct {
	conn quic.Stream
	err  error
}

func (s *QuicClient) Stop(ctx context.Context) error {
	if s.conn == nil {
		return errors.New("no quic connection yet")
	}
	s.stopping <- true
	s.conn.CloseWithError(0, "ByeBye")
	return nil
}

func (s *QuicClient) AcceptUpstreams(ctx context.Context) error {
	if s.conn == nil {
		return errors.New("no quic connection yet")
	}
	go func() {
		logrus.Infoln("Looking for forwarded connections ...")
		for {
			c := make(chan accepted, 1)
			go func() {
				conn, err := s.conn.AcceptStream(ctx)
				c <- accepted{conn, err}
			}()
			select {
			case <-s.stopping:
				logrus.Errorln("Got shutdown signal")
				return
			case a := <-c:
				if a.err != nil {
					logrus.Errorln("Failed to accept Upstream connection", a.err)
				} else {
					logrus.Infoln("Forwarding connection for stream ", a.conn.StreamID())
					s.ForwardConnection(ctx, a.conn)
				}
			}
		}
	}()

	return nil
}

func (s *QuicClient) ForwardConnection(ctx context.Context, conn quic.Stream) error {
	go func() {
		target, err := s.configurationService.GetUpstreamForConnection()
		if err != nil {
			logrus.Errorln("Failed to find upstream target", err)
			conn.Close()
			return
		}

		logrus.Infof("Trying to connect upstream to %s\n", target.String())
		client, err := net.Dial(target.Network(), target.String())

		if err != nil {
			logrus.Errorln("Failed to open connection to upstream target", err)
			client.Close()
			return
		}

		go func() {
			defer client.Close()
			io.Copy(client, conn)
		}()

		go func() {
			defer conn.Close()
			io.Copy(conn, client)
		}()
	}()

	return nil
}
