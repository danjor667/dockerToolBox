package commands

import (
	"context"

	"github.com/spf13/cobra"

	"dockerToolBox/internal/config"
	"dockerToolBox/internal/executor"
	"dockerToolBox/internal/ui"
)

// registerCustomCommands loads ~/.docker-toolbox/config.yaml and adds one
// command per user-defined workflow. A missing or invalid config never
// prevents the built-in commands from working.
func registerCustomCommands(root *cobra.Command) {
	cfg, err := config.Load()
	if err != nil {
		ui.Warn("ignoring user config: %v", err)
		return
	}

	builtin := builtinNames(root)
	for name, cc := range cfg.Commands {
		if builtin[name] {
			ui.Warn("custom command %q shadows a built-in and was skipped", name)
			continue
		}
		root.AddCommand(newCustomCmd(name, cc))
	}
}

func newCustomCmd(name string, cc config.CustomCommand) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: cc.Description,
		RunE: func(_ *cobra.Command, _ []string) error {
			ui.Header("Custom command: %s", name)
			return executor.Run(context.Background(), cc, executor.Options{
				DryRun:     flagDryRun,
				AllowShell: flagAllowShell,
			})
		},
	}
}

func builtinNames(root *cobra.Command) map[string]bool {
	names := map[string]bool{}
	for _, c := range root.Commands() {
		names[c.Name()] = true
	}
	return names
}
