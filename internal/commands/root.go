// Package commands defines the DockerToolBox command tree (Cobra).
//
// Command definitions live here; the operations they invoke live in package
// ops, Docker SDK calls in package engine, and custom-step execution in
// package executor. Commands are thin: they parse flags and delegate.
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ops"
)

const (
	version = "0.3.0-dev"
	// pluginName is the sub-command Docker uses to dispatch to this binary
	// when it is installed as a CLI plugin (`docker toolbox …`).
	pluginName = "toolbox"
	// metadataName is the command the Docker CLI calls to discover plugins.
	metadataName = "docker-cli-plugin-metadata"
)

// Global flags, shared by every subcommand via persistent flags.
var (
	flagDryRun     bool
	flagYes        bool
	flagAllowShell bool
)

// newEngine constructs the Docker engine. It is a package variable so tests
// can inject a fake.
var newEngine = func() (engine.API, error) { return engine.New() }

var rootCmd = &cobra.Command{
	Use:           "dockertoolbox",
	Short:         "DockerToolBox — simplify and automate Docker management",
	Long:          "DockerToolBox turns complex Docker operations into simple, safe commands.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute builds the command tree and runs it. It handles both standalone
// invocation and Docker CLI plugin invocation, and installs a signal-aware
// context so in-progress bulk operations cancel cleanly on Ctrl-C.
func Execute() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	args := os.Args[1:]
	// When run as a Docker CLI plugin, the CLI invokes the binary as
	// `docker-toolbox toolbox <args…>`. Strip the injected plugin name so
	// subcommands resolve the same way as in standalone mode.
	if len(args) > 0 && args[0] == pluginName {
		args = args[1:]
	}

	// Skip user-config loading for the metadata probe: its output must be
	// exactly the JSON blob, with no config warnings mixed in.
	if len(args) == 0 || args[0] != metadataName {
		registerCustomCommands(rootCmd)
	}

	rootCmd.SetArgs(args)
	return rootCmd.ExecuteContext(ctx)
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
		newNetworkCmd(),
		newSystemCmd(),
		newPluginMetadataCmd(),
	)
}

// newPluginMetadataCmd implements the Docker CLI plugin metadata protocol.
func newPluginMetadataCmd() *cobra.Command {
	return &cobra.Command{
		Use:    metadataName,
		Hidden: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			b, err := json.Marshal(map[string]string{
				"SchemaVersion":    "0.1.0",
				"Vendor":           "DockerToolBox",
				"Version":          version,
				"ShortDescription": "Simplify and automate Docker management",
			})
			if err != nil {
				return err
			}
			fmt.Println(string(b))
			return nil
		},
	}
}

// opFlags snapshots the global safety flags for an ops call.
func opFlags() ops.Flags { return ops.Flags{DryRun: flagDryRun, Yes: flagYes} }

// withEngine connects to Docker, verifies the daemon is reachable, and runs
// fn with a live client. The client is always closed.
func withEngine(ctx context.Context, fn func(ctx context.Context, eng engine.API) error) error {
	eng, err := newEngine()
	if err != nil {
		return err
	}
	defer eng.Close()

	if err := eng.Ping(ctx); err != nil {
		return err
	}
	return fn(ctx, eng)
}
