package quicserver

import (
	"context"
	"io"
	"net"

	"github.com/codeine-dev/go-gateway/pkg/utils"
	"github.com/quic-go/quic-go"
	"github.com/sirupsen/logrus"
)

type QuicConnectedAgent struct {
	conn     quic.Connection
	stopping chan bool
	token    chan string
	running  bool
}

type acceptedQuicStream struct {
	conn quic.Stream
	err  error
}

func NewQuicConnectedAgent(conn quic.Connection) QuicConnectedAgent {
	agent := QuicConnectedAgent{
		conn:     conn,
		token:    make(chan string),
		stopping: make(chan bool, 1),
	}

	agent.Start()

	return agent
}

func (q QuicConnectedAgent) Start() {
	// at this point s.listener has to be != nil!!
	go func() {
		defer func() {
			q.running = false
		}()
		defer q.conn.CloseWithError(0, "Connction closed")
		q.running = true
		for {
			c := make(chan acceptedQuicStream, 1)
			go func() {
				conn, err := q.conn.AcceptStream(q.conn.Context())
				c <- acceptedQuicStream{conn, err}
			}()
			select {
			case <-q.stopping:
				logrus.Infoln("Closing agent connection ...")
				return
			case a := <-c:
				if a.err != nil {
					logrus.Errorln("Failed to accept QUIC stream", a.err)
					return
				} else {
					logrus.Infoln("New Stream on QUIC connection from agent")
					tag, buffer, err := utils.Read(a.conn)
					if err != nil && err != io.EOF {
						logrus.Errorln("Failed to read QUIC stream", err)
					} else {
						logrus.Infoln("Got tag: ", tag)
						logrus.Infoln("Got Buffer: ", string(buffer))
						if tag == 10 {
							q.token <- string(buffer)
						}
						logrus.Infoln("Token send to channel")
					}
				}
			}
		}
	}()
}

func (q QuicConnectedAgent) ForwardConnection(conn net.Conn) error {
	upstream, err := q.conn.OpenStream()
	if err != nil {
		return err
	}

	go func() {
		defer upstream.Close()
		io.Copy(upstream, conn)
	}()

	go func() {
		defer conn.Close()
		io.Copy(conn, upstream)
	}()

	return nil
}

func (q QuicConnectedAgent) GetContext() context.Context {
	return q.conn.Context()
}

func (q QuicConnectedAgent) GetRemoteAddr(ctx context.Context) net.Addr {
	return q.conn.RemoteAddr()
}

func (q QuicConnectedAgent) Close(err error) {
	logrus.Infof("Called Close on Agent %s\n", err)
	q.stopping <- true

	reason := "Connection closed"

	if err != nil {
		reason = err.Error()
	}

	q.conn.CloseWithError(0, reason)
}

func (q QuicConnectedAgent) GetTokenChannel() chan string {
	return q.token
}
