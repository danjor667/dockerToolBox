// Package ops holds the built-in Docker operations: list, bulk stop/remove,
// prune, disk usage, overview, and system reset. Each operation follows the
// safety contract (list -> dry-run gate -> confirm -> act) and talks to
// Docker only through engine.API.
//
// ops is the single implementation shared by two callers: the Cobra command
// tree (package commands) and the native custom-command verbs (package
// executor). Neither reimplements an operation; both call the functions here.
package ops

import (
	"context"
	"fmt"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ui"
)

// Overview prints a high-level snapshot of the Docker environment.
func Overview(ctx context.Context, eng engine.API) error {
	ov, err := eng.Overview(ctx)
	if err != nil {
		return err
	}
	ui.Header("Docker Environment")
	ui.Infof("Server version : %s", ov.ServerVersion)
	ui.Infof("Containers     : %d running / %d total", ov.ContainersRunning, ov.ContainersTotal)
	ui.Infof("Images         : %d", ov.Images)
	ui.Infof("Volumes        : %d", ov.Volumes)
	ui.Infof("Networks       : %d", ov.Networks)
	return nil
}

// ContainerList prints the containers matching sel.
func ContainerList(ctx context.Context, eng engine.API, sel engine.Selector) error {
	cs, err := eng.ListContainers(ctx, true)
	if err != nil {
		return err
	}
	cs = sel.FilterContainers(cs)
	if len(cs) == 0 {
		ui.Infof("no containers match")
		return nil
	}
	rows := make([][]string, 0, len(cs))
	for _, c := range cs {
		rows = append(rows, []string{shortID(c.ID), nameOr(c.Name), c.State, c.Image})
	}
	ui.Table([]string{"ID", "NAME", "STATE", "IMAGE"}, rows)
	return nil
}

// ContainerStopAll stops the running containers matching sel.
func ContainerStopAll(ctx context.Context, eng engine.API, sel engine.Selector, f Flags) error {
	cs, err := eng.ListContainers(ctx, false) // running only
	if err != nil {
		return err
	}
	cs = sel.FilterContainers(cs)
	ok, err := confirmDestructive("stop", containerLabels(cs), f)
	if err != nil || !ok {
		return err
	}
	for _, c := range cs {
		if err := eng.StopContainer(ctx, c.ID); err != nil {
			ui.Fail("stop %s: %v", containerLabel(c), err)
			continue
		}
		ui.Success("stopped %s", containerLabel(c))
	}
	return nil
}

// ContainerRemoveAll removes the containers matching sel (running and stopped).
func ContainerRemoveAll(ctx context.Context, eng engine.API, sel engine.Selector, f Flags) error {
	cs, err := eng.ListContainers(ctx, true)
	if err != nil {
		return err
	}
	cs = sel.FilterContainers(cs)
	ok, err := confirmDestructive("remove", containerLabels(cs), f)
	if err != nil || !ok {
		return err
	}
	for _, c := range cs {
		if err := eng.RemoveContainer(ctx, c.ID, true); err != nil {
			ui.Fail("remove %s: %v", containerLabel(c), err)
			continue
		}
		ui.Success("removed %s", containerLabel(c))
	}
	return nil
}

// ContainerRemoveStopped removes only the stopped containers matching sel.
func ContainerRemoveStopped(ctx context.Context, eng engine.API, sel engine.Selector, f Flags) error {
	all, err := eng.ListContainers(ctx, true)
	if err != nil {
		return err
	}
	var stopped []engine.Container
	for _, c := range all {
		if !c.Running() {
			stopped = append(stopped, c)
		}
	}
	stopped = sel.FilterContainers(stopped)
	ok, err := confirmDestructive("remove", containerLabels(stopped), f)
	if err != nil || !ok {
		return err
	}
	for _, c := range stopped {
		if err := eng.RemoveContainer(ctx, c.ID, false); err != nil {
			ui.Fail("remove %s: %v", containerLabel(c), err)
			continue
		}
		ui.Success("removed %s", containerLabel(c))
	}
	return nil
}

// ImageList prints the images matching sel.
func ImageList(ctx context.Context, eng engine.API, sel engine.Selector) error {
	ims, err := eng.ListImages(ctx)
	if err != nil {
		return err
	}
	ims = sel.FilterImages(ims)
	if len(ims) == 0 {
		ui.Infof("no images match")
		return nil
	}
	rows := make([][]string, 0, len(ims))
	for _, im := range ims {
		rows = append(rows, []string{shortID(im.ID), imageLabel(im), humanSizeI(im.Size)})
	}
	ui.Table([]string{"ID", "TAGS", "SIZE"}, rows)
	return nil
}

// ImageRemoveAll removes the images matching sel.
func ImageRemoveAll(ctx context.Context, eng engine.API, sel engine.Selector, f Flags) error {
	ims, err := eng.ListImages(ctx)
	if err != nil {
		return err
	}
	ims = sel.FilterImages(ims)
	ok, err := confirmDestructive("remove", imageLabels(ims), f)
	if err != nil || !ok {
		return err
	}
	for _, im := range ims {
		if err := eng.RemoveImage(ctx, im.ID, true); err != nil {
			ui.Fail("remove %s: %v", imageLabel(im), err)
			continue
		}
		ui.Success("removed %s", imageLabel(im))
	}
	return nil
}

// ImageClean prunes dangling images.
func ImageClean(ctx context.Context, eng engine.API, f Flags) error {
	ok, err := confirmPrune("would prune dangling images", "Prune all dangling images?", f)
	if err != nil || !ok {
		return err
	}
	reclaimed, err := eng.PruneImages(ctx)
	if err != nil {
		return err
	}
	ui.Success("pruned dangling images, reclaimed %s", humanSize(reclaimed))
	return nil
}

// VolumeList prints the volumes matching sel.
func VolumeList(ctx context.Context, eng engine.API, sel engine.Selector) error {
	vols, err := eng.ListVolumes(ctx)
	if err != nil {
		return err
	}
	vols = sel.FilterVolumes(vols)
	if len(vols) == 0 {
		ui.Infof("no volumes match")
		return nil
	}
	rows := make([][]string, 0, len(vols))
	for _, v := range vols {
		rows = append(rows, []string{v.Name})
	}
	ui.Table([]string{"NAME"}, rows)
	return nil
}

// VolumeClean prunes unused volumes.
func VolumeClean(ctx context.Context, eng engine.API, f Flags) error {
	ok, err := confirmPrune("would prune unused volumes", "Prune all unused volumes?", f)
	if err != nil || !ok {
		return err
	}
	reclaimed, err := eng.PruneVolumes(ctx)
	if err != nil {
		return err
	}
	ui.Success("pruned unused volumes, reclaimed %s", humanSize(reclaimed))
	return nil
}

// NetworkList prints the networks matching sel.
func NetworkList(ctx context.Context, eng engine.API, sel engine.Selector) error {
	nets, err := eng.ListNetworks(ctx)
	if err != nil {
		return err
	}
	nets = sel.FilterNetworks(nets)
	if len(nets) == 0 {
		ui.Infof("no networks match")
		return nil
	}
	rows := make([][]string, 0, len(nets))
	for _, n := range nets {
		rows = append(rows, []string{shortID(n.ID), n.Name, n.Driver, n.Scope})
	}
	ui.Table([]string{"ID", "NAME", "DRIVER", "SCOPE"}, rows)
	return nil
}

// NetworkClean prunes unused networks.
func NetworkClean(ctx context.Context, eng engine.API, f Flags) error {
	ok, err := confirmPrune("would prune unused networks", "Prune all unused networks?", f)
	if err != nil || !ok {
		return err
	}
	removed, err := eng.PruneNetworks(ctx)
	if err != nil {
		return err
	}
	ui.Success("removed %d unused network(s)", len(removed))
	return nil
}

// SystemPrune removes unused data. Unlike SystemReset, it only touches
// resources Docker considers unused (stopped containers, unused networks,
// dangling images, and optionally unused volumes).
func SystemPrune(ctx context.Context, eng engine.API, f Flags, withVolumes bool) error {
	scope := "stopped containers, unused networks, and dangling images"
	if withVolumes {
		scope += ", and unused volumes"
	}
	ok, err := confirmPrune(
		fmt.Sprintf("would prune %s", scope),
		fmt.Sprintf("Prune %s?", scope),
		f,
	)
	if err != nil || !ok {
		return err
	}

	cids, csize, err := eng.PruneContainers(ctx)
	if err != nil {
		return err
	}
	ui.Success("removed %d stopped container(s), reclaimed %s", len(cids), humanSize(csize))

	nets, err := eng.PruneNetworks(ctx)
	if err != nil {
		return err
	}
	ui.Success("removed %d unused network(s)", len(nets))

	isize, err := eng.PruneImages(ctx)
	if err != nil {
		return err
	}
	ui.Success("pruned dangling images, reclaimed %s", humanSize(isize))

	if withVolumes {
		vsize, err := eng.PruneVolumes(ctx)
		if err != nil {
			return err
		}
		ui.Success("pruned unused volumes, reclaimed %s", humanSize(vsize))
	}
	return nil
}

// SystemDF prints how much disk space Docker is using.
func SystemDF(ctx context.Context, eng engine.API) error {
	du, err := eng.DiskUsage(ctx)
	if err != nil {
		return err
	}
	ui.Header("Docker Disk Usage")
	ui.Table(
		[]string{"TYPE", "ITEMS", "SIZE"},
		[][]string{
			{"Images", fmt.Sprint(du.Images), humanSizeI(du.ImagesSize)},
			{"Containers", fmt.Sprint(du.Containers), "-"},
			{"Volumes", fmt.Sprint(du.Volumes), humanSizeI(du.VolumesSize)},
			{"Build cache", fmt.Sprint(du.BuildCache), humanSizeI(du.BuildCacheSize)},
		},
	)
	return nil
}

// SystemReset is the most destructive operation. It summarizes everything
// that will be affected and asks for a single confirmation before doing
// anything.
func SystemReset(ctx context.Context, eng engine.API, f Flags) error {
	containers, err := eng.ListContainers(ctx, true)
	if err != nil {
		return err
	}
	images, err := eng.ListImages(ctx)
	if err != nil {
		return err
	}
	volumes, err := eng.ListVolumes(ctx)
	if err != nil {
		return err
	}

	summary := []string{
		fmt.Sprintf("%d container(s) will be stopped and removed", len(containers)),
		fmt.Sprintf("%d image(s) will be removed", len(images)),
		fmt.Sprintf("%d volume(s) present (unused ones will be pruned)", len(volumes)),
	}
	ok, err := confirmDestructive("reset the Docker environment", summary, f)
	if err != nil || !ok {
		return err
	}

	for _, c := range containers {
		if err := eng.RemoveContainer(ctx, c.ID, true); err != nil {
			ui.Fail("remove container %s: %v", containerLabel(c), err)
			continue
		}
		ui.Success("removed container %s", containerLabel(c))
	}
	for _, im := range images {
		if err := eng.RemoveImage(ctx, im.ID, true); err != nil {
			ui.Fail("remove image %s: %v", imageLabel(im), err)
			continue
		}
		ui.Success("removed image %s", imageLabel(im))
	}
	reclaimed, err := eng.PruneVolumes(ctx)
	if err != nil {
		return err
	}
	ui.Success("pruned unused volumes, reclaimed %s", humanSize(reclaimed))
	return nil
}
