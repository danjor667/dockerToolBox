// Package commands defines the DockerToolBox command tree (Cobra).
//
// Command definitions live here; Docker SDK calls live in package
// engine, and custom-step execution lives in package executor.
package commands

import (
	"context"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/engine"
)

// Global flags, shared by every subcommand via persistent flags.
var (
	flagDryRun     bool
	flagYes        bool
	flagAllowShell bool
)

var rootCmd = &cobra.Command{
	Use:           "dockertoolbox",
	Short:         "DockerToolBox — simplify and automate Docker management",
	Long:          "DockerToolBox turns complex Docker operations into simple, safe commands.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute builds the command tree and runs it.
func Execute() error {
	registerCustomCommands(rootCmd)
	return rootCmd.Execute()
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.BoolVar(&flagDryRun, "dry-run", false, "show what would happen without making changes")
	pf.BoolVarP(&flagYes, "yes", "y", false, "skip confirmation prompts for destructive actions")
	pf.BoolVar(&flagAllowShell, "allow-shell", false, "allow custom commands to run external shell commands")

	rootCmd.AddCommand(
		newOverviewCmd(),
		newContainerCmd(),
		newImageCmd(),
		newVolumeCmd(),
		newSystemCmd(),
	)
}

// withEngine connects to Docker, verifies the daemon is reachable, and
// runs fn with a live client. The client is always closed.
func withEngine(fn func(ctx context.Context, eng *engine.Client) error) error {
	eng, err := engine.New()
	if err != nil {
		return err
	}
	defer eng.Close()

	ctx := context.Background()
	if err := eng.Ping(ctx); err != nil {
		return err
	}
	return fn(ctx, eng)
}
