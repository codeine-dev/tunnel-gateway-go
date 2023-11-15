package configuration

import "errors"

type MockConfigurationService struct {
	agents map[string]AgentInfo
}

func NewMockConfigurationService() *MockConfigurationService {
	return &MockConfigurationService{
		agents: map[string]AgentInfo{
			"1234": {
				ID: "a0d2c4d4-3944-416f-9856-45ca1940534b",
			},
			"5678": {
				ID: "5c14f9da-ee1f-4bcd-bd3f-0a2e14a7c65c",
			},
		},
	}
}

func (s *MockConfigurationService) GetAgentForConnection( /* need to think which information we get here ... */ ) (*AgentInfo, error) {
	agent, found := s.agents["1234"]
	if !found {
		return nil, errors.New("agent unkown")
	}
	return &agent, nil
}

func (s *MockConfigurationService) GetAgentInfo(token string) (*AgentInfo, error) {
	agent, found := s.agents[token]
	if !found {
		return nil, errors.New("agent unkown")
	}

	return &agent, nil
}
