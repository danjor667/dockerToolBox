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

	ls := &cobra.Command{
		Use:   "ls",
		Short: "List volumes",
		RunE:  func(c *cobra.Command, _ []string) error { return volumeList(selectorFrom(c)) },
	}
	clean := &cobra.Command{
		Use:   "clean",
		Short: "Remove unused volumes (prune)",
		RunE:  func(_ *cobra.Command, _ []string) error { return volumeClean() },
	}

	addSelectorFlags(ls)
	cmd.AddCommand(ls, clean)
	return cmd
}

func volumeList(sel engine.Selector) error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
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
	})
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
