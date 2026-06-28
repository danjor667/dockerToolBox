// Package engine is the only place that talks to the Docker Engine API.
//
// Built-in commands go through this package and never shell out. Running
// user-defined shell steps is the job of package executor.
package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// Client wraps the Docker Engine API client.
type Client struct {
	api *client.Client
}

// New connects to the Docker Engine using the standard environment
// (DOCKER_HOST, etc.) and negotiates the API version with the daemon.
func New() (*Client, error) {
	api, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("creating docker client: %w", err)
	}
	return &Client{api: api}, nil
}

// Close releases the underlying API client.
func (c *Client) Close() error { return c.api.Close() }

// Ping verifies the daemon is reachable.
func (c *Client) Ping(ctx context.Context) error {
	if _, err := c.api.Ping(ctx); err != nil {
		return fmt.Errorf("docker daemon unreachable (is Docker running?): %w", err)
	}
	return nil
}

// Container is a trimmed view of a Docker container.
type Container struct {
	ID     string
	Name   string
	State  string
	Image  string
	Labels map[string]string
}

// Running reports whether the container is currently running.
func (c Container) Running() bool { return c.State == "running" }

// ListContainers returns containers. When all is false, only running
// containers are returned (Docker's default).
func (c *Client) ListContainers(ctx context.Context, all bool) ([]Container, error) {
	list, err := c.api.ContainerList(ctx, container.ListOptions{All: all})
	if err != nil {
		return nil, err
	}
	out := make([]Container, 0, len(list))
	for _, ct := range list {
		name := ""
		if len(ct.Names) > 0 {
			name = strings.TrimPrefix(ct.Names[0], "/")
		}
		out = append(out, Container{ID: ct.ID, Name: name, State: ct.State, Image: ct.Image, Labels: ct.Labels})
	}
	return out, nil
}

// StopContainer stops a running container.
func (c *Client) StopContainer(ctx context.Context, id string) error {
	return c.api.ContainerStop(ctx, id, container.StopOptions{})
}

// RemoveContainer removes a container, optionally forcing removal of a
// running one.
func (c *Client) RemoveContainer(ctx context.Context, id string, force bool) error {
	return c.api.ContainerRemove(ctx, id, container.RemoveOptions{Force: force})
}

// Image is a trimmed view of a Docker image.
type Image struct {
	ID     string
	Tags   []string
	Size   int64
	Labels map[string]string
}

// ListImages returns the top-level images.
func (c *Client) ListImages(ctx context.Context) ([]Image, error) {
	list, err := c.api.ImageList(ctx, image.ListOptions{All: false})
	if err != nil {
		return nil, err
	}
	out := make([]Image, 0, len(list))
	for _, im := range list {
		out = append(out, Image{ID: im.ID, Tags: im.RepoTags, Size: im.Size, Labels: im.Labels})
	}
	return out, nil
}

// RemoveImage removes an image by ID.
func (c *Client) RemoveImage(ctx context.Context, id string, force bool) error {
	_, err := c.api.ImageRemove(ctx, id, image.RemoveOptions{Force: force, PruneChildren: true})
	return err
}

// PruneImages removes dangling images and returns bytes reclaimed.
func (c *Client) PruneImages(ctx context.Context) (uint64, error) {
	report, err := c.api.ImagesPrune(ctx, filters.NewArgs())
	if err != nil {
		return 0, err
	}
	return report.SpaceReclaimed, nil
}

// Volume is a trimmed view of a Docker volume.
type Volume struct {
	Name   string
	Labels map[string]string
}

// ListVolumes returns volumes.
func (c *Client) ListVolumes(ctx context.Context) ([]Volume, error) {
	resp, err := c.api.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]Volume, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		out = append(out, Volume{Name: v.Name, Labels: v.Labels})
	}
	return out, nil
}

// PruneContainers removes stopped containers and returns the removed IDs
// and the number of bytes reclaimed.
func (c *Client) PruneContainers(ctx context.Context) ([]string, uint64, error) {
	report, err := c.api.ContainersPrune(ctx, filters.NewArgs())
	if err != nil {
		return nil, 0, err
	}
	return report.ContainersDeleted, report.SpaceReclaimed, nil
}

// DiskUsageSummary is a high-level breakdown of Docker disk usage.
type DiskUsageSummary struct {
	Images         int
	ImagesSize     int64
	Containers     int
	Volumes        int
	VolumesSize    int64
	BuildCache     int
	BuildCacheSize int64
}

// DiskUsage reports how much disk space Docker is using.
func (c *Client) DiskUsage(ctx context.Context) (DiskUsageSummary, error) {
	du, err := c.api.DiskUsage(ctx, types.DiskUsageOptions{})
	if err != nil {
		return DiskUsageSummary{}, err
	}
	var s DiskUsageSummary
	s.Images = len(du.Images)
	for _, im := range du.Images {
		s.ImagesSize += im.Size
	}
	s.Containers = len(du.Containers)
	s.Volumes = len(du.Volumes)
	for _, v := range du.Volumes {
		if v.UsageData != nil && v.UsageData.Size > 0 {
			s.VolumesSize += v.UsageData.Size
		}
	}
	s.BuildCache = len(du.BuildCache)
	for _, bc := range du.BuildCache {
		s.BuildCacheSize += bc.Size
	}
	return s, nil
}

// PruneVolumes removes unused volumes and returns bytes reclaimed.
func (c *Client) PruneVolumes(ctx context.Context) (uint64, error) {
	report, err := c.api.VolumesPrune(ctx, filters.NewArgs())
	if err != nil {
		return 0, err
	}
	return report.SpaceReclaimed, nil
}

// Overview is a snapshot of the Docker environment.
type Overview struct {
	ServerVersion     string
	ContainersRunning int
	ContainersTotal   int
	Images            int
	Volumes           int
	Networks          int
}

// Overview gathers a high-level summary of the Docker environment.
func (c *Client) Overview(ctx context.Context) (Overview, error) {
	info, err := c.api.Info(ctx)
	if err != nil {
		return Overview{}, err
	}
	all, err := c.ListContainers(ctx, true)
	if err != nil {
		return Overview{}, err
	}
	running := 0
	for _, ct := range all {
		if ct.Running() {
			running++
		}
	}
	imgs, err := c.ListImages(ctx)
	if err != nil {
		return Overview{}, err
	}
	vols, err := c.ListVolumes(ctx)
	if err != nil {
		return Overview{}, err
	}
	nets, err := c.ListNetworks(ctx)
	if err != nil {
		return Overview{}, err
	}
	return Overview{
		ServerVersion:     info.ServerVersion,
		ContainersRunning: running,
		ContainersTotal:   len(all),
		Images:            len(imgs),
		Volumes:           len(vols),
		Networks:          len(nets),
	}, nil
}
