package commands

import (
	"context"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ui"
)

func newContainerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "container",
		Short: "Manage containers in bulk",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "stop-all",
			Short: "Stop all running containers",
			RunE:  func(_ *cobra.Command, _ []string) error { return containerStopAll() },
		},
		&cobra.Command{
			Use:   "rm-all",
			Short: "Remove all containers (running and stopped)",
			RunE:  func(_ *cobra.Command, _ []string) error { return containerRemoveAll() },
		},
		&cobra.Command{
			Use:   "rm-stopped",
			Short: "Remove stopped containers",
			RunE:  func(_ *cobra.Command, _ []string) error { return containerRemoveStopped() },
		},
	)
	return cmd
}

func containerStopAll() error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
		cs, err := eng.ListContainers(ctx, false) // running only
		if err != nil {
			return err
		}
		ok, err := confirmDestructive("stop", containerLabels(cs))
		if err != nil || !ok {
			return err
		}
		for _, c := range cs {
			if err := eng.StopContainer(ctx, c.ID); err != nil {
				ui.Fail("stop %s: %v", containerLabel(c), err)
				continue
			}
			ui.Success("stopped %s", containerLabel(c))
		}
		return nil
	})
}

func containerRemoveAll() error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
		cs, err := eng.ListContainers(ctx, true)
		if err != nil {
			return err
		}
		ok, err := confirmDestructive("remove", containerLabels(cs))
		if err != nil || !ok {
			return err
		}
		for _, c := range cs {
			if err := eng.RemoveContainer(ctx, c.ID, true); err != nil {
				ui.Fail("remove %s: %v", containerLabel(c), err)
				continue
			}
			ui.Success("removed %s", containerLabel(c))
		}
		return nil
	})
}

func containerRemoveStopped() error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
		all, err := eng.ListContainers(ctx, true)
		if err != nil {
			return err
		}
		var stopped []engine.Container
		for _, c := range all {
			if !c.Running() {
				stopped = append(stopped, c)
			}
		}
		ok, err := confirmDestructive("remove", containerLabels(stopped))
		if err != nil || !ok {
			return err
		}
		for _, c := range stopped {
			if err := eng.RemoveContainer(ctx, c.ID, false); err != nil {
				ui.Fail("remove %s: %v", containerLabel(c), err)
				continue
			}
			ui.Success("removed %s", containerLabel(c))
		}
		return nil
	})
}
