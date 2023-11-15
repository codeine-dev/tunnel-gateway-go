package configuration

import "net"

type AgentInfo struct {
	ID string
}

type ConfigurationService interface {
	GetAgentInfo(token string) (*AgentInfo, error)
	GetAgentForConnection( /* need to think which information we get here ... */ ) (*AgentInfo, error)
}

type ConfigurationAgentService interface {
	GetUpstreamForConnection( /* need to think which information we get here ... */ ) (net.Addr, error)
}
