package commands

import (
	"context"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ui"
)

func newVolumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume",
		Short: "Manage volumes",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "clean",
		Short: "Remove unused volumes (prune)",
		RunE:  func(_ *cobra.Command, _ []string) error { return volumeClean() },
	})
	return cmd
}

func volumeClean() error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
		if flagDryRun {
			ui.Infof("dry-run: would prune unused volumes")
			return nil
		}
		if !flagYes {
			ok, err := prompt("Prune all unused volumes?")
			if err != nil || !ok {
				return err
			}
		}
		reclaimed, err := eng.PruneVolumes(ctx)
		if err != nil {
			return err
		}
		ui.Success("pruned unused volumes, reclaimed %s", humanSize(reclaimed))
		return nil
	})
}
