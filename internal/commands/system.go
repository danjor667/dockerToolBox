package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ui"
)

func newSystemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "System-wide operations",
	}

	reset := &cobra.Command{
		Use:   "reset",
		Short: "Stop and remove all containers and images, and prune unused volumes",
		RunE:  func(_ *cobra.Command, _ []string) error { return systemReset() },
	}

	prune := &cobra.Command{
		Use:   "prune",
		Short: "Remove unused data: stopped containers, unused networks, and dangling images",
		RunE: func(c *cobra.Command, _ []string) error {
			withVolumes, _ := c.Flags().GetBool("volumes")
			return systemPrune(withVolumes)
		},
	}
	prune.Flags().Bool("volumes", false, "also prune unused volumes")

	df := &cobra.Command{
		Use:   "df",
		Short: "Show Docker disk usage",
		RunE:  func(_ *cobra.Command, _ []string) error { return systemDF() },
	}

	cmd.AddCommand(reset, prune, df)
	return cmd
}

// systemPrune removes unused data. Unlike reset, it only touches
// resources Docker considers unused (stopped containers, unused networks,
// dangling images, and optionally unused volumes).
func systemPrune(withVolumes bool) error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
		scope := "stopped containers, unused networks, and dangling images"
		if withVolumes {
			scope += ", and unused volumes"
		}
		if flagDryRun {
			ui.Infof("dry-run: would prune %s", scope)
			return nil
		}
		if !flagYes {
			ok, err := prompt(fmt.Sprintf("Prune %s?", scope))
			if err != nil || !ok {
				return err
			}
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
	})
}

func systemDF() error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
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
	})
}

// systemReset is the most destructive command. It summarizes everything
// that will be affected and asks for a single confirmation before doing
// anything.
func systemReset() error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
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
		ok, err := confirmDestructive("reset the Docker environment", summary)
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
	})
}
