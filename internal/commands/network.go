package commands

import (
	"context"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ui"
)

func newNetworkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Manage networks",
	}

	ls := &cobra.Command{
		Use:   "ls",
		Short: "List networks",
		RunE:  func(c *cobra.Command, _ []string) error { return networkList(selectorFrom(c)) },
	}
	clean := &cobra.Command{
		Use:   "clean",
		Short: "Remove unused networks (prune)",
		RunE:  func(_ *cobra.Command, _ []string) error { return networkClean() },
	}

	addSelectorFlags(ls)
	cmd.AddCommand(ls, clean)
	return cmd
}

func networkList(sel engine.Selector) error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
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
	})
}

func networkClean() error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
		if flagDryRun {
			ui.Infof("dry-run: would prune unused networks")
			return nil
		}
		if !flagYes {
			ok, err := prompt("Prune all unused networks?")
			if err != nil || !ok {
				return err
			}
		}
		removed, err := eng.PruneNetworks(ctx)
		if err != nil {
			return err
		}
		ui.Success("removed %d unused network(s)", len(removed))
		return nil
	})
}
