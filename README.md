# DockerToolBox

> Turn complex, repetitive Docker operations into simple, safe commands.

DockerToolBox is a Go CLI built on top of the Docker Engine API. It collapses
the everyday "stop everything, remove everything, prune the rest" dance into a
handful of clear commands — and it never destroys anything without showing you
what it's about to do first.

```
dockertoolbox overview
dockertoolbox container rm-all
dockertoolbox system reset --dry-run
```

- **Status:** v0.1 (early, but working end to end)
- **Distribution:** standalone `dockertoolbox` binary (a `docker toolbox …`
  CLI plugin is planned — see [Roadmap](#roadmap))
- **Requires:** Go 1.26+ to build, and a running Docker daemon to use the
  built-in commands

---

## Table of contents

- [Why](#why)
- [Features](#features)
- [Installation](#installation)
- [Quick start](#quick-start)
- [Commands](#commands)
- [Global flags & safety model](#global-flags--safety-model)
- [Custom commands](#custom-commands)
- [Architecture](#architecture)
- [Project layout](#project-layout)
- [Development](#development)
- [Roadmap](#roadmap)
- [License](#license)

---

## Why

Cleaning up a Docker environment usually means stringing together commands you
half-remember:

```bash
docker stop $(docker ps -q)
docker rm $(docker ps -aq)
docker rmi $(docker images -q)
docker volume prune -f
```

They're easy to get wrong, easy to run against the wrong context, and offer no
preview before they delete things. DockerToolBox replaces them with named
commands that:

- show you exactly what will be affected,
- ask before doing anything destructive,
- support a `--dry-run` mode everywhere, and
- let you save your own multi-step workflows in a config file.

## Features

- **Environment overview** — daemon version and counts of containers, images,
  and volumes at a glance.
- **Bulk container management** — stop all, remove all, or remove only stopped
  containers.
- **Image management** — remove all images or prune dangling ones.
- **Volume cleanup** — prune unused volumes.
- **System reset** — one command to clear containers, images, and unused
  volumes, behind a single confirmation.
- **Custom commands** — define reusable workflows in
  `~/.docker-toolbox/config.yaml`.
- **Safe by default** — destructive actions list their targets and require
  confirmation; `--dry-run` previews without changing anything.
- **Readable output** — colors and status symbols (`✓ ⚠ ✗ ℹ`), with
  [`NO_COLOR`](https://no-color.org) support.

## Installation

### From source (recommended for now)

```bash
git clone https://github.com/danjor667/dockerToolBox
cd dockerToolBox
go install ./cmd/dockertoolbox
```

`go install` places the binary in your Go bin directory (`go env GOPATH`/bin,
typically `~/go/bin`). Make sure that directory is on your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"   # add to ~/.zshrc or ~/.bashrc
```

> **zsh tip:** if you just installed the binary into an already-open shell and
> still see `command not found`, run `hash -r` (or open a new tab) to refresh
> zsh's command cache.


## Quick start

```bash
# See what's in your Docker environment
dockertoolbox overview

# Preview removing every container (nothing is changed)
dockertoolbox container rm-all --dry-run

# Actually remove them (you'll be asked to confirm)
dockertoolbox container rm-all

# Remove them without the prompt (use with care)
dockertoolbox container rm-all --yes
```

If Docker isn't running, built-in commands fail cleanly:

```
docker daemon unreachable (is Docker running?): ...
```

## Commands

| Command                          | Description                                              |
| -------------------------------- | -------------------------------------------------------- |
| `dockertoolbox overview`         | Show daemon version and resource counts.                 |
| `dockertoolbox container ls`     | List containers.                                         |
| `dockertoolbox container stop-all`   | Stop all running containers.                         |
| `dockertoolbox container rm-all`     | Remove all containers (running and stopped).         |
| `dockertoolbox container rm-stopped` | Remove only stopped containers.                      |
| `dockertoolbox image ls`         | List images.                                             |
| `dockertoolbox image rm-all`     | Remove all images.                                       |
| `dockertoolbox image clean`      | Prune dangling images.                                   |
| `dockertoolbox volume ls`        | List volumes.                                            |
| `dockertoolbox volume clean`     | Prune unused volumes.                                    |
| `dockertoolbox network ls`       | List networks.                                           |
| `dockertoolbox network clean`    | Prune unused networks.                                   |
| `dockertoolbox system df`        | Show Docker disk usage.                                  |
| `dockertoolbox system prune`     | Remove unused data (`--volumes` to also prune volumes).  |
| `dockertoolbox system reset`     | Remove all containers and images, then prune volumes.    |

Run `dockertoolbox <command> --help` for details on any command.

### Selectors (`--name` / `--label`)

The listing and bulk container/image commands accept selectors so you can act
on a subset instead of all-or-nothing:

| Flag             | Semantics                                                            |
| ---------------- | ------------------------------------------------------------------- |
| `--name <text>`  | Match resources whose name/tag **contains** `<text>`. Repeatable; multiple values are **OR**'d. |
| `--label <l>`    | Match resources carrying label `<l>` (`key` or `key=value`). Repeatable; multiple values are **AND**'d. |

```bash
# List only containers whose name contains "web"
dockertoolbox container ls --name web

# Remove containers labeled env=dev (with confirmation)
dockertoolbox container rm-all --label env=dev

# Combine: name contains "api" AND label tier=backend
dockertoolbox container ls --name api --label tier=backend
```

Selectors compose with `--dry-run` and `--yes`, so you can preview exactly
which resources a bulk command would affect.

## Global flags & safety model

These flags apply to every command:

| Flag            | Description                                                        |
| --------------- | ----------------------------------------------------------------- |
| `--dry-run`     | Show what would happen without making any changes.                |
| `-y`, `--yes`   | Skip confirmation prompts for destructive actions.                |
| `--allow-shell` | Permit custom commands to run external shell commands (off by default). |

Every destructive command follows the same contract:

1. **List** the items that will be affected.
2. Under `--dry-run`, **stop here** — nothing is changed.
3. Otherwise, **prompt** for confirmation (default answer is *No*).
4. `--yes` skips the prompt for automation.

`system reset` is the most destructive command, so it summarizes everything it
will touch and asks for a single confirmation up front.

## Custom commands

Define your own workflows in `~/.docker-toolbox/config.yaml`. A starter file
lives at [`examples/config.yaml`](examples/config.yaml).

```yaml
commands:
  deploy-dev:
    description: Deploy development environment
    steps:
      - docker compose down
      - docker compose pull
      - docker compose up -d
```

Each entry becomes a top-level command:

```bash
dockertoolbox deploy-dev --dry-run      # preview the steps
dockertoolbox deploy-dev --allow-shell  # actually run them
```

### Step forms

A step may be a plain string or an explicit shell step:

```yaml
steps:
  - docker compose up -d              # scalar form
  - shell: "docker compose logs -f"   # explicit shell form
```

### Execution & safety

In v0.1, custom steps run as **external shell commands**, which is a security
consideration. Accordingly:

- Steps require the global `--allow-shell` flag. Without it, the command is
  refused with a clear message.
- With `--dry-run`, steps are printed but never executed (no `--allow-shell`
  needed).
- A custom command whose name collides with a built-in is skipped with a
  warning, so it can never shadow `container`, `image`, etc.

> The scalar vs. `shell:` distinction is intentional: scalar steps are
> reserved for future **SDK-native toolbox verbs** that will run common
> workflows without `--allow-shell`. See the [Roadmap](#roadmap).

## Architecture

```
        Docker CLI / terminal
                 │
          dockertoolbox
                 │
   ┌─────────────┼──────────────┐
 cleanup     monitoring     automation
                 │
        Docker Engine API
```

The codebase is organized around a single rule: **only one package talks to
Docker, and only one package shells out.**

- `internal/engine` is the *only* package that calls the Docker SDK. All
  built-in commands go through it and never shell out.
- `internal/executor` is the *only* place that runs external commands, and only
  for user-defined custom workflows (gated by `--allow-shell`).

This keeps the Docker boundary and the "runs arbitrary commands" boundary
separate and easy to audit.

## Project layout

```
dockerToolBox/
├── cmd/
│   └── dockertoolbox/
│       └── main.go          # thin entry point (plugin-ready)
├── internal/
│   ├── commands/            # Cobra command tree + safety helpers
│   ├── engine/              # the ONLY package that calls the Docker SDK
│   ├── executor/            # runs user-defined custom-command steps
│   ├── config/              # loads ~/.docker-toolbox/config.yaml
│   └── ui/                  # styled terminal output (ANSI)
├── examples/
│   └── config.yaml          # sample user config
├── go.mod
├── dockerToolBox.md         # design/vision document
└── README.md
```

Non-command packages live under `internal/` so they can't be imported by
outside modules.

## Development

```bash
# Build everything
go build ./...

# Vet
go vet ./...

# Run without installing
go run ./cmd/dockertoolbox overview

# Re-install after changes
go install ./cmd/dockertoolbox
```

**Tech stack:** Go, [Cobra](https://github.com/spf13/cobra) for the CLI, the
[Docker Go SDK](https://pkg.go.dev/github.com/docker/docker/client) for engine
communication, and [`gopkg.in/yaml.v3`](https://gopkg.in/yaml.v3) for config.
The interactive terminal UI (Bubble Tea / Lip Gloss) is planned for v0.4; v0.1
keeps output to plain ANSI so the binary stays lean.

> **Note on dependencies:** the project pins
> `github.com/docker/go-connections v0.5.0` to stay compatible with Docker SDK
> `v27.5.1` (newer `go-connections` removed a symbol the SDK references).

## Roadmap

| Version | Focus |
| ------- | ----- |
| **0.1** | Standalone CLI, Docker API connection, built-in commands, global safety flags, custom commands |
| **0.2** *(current)* | Resource listing (`ls`), network commands, `system prune`/`system df`, and `--name`/`--label` selectors |
| **0.3** | Docker CLI plugin entry point (`docker toolbox …`); SDK-native custom-command verbs; variables and hooks |
| **0.4** | Interactive terminal UI (Bubble Tea / Lip Gloss) |
| **1.0** | Plugin ecosystem |

