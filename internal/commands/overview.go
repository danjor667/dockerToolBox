package commands

import (
	"context"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ui"
)

func newOverviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "overview",
		Short: "Show a Docker environment overview",
		RunE: func(_ *cobra.Command, _ []string) error {
			return withEngine(func(ctx context.Context, eng *engine.Client) error {
				ov, err := eng.Overview(ctx)
				if err != nil {
					return err
				}
				ui.Header("Docker Environment")
				ui.Infof("Server version : %s", ov.ServerVersion)
				ui.Infof("Containers     : %d running / %d total", ov.ContainersRunning, ov.ContainersTotal)
				ui.Infof("Images         : %d", ov.Images)
				ui.Infof("Volumes        : %d", ov.Volumes)
				ui.Infof("Networks       : %d", ov.Networks)
				return nil
			})
		},
	}
}
