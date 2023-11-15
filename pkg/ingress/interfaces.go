package ingress

import (
	"context"

	"github.com/codeine-dev/go-gateway/pkg/interfaces"
)

type IngressServer interface {
	Status(ctx context.Context) (interfaces.ServerStatus, error)
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
}
