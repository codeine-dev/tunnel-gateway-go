package gateway

import (
	"context"
	"net"
	"time"

	"github.com/codeine-dev/go-gateway/pkg/configuration"
	"github.com/sirupsen/logrus"
)

type ConnectedAgent interface {
	GetContext() context.Context
	Close(err error)
	GetRemoteAddr(ctx context.Context) net.Addr
	GetTokenChannel() chan string
	ForwardConnection(conn net.Conn) error
}

type AgentHandler interface {
	HandleNewAgent(agent ConnectedAgent)
	NewIngressConnection(conn net.Conn)
	CloseAll()
}

func MakeAgentHandler(configurationService configuration.ConfigurationService) AgentHandler {
	return &AgentHandlerImplementation{
		agents:               make(map[string]ConnectedAgent),
		configurationService: configurationService,
	}
}

type AgentHandlerImplementation struct {
	agents               map[string]ConnectedAgent
	configurationService configuration.ConfigurationService
}

func (handler *AgentHandlerImplementation) CloseAll() {
	for _, agent := range handler.agents {
		agent.Close(nil)
	}
}

func (handler *AgentHandlerImplementation) NewIngressConnection(conn net.Conn) {
	go func() {
		logrus.Infof("Handling ingress connection from %s", conn.RemoteAddr().String())
		info, err := handler.configurationService.GetAgentForConnection()
		if err != nil {
			logrus.Errorf("No agent found for target connection %s", err)
			conn.Close()
			return
		}

		agent, found := handler.agents[info.ID]

		if !found {
			logrus.Errorf("Agent %s not online", info.ID)
			conn.Close()
			return
		}

		logrus.Infoln("Piping data .....")

		err = agent.ForwardConnection(conn)

		if err != nil {
			logrus.Errorf("Failed to forward connection %s", err)
			conn.Close()
			return
		}
	}()
}

func (handler *AgentHandlerImplementation) HandleNewAgent(agent ConnectedAgent) {
	go func() {
		ctx := agent.GetContext()
		logrus.Info("New client from ", agent.GetRemoteAddr(ctx))

		tokenChannel := agent.GetTokenChannel()

		token := ""
		select {
		case <-time.After(1 * time.Second):
			return
		case token = <-tokenChannel:
		}

		logrus.Infoln("Got token: ", token)

		info, err := handler.configurationService.GetAgentInfo(token)

		if err != nil {
			agent.Close(err)
			return
		}

		logrus.Infof("Agent: %s connected\n", info.ID)

		handler.agents[info.ID] = agent
		defer func() {
			delete(handler.agents, token)
		}()
	}()
}
