package core

import "context"

// Gateway is the common interface for all platform adapters.
type Gateway interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
