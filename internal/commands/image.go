package commands

import (
	"context"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ops"
)

func newImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Manage images",
	}

	ls := &cobra.Command{
		Use:   "ls",
		Short: "List images",
		RunE: func(c *cobra.Command, _ []string) error {
			return withEngine(c.Context(), func(ctx context.Context, eng engine.API) error {
				return ops.ImageList(ctx, eng, selectorFrom(c))
			})
		},
	}
	rmAll := &cobra.Command{
		Use:   "rm-all",
		Short: "Remove all images",
		RunE: func(c *cobra.Command, _ []string) error {
			return withEngine(c.Context(), func(ctx context.Context, eng engine.API) error {
				return ops.ImageRemoveAll(ctx, eng, selectorFrom(c), opFlags())
			})
		},
	}
	clean := &cobra.Command{
		Use:   "clean",
		Short: "Remove dangling images (prune)",
		RunE: func(c *cobra.Command, _ []string) error {
			return withEngine(c.Context(), func(ctx context.Context, eng engine.API) error {
				return ops.ImageClean(ctx, eng, opFlags())
			})
		},
	}

	addSelectorFlags(ls)
	addSelectorFlags(rmAll)
	cmd.AddCommand(ls, rmAll, clean)
	return cmd
}
