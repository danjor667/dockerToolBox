// Package executor runs the steps of user-defined custom commands.
//
// A step resolves one of two ways:
//
//   - a scalar step naming a native verb (e.g. "container rm-all") runs
//     through package ops and the Docker engine, with no shell involved and
//     no --allow-shell required; or
//   - anything else (a "shell:" step, or a scalar that is not a known verb)
//     runs as an external shell command.
//
// Shelling out is a security consideration, so it is gated behind the global
// --allow-shell flag (see Options.AllowShell) and is a no-op under --dry-run.
// This package is the only place in the tool that shells out.
package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"dockerToolBox/internal/config"
	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ops"
	"dockerToolBox/internal/ui"
)

// Options controls how steps are executed.
type Options struct {
	// DryRun prints each step without running it.
	DryRun bool
	// Yes skips confirmation prompts for native destructive verbs.
	Yes bool
	// AllowShell must be true for external shell steps to actually run.
	AllowShell bool
	// Engine lazily connects to Docker. It is called at most once, the first
	// time a step resolves to a native verb, so an all-shell workflow never
	// needs a running daemon. May be nil when no engine is available.
	Engine func() (engine.API, error)
}

// Run executes the steps of a custom command in order, stopping at the
// first failure.
func Run(ctx context.Context, cmd config.CustomCommand, opts Options) error {
	var eng engine.API
	ensureEngine := func() (engine.API, error) {
		if eng != nil {
			return eng, nil
		}
		if opts.Engine == nil {
			return nil, fmt.Errorf("no docker engine available for native verb")
		}
		e, err := opts.Engine()
		if err != nil {
			return nil, err
		}
		if err := e.Ping(ctx); err != nil {
			_ = e.Close()
			return nil, err
		}
		eng = e
		return eng, nil
	}
	defer func() {
		if eng != nil {
			_ = eng.Close()
		}
	}()

	for i, step := range cmd.Steps {
		line := strings.TrimSpace(step.Command())
		if line == "" {
			continue
		}

		// Scalar steps may resolve to a native verb; "shell:" steps never do.
		if !step.IsShell() {
			if v, ok := lookupVerb(line); ok {
				e, err := ensureEngine()
				if err != nil {
					return fmt.Errorf("step %d (%q): %w", i+1, line, err)
				}
				if err := v(ctx, e, ops.Flags{DryRun: opts.DryRun, Yes: opts.Yes}); err != nil {
					return fmt.Errorf("step %d (%q): %w", i+1, line, err)
				}
				continue
			}
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
