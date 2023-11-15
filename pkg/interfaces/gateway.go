package interfaces

import "context"

type ServerStatus int

const (
	Created ServerStatus = iota
	Stopped
	Running
)

type GatewayServer interface {
	Status(ctx context.Context) (ServerStatus, error)
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
}
