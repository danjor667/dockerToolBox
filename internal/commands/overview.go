package commands

import (
	"context"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ops"
)

func newOverviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "overview",
		Short: "Show a Docker environment overview",
		RunE: func(c *cobra.Command, _ []string) error {
			return withEngine(c.Context(), func(ctx context.Context, eng engine.API) error {
				return ops.Overview(ctx, eng)
			})
		},
	}
}
