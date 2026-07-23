package commands

import (
	"context"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ops"
)

func newContainerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "container",
		Short: "Manage containers in bulk",
	}

	ls := &cobra.Command{
		Use:   "ls",
		Short: "List containers",
		RunE: func(c *cobra.Command, _ []string) error {
			return withEngine(c.Context(), func(ctx context.Context, eng engine.API) error {
				return ops.ContainerList(ctx, eng, selectorFrom(c))
			})
		},
	}
	stopAll := &cobra.Command{
		Use:   "stop-all",
		Short: "Stop all running containers",
		RunE: func(c *cobra.Command, _ []string) error {
			return withEngine(c.Context(), func(ctx context.Context, eng engine.API) error {
				return ops.ContainerStopAll(ctx, eng, selectorFrom(c), opFlags())
			})
		},
	}
	rmAll := &cobra.Command{
		Use:   "rm-all",
		Short: "Remove all containers (running and stopped)",
		RunE: func(c *cobra.Command, _ []string) error {
			return withEngine(c.Context(), func(ctx context.Context, eng engine.API) error {
				return ops.ContainerRemoveAll(ctx, eng, selectorFrom(c), opFlags())
			})
		},
	}
	rmStopped := &cobra.Command{
		Use:   "rm-stopped",
		Short: "Remove stopped containers",
		RunE: func(c *cobra.Command, _ []string) error {
			return withEngine(c.Context(), func(ctx context.Context, eng engine.API) error {
				return ops.ContainerRemoveStopped(ctx, eng, selectorFrom(c), opFlags())
			})
		},
	}

	for _, sc := range []*cobra.Command{ls, stopAll, rmAll, rmStopped} {
		addSelectorFlags(sc)
	}
	cmd.AddCommand(ls, stopAll, rmAll, rmStopped)
	return cmd
}
