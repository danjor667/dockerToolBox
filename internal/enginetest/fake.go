// Package enginetest provides a fake engine.API for tests. It records the
// calls made against it and returns canned data, so command, ops, and
// executor logic can be exercised without a running Docker daemon.
package enginetest

import (
	"context"

	"dockerToolBox/internal/engine"
)

// Fake is an in-memory engine.API. Populate the resource slices to define
// what the daemon "contains"; inspect the recorded slices after a call.
type Fake struct {
	Containers []engine.Container
	Images     []engine.Image
	Volumes    []engine.Volume
	Networks   []engine.Network
	Disk       engine.DiskUsageSummary
	Over       engine.Overview

	// Recorded calls.
	Stopped    []string
	Removed    []string // container IDs
	RemovedImg []string // image IDs
	PrunedCont bool
	PrunedImg  bool
	PrunedVol  bool
	PrunedNet  bool

	// Injected errors.
	PingErr   error
	RemoveErr map[string]error // container id -> error
}

var _ engine.API = (*Fake)(nil)

func (f *Fake) Ping(context.Context) error { return f.PingErr }
func (f *Fake) Close() error               { return nil }

func (f *Fake) ListContainers(_ context.Context, all bool) ([]engine.Container, error) {
	if all {
		return f.Containers, nil
	}
	var out []engine.Container
	for _, c := range f.Containers {
		if c.Running() {
			out = append(out, c)
		}
	}
	return out, nil
}

func (f *Fake) StopContainer(_ context.Context, id string) error {
	f.Stopped = append(f.Stopped, id)
	return nil
}

func (f *Fake) RemoveContainer(_ context.Context, id string, _ bool) error {
	if f.RemoveErr != nil {
		if err := f.RemoveErr[id]; err != nil {
			return err
		}
	}
	f.Removed = append(f.Removed, id)
	return nil
}

func (f *Fake) PruneContainers(context.Context) ([]string, uint64, error) {
	f.PrunedCont = true
	return f.Removed, 0, nil
}

func (f *Fake) ListImages(context.Context) ([]engine.Image, error) { return f.Images, nil }

func (f *Fake) RemoveImage(_ context.Context, id string, _ bool) error {
	f.RemovedImg = append(f.RemovedImg, id)
	return nil
}

func (f *Fake) PruneImages(context.Context) (uint64, error) { f.PrunedImg = true; return 0, nil }

func (f *Fake) ListVolumes(context.Context) ([]engine.Volume, error) { return f.Volumes, nil }
func (f *Fake) PruneVolumes(context.Context) (uint64, error)         { f.PrunedVol = true; return 0, nil }

func (f *Fake) ListNetworks(context.Context) ([]engine.Network, error) { return f.Networks, nil }
func (f *Fake) PruneNetworks(context.Context) ([]string, error)        { f.PrunedNet = true; return nil, nil }

func (f *Fake) DiskUsage(context.Context) (engine.DiskUsageSummary, error) { return f.Disk, nil }
func (f *Fake) Overview(context.Context) (engine.Overview, error)          { return f.Over, nil }
