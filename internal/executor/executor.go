// Package executor runs the steps of user-defined custom commands.
//
// This is the only place in the tool that shells out. Because running
// arbitrary commands is a security consideration, it is gated behind the
// global --allow-shell flag (see Options.AllowShell) and is a no-op under
// --dry-run.
package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"dockerToolBox/internal/config"
	"dockerToolBox/internal/ui"
)

// Options controls how steps are executed.
type Options struct {
	// DryRun prints each step without running it.
	DryRun bool
	// AllowShell must be true for steps to actually run.
	AllowShell bool
}

// Run executes the steps of a custom command in order, stopping at the
// first failure.
func Run(ctx context.Context, cmd config.CustomCommand, opts Options) error {
	for i, step := range cmd.Steps {
		line := strings.TrimSpace(step.Command())
		if line == "" {
			continue
		}
		if err := execExternal(ctx, line, opts); err != nil {
			return fmt.Errorf("step %d (%q): %w", i+1, line, err)
		}
	}
	return nil
}

func execExternal(ctx context.Context, line string, opts Options) error {
	if opts.DryRun {
		ui.Step("would run: %s", line)
		return nil
	}
	if !opts.AllowShell {
		return fmt.Errorf("refusing to run external command without --allow-shell: %s", line)
	}
	ui.Infof("running: %s", line)
	c := exec.CommandContext(ctx, "sh", "-c", line)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c.Run()
}
