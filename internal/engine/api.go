package engine

import "context"

// API is the Docker surface the command and executor layers depend on.
// The concrete *Client satisfies it; tests use a fake. Keeping the
// dependency an interface is what lets command/verb logic be tested
// without a live Docker daemon.
type API interface {
	Ping(ctx context.Context) error
	Close() error

	ListContainers(ctx context.Context, all bool) ([]Container, error)
	StopContainer(ctx context.Context, id string) error
	RemoveContainer(ctx context.Context, id string, force bool) error
	PruneContainers(ctx context.Context) ([]string, uint64, error)

	ListImages(ctx context.Context) ([]Image, error)
	RemoveImage(ctx context.Context, id string, force bool) error
	PruneImages(ctx context.Context) (uint64, error)

	ListVolumes(ctx context.Context) ([]Volume, error)
	PruneVolumes(ctx context.Context) (uint64, error)

	ListNetworks(ctx context.Context) ([]Network, error)
	PruneNetworks(ctx context.Context) ([]string, error)

	DiskUsage(ctx context.Context) (DiskUsageSummary, error)
	Overview(ctx context.Context) (Overview, error)
}

// Compile-time assertion that the concrete client satisfies the interface.
var _ API = (*Client)(nil)
