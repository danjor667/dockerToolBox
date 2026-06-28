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

	ls := &cobra.Command{
		Use:   "ls",
		Short: "List containers",
		RunE:  func(c *cobra.Command, _ []string) error { return containerList(selectorFrom(c)) },
	}
	stopAll := &cobra.Command{
		Use:   "stop-all",
		Short: "Stop all running containers",
		RunE:  func(c *cobra.Command, _ []string) error { return containerStopAll(selectorFrom(c)) },
	}
	rmAll := &cobra.Command{
		Use:   "rm-all",
		Short: "Remove all containers (running and stopped)",
		RunE:  func(c *cobra.Command, _ []string) error { return containerRemoveAll(selectorFrom(c)) },
	}
	rmStopped := &cobra.Command{
		Use:   "rm-stopped",
		Short: "Remove stopped containers",
		RunE:  func(c *cobra.Command, _ []string) error { return containerRemoveStopped(selectorFrom(c)) },
	}

	for _, sc := range []*cobra.Command{ls, stopAll, rmAll, rmStopped} {
		addSelectorFlags(sc)
	}
	cmd.AddCommand(ls, stopAll, rmAll, rmStopped)
	return cmd
}

func containerList(sel engine.Selector) error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
		cs, err := eng.ListContainers(ctx, true)
		if err != nil {
			return err
		}
		cs = sel.FilterContainers(cs)
		if len(cs) == 0 {
			ui.Infof("no containers match")
			return nil
		}
		rows := make([][]string, 0, len(cs))
		for _, c := range cs {
			rows = append(rows, []string{shortID(c.ID), nameOr(c.Name), c.State, c.Image})
		}
		ui.Table([]string{"ID", "NAME", "STATE", "IMAGE"}, rows)
		return nil
	})
}

func containerStopAll(sel engine.Selector) error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
		cs, err := eng.ListContainers(ctx, false) // running only
		if err != nil {
			return err
		}
		cs = sel.FilterContainers(cs)
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

func containerRemoveAll(sel engine.Selector) error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
		cs, err := eng.ListContainers(ctx, true)
		if err != nil {
			return err
		}
		cs = sel.FilterContainers(cs)
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

func containerRemoveStopped(sel engine.Selector) error {
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
		stopped = sel.FilterContainers(stopped)
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
