package commands

import (
	"context"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ops"
)

func newVolumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume",
		Short: "Manage volumes",
	}

	ls := &cobra.Command{
		Use:   "ls",
		Short: "List volumes",
		RunE: func(c *cobra.Command, _ []string) error {
			return withEngine(c.Context(), func(ctx context.Context, eng engine.API) error {
				return ops.VolumeList(ctx, eng, selectorFrom(c))
			})
		},
	}
	clean := &cobra.Command{
		Use:   "clean",
		Short: "Remove unused volumes (prune)",
		RunE: func(c *cobra.Command, _ []string) error {
			return withEngine(c.Context(), func(ctx context.Context, eng engine.API) error {
				return ops.VolumeClean(ctx, eng, opFlags())
			})
		},
	}

	addSelectorFlags(ls)
	cmd.AddCommand(ls, clean)
	return cmd
}
