package executor_test

import (
	"context"
	"strings"
	"testing"

	"dockerToolBox/internal/config"
	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/enginetest"
	"dockerToolBox/internal/executor"
)

// A scalar step naming a native verb runs through the engine, not the shell,
// and needs no --allow-shell.
func TestNativeVerbRunsWithoutShell(t *testing.T) {
	f := &enginetest.Fake{Containers: []engine.Container{{ID: "a", State: "running"}}}
	cmd := config.CustomCommand{Steps: []config.Step{{Raw: "container rm-all"}}}

	err := executor.Run(context.Background(), cmd, executor.Options{
		Yes:    true,
		Engine: func() (engine.API, error) { return f, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Removed) != 1 {
		t.Fatalf("native verb should have removed 1 container, got %v", f.Removed)
	}
}

// An unknown scalar step falls through to the shell and is refused without
// --allow-shell (back-compat with existing configs).
func TestUnknownScalarRequiresAllowShell(t *testing.T) {
	cmd := config.CustomCommand{Steps: []config.Step{{Raw: "echo hi"}}}
	err := executor.Run(context.Background(), cmd, executor.Options{})
	if err == nil || !strings.Contains(err.Error(), "allow-shell") {
		t.Fatalf("want allow-shell refusal, got %v", err)
	}
}

// A dry-run over shell-only steps must not require a Docker engine at all.
func TestShellStepDryRunNeedsNoEngine(t *testing.T) {
	cmd := config.CustomCommand{Steps: []config.Step{{Shell: "echo hi"}}}
	if err := executor.Run(context.Background(), cmd, executor.Options{DryRun: true}); err != nil {
		t.Fatal(err)
	}
}

// A native verb needs the engine; if none is provided, that step errors.
func TestNativeVerbWithoutEngineErrors(t *testing.T) {
	cmd := config.CustomCommand{Steps: []config.Step{{Raw: "container rm-all"}}}
	err := executor.Run(context.Background(), cmd, executor.Options{Yes: true})
	if err == nil || !strings.Contains(err.Error(), "no docker engine") {
		t.Fatalf("want no-engine error, got %v", err)
	}
}

// A "shell:" step is always external, even when its text matches a verb name.
func TestShellFormNeverResolvesToVerb(t *testing.T) {
	cmd := config.CustomCommand{Steps: []config.Step{{Shell: "container rm-all"}}}
	err := executor.Run(context.Background(), cmd, executor.Options{})
	if err == nil || !strings.Contains(err.Error(), "allow-shell") {
		t.Fatalf("shell form should shell out (and be refused), got %v", err)
	}
}

func TestIsVerb(t *testing.T) {
	if !executor.IsVerb("container rm-all") {
		t.Fatal("container rm-all should be a verb")
	}
	if executor.IsVerb("docker compose up -d") {
		t.Fatal("arbitrary shell line should not be a verb")
	}
}
