# DockerToolBox v0.3 — Implementation Plan

**Branch:** `v0.3-plan`
**Baseline:** v0.2 (`master` @ `4904cd9`)
**Theme:** Make custom commands run *through the engine* (no shell), ship the
Docker CLI plugin entry point, and pay down the testability debt that currently
blocks all of the above.

The organizing principle stays the one the project already commits to: **only
`internal/engine` talks to Docker; only `internal/executor` shells out.** Every
change below reinforces that boundary rather than blurring it.

---

## Goals

1. **SDK-native custom-command verbs** — a scalar step can name a built-in verb
   (`container rm-all`, `image clean`, …) that runs through `engine` with no
   shell and no `--allow-shell`. This is the promised payoff of the
   scalar-vs-`shell:` distinction that already exists in `config.Step` but today
   does nothing.
2. **Docker CLI plugin entry point** — invocable as `docker toolbox …` in
   addition to the standalone binary.
3. **Testability** — introduce an engine interface so command/executor logic can
   be tested against a fake, without a live Docker daemon.
4. **Variables & hooks** (stretch) — config-level variable substitution and
   `pre`/`post` hooks for custom commands.

Non-goals for v0.3: the interactive TUI (stays v0.4), remote/multi-host support,
and any change to the built-in command surface's behavior.

---

## Phase 0 — Testability foundation (do this first)

Everything else is easier and safer once command logic can be tested against a
fake engine. Today only `engine/selector_test.go` exists because the selector is
pure; `commands/`, `executor`, `config`, and `safety.go` are untested because
they reach the concrete `*engine.Client`.

### 0.1 Extract an engine interface

Define the surface the command layer actually uses. Place it in
`internal/engine` (consumer-side interface is also fine, but a shared one keeps
the fake in one place).

```go
// internal/engine/api.go
package engine

import "context"

// API is the Docker surface the command layer depends on. The concrete
// *Client satisfies it; tests use a fake.
type API interface {
    Ping(ctx context.Context) error
    Close() error

    ListContainers(ctx context.Context, all bool) ([]Container, error)
    StopContainer(ctx context.Context, id string) error
    RemoveContainer(ctx context.Context, id string, force bool) error
    PruneContainers(ctx context.Context) ([]string, uint64, error)

    ListImages(ctx context.Context) ([]Image, error)
    RemoveImage(ctx context.Context, id string, force bool) error
    PruneImages(ctx context.Context) (uint64, error)

    ListVolumes(ctx context.Context) ([]Volume, error)
    PruneVolumes(ctx context.Context) (uint64, error)

    ListNetworks(ctx context.Context) ([]Network, error)
    PruneNetworks(ctx context.Context) ([]string, error)

    DiskUsage(ctx context.Context) (DiskUsageSummary, error)
    Overview(ctx context.Context) (Overview, error)
}

var _ API = (*Client)(nil)
```

### 0.2 Thread the interface through `commands`

- Change `withEngine` in `commands/root.go` to hand callers an `engine.API`
  instead of a concrete `*engine.Client`.
- Add a package-level seam so tests can inject a fake:

  ```go
  // newEngine is overridable in tests.
  var newEngine = func() (engine.API, error) { return engine.New() }
  ```

  `withEngine` calls `newEngine()` instead of `engine.New()` directly.
- Every `func(ctx, eng *engine.Client)` closure becomes `func(ctx, eng engine.API)`.
  Mechanical, no logic change.

### 0.3 Fake engine + first command tests

- Add `internal/engine/fake.go` (build tag or plain file) or an
  `internal/enginetest` package with a `Fake` that records calls and returns
  canned data.
- New tests in `internal/commands`:
  - `confirmDestructive`: no-targets → false; `--dry-run` → false; `--yes` →
    true; interactive path with piped stdin.
  - `containerRemoveStopped`: only non-running containers are targeted.
  - Selector application end-to-end through a command (fake returns a mixed set,
    assert the filtered subset is what gets removed).
  - Error-continue behavior: one `RemoveContainer` fails, the loop still
    processes the rest.

**Acceptance:** `go test ./...` covers `commands` and `executor`, not just the
selector. No live daemon required.

---

## Phase 1 — SDK-native custom-command verbs

The heart of v0.3. Today `executor.Run` treats scalar and `shell:` steps
identically — both go through `sh -c` and need `--allow-shell`. After this
phase, a scalar step that names a known verb runs natively.

### 1.1 Define the verb registry

A verb maps a name to an engine-backed action that honors the existing safety
flags (`--dry-run`, `--yes`).

```go
// internal/executor/verbs.go (or a small internal/verbs package)
type Verb func(ctx context.Context, eng engine.API, opts Options) error

var verbs = map[string]Verb{
    "container stop-all":   ...,
    "container rm-all":     ...,
    "container rm-stopped": ...,
    "image rm-all":         ...,
    "image clean":          ...,
    "volume clean":         ...,
    "network clean":        ...,
    "system prune":         ...,
    "system df":            ...,
    "overview":             ...,
}
```

**Refactor to avoid duplication:** the per-resource logic (list → filter →
confirm → act) currently lives inside the command closures in
`commands/*.go`. Extract each into a plain function that takes
`(ctx, engine.API, engine.Selector, flags)` so *both* the Cobra command and the
verb registry call the same code. This keeps one implementation per operation.

Concretely: `containerRemoveAll(sel)` in `commands/container.go` becomes a thin
wrapper over a shared `engine`-level or `ops`-level function that both the CLI
and the executor invoke.

### 1.2 Executor dispatch

```go
func Run(ctx context.Context, eng engine.API, cmd config.CustomCommand, opts Options) error {
    for i, step := range cmd.Steps {
        line := strings.TrimSpace(step.Command())
        if line == "" { continue }

        if !step.IsShell() {
            if verb, ok := lookupVerb(line); ok {
                if err := verb(ctx, eng, opts); err != nil {
                    return fmt.Errorf("step %d (%q): %w", i+1, line, err)
                }
                continue
            }
            // scalar, not a known verb → fall through to shell (back-compat)
        }
        if err := execExternal(ctx, line, opts); err != nil {
            return fmt.Errorf("step %d (%q): %w", i+1, line, err)
        }
    }
    return nil
}
```

**Back-compat decision (needs sign-off):** a scalar step that is *not* a known
verb currently shells out. Two options:

- **(A) Keep shelling out** for unknown scalars (fully backward compatible;
  existing `docker compose up -d` configs keep working).
- **(B) Reserve scalar form for verbs only** and require `shell:` for anything
  external (cleaner model, but breaks existing configs).

Recommend **(A)** for v0.3 with a deprecation note steering external commands
toward `shell:`, then reconsider (B) at v1.0. This matches the README's stated
intent without breaking users.

### 1.3 Wire the engine into the executor

`newCustomCmd` in `commands/custom.go` must now pass a live `engine.API` into
`executor.Run`. Route it through `withEngine` so native-verb custom commands get
the same daemon-reachability check as built-ins. A custom command that is
*all* `shell:` steps should still work when Docker is down — so only connect the
engine lazily (on first native verb) or tolerate a `Ping` failure when no verb
is present. Simplest: connect lazily via a `func() (engine.API, error)` the
executor calls the first time it hits a native verb.

### 1.4 Config validation

- Validate verb names at load/registration time and `ui.Warn` on an unknown
  scalar that *looks* like a verb (helps typos).
- Update `examples/config.yaml` with a native-verb example, e.g.:

  ```yaml
  commands:
    nuke-dev:
      description: Tear down the dev stack safely
      steps:
        - container rm-all        # native verb, no --allow-shell
        - image clean             # native verb
        - shell: docker compose up -d   # explicit external step
  ```

### 1.5 Tests

- Verb dispatch: scalar `container rm-all` calls the fake's `RemoveContainer`;
  never calls `execExternal`.
- `--dry-run`: native verbs preview, don't act; `shell:` steps print "would run".
- Unknown scalar under option (A): shells out (gated by `--allow-shell`).
- Native verbs honor `--yes` / prompt.

**Acceptance:** a custom command of native verbs runs with **no** `--allow-shell`
and performs the same safe list→confirm→act flow as the built-ins.

---

## Phase 2 — Docker CLI plugin entry point

Goal: `docker toolbox overview` works, alongside the standalone binary.
`main.go` is already described as "plugin-ready," so this is mostly a metadata
command + build/install target.

### 2.1 Plugin metadata

Docker CLI plugins must (a) be named `docker-toolbox` on `PATH`, and (b)
respond to `docker-cli-plugin-metadata` with a JSON blob
(`SchemaVersion`, `Vendor`, `Version`, `ShortDescription`).

- Add a hidden `docker-cli-plugin-metadata` command to the root tree that prints
  the JSON and exits.
- Detect plugin invocation: when argv is `docker toolbox …`, the CLI calls the
  binary as `docker-toolbox toolbox …`. Handle the injected `toolbox` arg so the
  root command still resolves subcommands correctly (either strip it in `main`
  or register `toolbox` as a passthrough parent).

### 2.2 Build & install

- `cmd/docker-toolbox/` build target (or reuse the existing binary named
  appropriately) that installs to `~/.docker/cli-plugins/docker-toolbox`.
- Document both install paths in the README (standalone vs. plugin).

### 2.3 Tests / verification

- Unit test the metadata command's JSON output.
- Manual: `go build -o ~/.docker/cli-plugins/docker-toolbox ./cmd/... && docker toolbox overview`.

**Acceptance:** `docker toolbox --help` lists the same command tree; a smoke
command (`overview`) works through the plugin.

---

## Phase 3 (stretch) — Variables & hooks

Only if Phases 0–2 land comfortably. Adds the last two roadmap words for v0.3.

### 3.1 Variables

- `variables:` map in config, plus `${VAR}` / env substitution in step strings.
- Resolve at load time; document precedence (config var > env > empty).

### 3.2 Hooks

- Optional `pre:` / `post:` step lists on a custom command, run before/after the
  main steps; `post` runs even on failure if marked `always`.
- Reuse the same executor dispatch (native verbs + shell) for hook steps.

Keep this minimal — it's a stretch, and over-designing hooks now risks the v0.4
TUI timeline.

---

## Cross-cutting

- **Context cancellation:** replace `context.Background()` in `withEngine` and
  `custom.go` with a signal-aware context (`signal.NotifyContext` on
  SIGINT/SIGTERM) so long bulk operations stop cleanly. Small, high-value,
  fits naturally alongside the executor rework.
- **Docs:** update `README.md` (custom-command section, plugin install, roadmap
  bump v0.3→current) and `dockerToolBox.md` (design note on native verbs).
- **CI (recommended):** add a GitHub Actions workflow running `go vet ./...` and
  `go test ./...` so the new test suite actually guards regressions.

---

## Suggested sequencing & PRs

| PR | Scope | Depends on |
|----|-------|-----------|
| 1 | Phase 0: engine interface + `withEngine` seam + fake + first command tests | — |
| 2 | Phase 1: shared ops extraction + verb registry + executor dispatch + tests | PR 1 |
| 3 | Phase 2: plugin metadata + build/install + docs | PR 1 |
| 4 | Cross-cutting: signal context + CI workflow | PR 1 |
| 5 | Phase 3 (stretch): variables & hooks | PR 2 |

PRs 2 and 3 are independent after PR 1 and can go in parallel.

---

## Open decisions (need a call before coding)

1. **Unknown scalar steps:** option (A) keep shelling out (back-compat) vs.
   (B) reserve scalar for verbs only. Plan assumes **(A)**.
2. **Verb naming:** space-delimited (`"container rm-all"`) vs. colon
   (`"container:rm-all"`). Plan assumes space to mirror the CLI exactly.
3. **Engine when Docker is down:** lazy-connect on first native verb (so
   all-`shell:` custom commands work offline) vs. always `Ping` first. Plan
   assumes **lazy**.
4. **Where shared ops live:** new `internal/ops` package vs. exported functions
   in `commands`. Plan leans toward a small `ops` package to avoid an
   `executor → commands` import cycle.

---

## Acceptance criteria for v0.3

- [ ] `go test ./...` exercises `commands` and `executor` against a fake engine.
- [ ] A custom command composed of native verbs runs with no `--allow-shell`
      and follows list→confirm→act with `--dry-run`/`--yes` support.
- [ ] Existing shell-based custom commands still work unchanged.
- [ ] `docker toolbox overview` works via the CLI plugin.
- [ ] Ctrl-C cleanly cancels an in-progress bulk operation.
- [ ] README + design doc updated; roadmap marks v0.3 current.