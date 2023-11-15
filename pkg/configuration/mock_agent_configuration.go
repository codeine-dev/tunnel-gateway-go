package configuration

import (
	"errors"
	"net"
)

type MockAgentConfigurationService struct {
	upstreams map[string]net.Addr
}

func NewMockAgentConfigurationService() *MockAgentConfigurationService {
	localhost, _, _ := net.ParseCIDR("127.0.0.1/32")
	return &MockAgentConfigurationService{
		upstreams: map[string]net.Addr{
			"1234": &net.TCPAddr{
				IP:   localhost,
				Port: 8080,
			},
		},
	}
}

func (s *MockAgentConfigurationService) GetUpstreamForConnection( /* need to think which information we get here ... */ ) (net.Addr, error) {
	agent, found := s.upstreams["1234"]
	if !found {
		return nil, errors.New("agent unkown")
	}
	return agent, nil
}
