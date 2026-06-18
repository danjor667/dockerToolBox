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
	cmd.AddCommand(
		&cobra.Command{
			Use:   "rm-all",
			Short: "Remove all images",
			RunE:  func(_ *cobra.Command, _ []string) error { return imageRemoveAll() },
		},
		&cobra.Command{
			Use:   "clean",
			Short: "Remove dangling images (prune)",
			RunE:  func(_ *cobra.Command, _ []string) error { return imageClean() },
		},
	)
	return cmd
}

func imageRemoveAll() error {
	return withEngine(func(ctx context.Context, eng *engine.Client) error {
		ims, err := eng.ListImages(ctx)
		if err != nil {
			return err
		}
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
