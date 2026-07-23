package commands

import (
	"context"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ops"
)

func newSystemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "System-wide operations",
	}

	reset := &cobra.Command{
		Use:   "reset",
		Short: "Stop and remove all containers and images, and prune unused volumes",
		RunE: func(c *cobra.Command, _ []string) error {
			return withEngine(c.Context(), func(ctx context.Context, eng engine.API) error {
				return ops.SystemReset(ctx, eng, opFlags())
			})
		},
	}

	prune := &cobra.Command{
		Use:   "prune",
		Short: "Remove unused data: stopped containers, unused networks, and dangling images",
		RunE: func(c *cobra.Command, _ []string) error {
			withVolumes, _ := c.Flags().GetBool("volumes")
			return withEngine(c.Context(), func(ctx context.Context, eng engine.API) error {
				return ops.SystemPrune(ctx, eng, opFlags(), withVolumes)
			})
		},
	}
	prune.Flags().Bool("volumes", false, "also prune unused volumes")

	df := &cobra.Command{
		Use:   "df",
		Short: "Show Docker disk usage",
		RunE: func(c *cobra.Command, _ []string) error {
			return withEngine(c.Context(), func(ctx context.Context, eng engine.API) error {
				return ops.SystemDF(ctx, eng)
			})
		},
	}

	cmd.AddCommand(reset, prune, df)
	return cmd
}
