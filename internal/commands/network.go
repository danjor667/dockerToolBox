package commands

import (
	"context"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ops"
)

func newNetworkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Manage networks",
	}

	ls := &cobra.Command{
		Use:   "ls",
		Short: "List networks",
		RunE: func(c *cobra.Command, _ []string) error {
			return withEngine(c.Context(), func(ctx context.Context, eng engine.API) error {
				return ops.NetworkList(ctx, eng, selectorFrom(c))
			})
		},
	}
	clean := &cobra.Command{
		Use:   "clean",
		Short: "Remove unused networks (prune)",
		RunE: func(c *cobra.Command, _ []string) error {
			return withEngine(c.Context(), func(ctx context.Context, eng engine.API) error {
				return ops.NetworkClean(ctx, eng, opFlags())
			})
		},
	}

	addSelectorFlags(ls)
	cmd.AddCommand(ls, clean)
	return cmd
}
