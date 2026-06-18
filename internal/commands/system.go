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
	cmd.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: "Stop and remove all containers and images, and prune unused volumes",
		RunE:  func(_ *cobra.Command, _ []string) error { return systemReset() },
	})
	return cmd
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
