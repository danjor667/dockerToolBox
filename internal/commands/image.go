package commands

import (
	"context"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ui"
)

func newImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Manage images",
	}

	ls := &cobra.Command{
		Use:   "ls",
		Short: "List images",
		RunE:  func(c *cobra.Command, _ []string) error { return imageList(selectorFrom(c)) },
	}
	rmAll := &cobra.Command{
		Use:   "rm-all",
		Short: "Remove all images",
		RunE:  func(c *cobra.Command, _ []string) error { return imageRemoveAll(selectorFrom(c)) },
	}
	clean := &cobra.Command{
		Use:   "clean",
		Short: "Remove dangling images (prune)",
		RunE:  func(_ *cobra.Command, _ []string) error { return imageClean() },
	}

	addSelectorFlags(ls)
	addSelectorFlags(rmAll)
	cmd.AddCommand(ls, rmAll, clean)
	return cmd
}

func imageList(sel engine.Selector) error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
		ims, err := eng.ListImages(ctx)
		if err != nil {
			return err
		}
		ims = sel.FilterImages(ims)
		if len(ims) == 0 {
			ui.Infof("no images match")
			return nil
		}
		rows := make([][]string, 0, len(ims))
		for _, im := range ims {
			rows = append(rows, []string{shortID(im.ID), imageLabel(im), humanSizeI(im.Size)})
		}
		ui.Table([]string{"ID", "TAGS", "SIZE"}, rows)
		return nil
	})
}

func imageRemoveAll(sel engine.Selector) error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
		ims, err := eng.ListImages(ctx)
		if err != nil {
			return err
		}
		ims = sel.FilterImages(ims)
		ok, err := confirmDestructive("remove", imageLabels(ims))
		if err != nil || !ok {
			return err
		}
		for _, im := range ims {
			if err := eng.RemoveImage(ctx, im.ID, true); err != nil {
				ui.Fail("remove %s: %v", imageLabel(im), err)
				continue
			}
			ui.Success("removed %s", imageLabel(im))
		}
		return nil
	})
}

func imageClean() error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
		if flagDryRun {
			ui.Infof("dry-run: would prune dangling images")
			return nil
		}
		if !flagYes {
			ok, err := prompt("Prune all dangling images?")
			if err != nil || !ok {
				return err
			}
		}
		reclaimed, err := eng.PruneImages(ctx)
		if err != nil {
			return err
		}
		ui.Success("pruned dangling images, reclaimed %s", humanSize(reclaimed))
		return nil
	})
}
