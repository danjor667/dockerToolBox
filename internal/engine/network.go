package engine

import (
	"context"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
)

// Network is a trimmed view of a Docker network.
type Network struct {
	ID     string
	Name   string
	Driver string
	Scope  string
	Labels map[string]string
}

// ListNetworks returns all networks.
func (c *Client) ListNetworks(ctx context.Context) ([]Network, error) {
	list, err := c.api.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]Network, 0, len(list))
	for _, n := range list {
		out = append(out, Network{
			ID:     n.ID,
			Name:   n.Name,
			Driver: n.Driver,
			Scope:  n.Scope,
			Labels: n.Labels,
		})
	}
	return out, nil
}

// PruneNetworks removes unused networks and returns the names removed.
func (c *Client) PruneNetworks(ctx context.Context) ([]string, error) {
	report, err := c.api.NetworksPrune(ctx, filters.NewArgs())
	if err != nil {
		return nil, err
	}
	return report.NetworksDeleted, nil
}
