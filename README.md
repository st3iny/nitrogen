# nitrogen

A file watcher and service orchestrator for Go projects. Nitrogen watches your
source files, rebuilds your project when they change, and restarts your services
automatically — giving you hot-reload during local development.

## Features

- **File watching** — monitors write, create, rename, and delete events across
  your project directories.
- **Debouncing** — waits for a configurable quiet period before rebuilding, so a
  burst of file changes triggers a single rebuild.
- **Sequential build steps** — runs your build commands in order and stops on the
  first failure. Services are only restarted after a successful build.
- **Service management** — starts multiple services in parallel and restarts them
  together on each successful build. Services that exit on their own are restarted
  automatically.
- **`go.mod` replace support** — local paths from `replace` directives in `go.mod`
  are automatically added to the watch set, so changes to local dependencies
  trigger a rebuild.
- **Exclude patterns** — glob patterns to skip directories you don't want watched
  (e.g. `.git`, `.idea`).

## Installation

Requires Go 1.26 or newer.

```bash
make build
```

This produces the binary at `bin/nitrogen`.

Alternatively, install with the Go toolchain:

```bash
go install github.com/st3iny/nitrogen/cmd/nitrogen@latest
```

## Usage

Create a `.nitrogen.yaml` in your project root (see [Configuration](#configuration)),
then run the `nitrogen` binary.

### Flags

| Flag      | Default          | Description              |
| --------- | ---------------- | ------------------------ |
| `-config` | `.nitrogen.yaml` | Path to the config file. |

```bash
./bin/nitrogen --config path/to/config.yaml
```

## Configuration

Nitrogen is configured with a YAML file (`.nitrogen.yaml` by default).

```yaml
# Log level: debug | info | warn | error (default: info)
#logLevel: debug

# How long to wait after the last file change before rebuilding.
debounceDelay: 500ms

# Build steps, run in order. Each step is a command as a list of arguments.
# Building stops on the first failing step.
build:
  - ["make", "-f", "./test/Makefile", "build"]

# Services to run. All are restarted together after a successful build.
services:
  - command: ["./bin/svc1"]
  - command: ["./bin/svc2"]

# Watch configuration.
watch:
  # Extra directories to watch (the current directory and go.mod replace
  # directive paths are always watched).
  include:
    - ../shared-lib
  # Glob patterns for directories to ignore.
  exclude:
    - .git*
    - .idea

# Path to the go.mod file used to resolve replace directives (default: go.mod).
goModFile: go.mod
```

### Fields

| Field           | Type                  | Default   | Description                                                        |
| --------------- | --------------------- | --------- | ------------------------------------------------------------------ |
| `logLevel`      | string                | `info`    | Logging verbosity (`debug`, `info`, `warn`, `error`).              |
| `debounceDelay` | duration              | `500ms`   | Quiet period after a change before rebuilding.                     |
| `build`         | list of commands      | —         | Build steps, run sequentially. Each is a list of arguments.        |
| `services`      | list of services      | —         | Services to run; each has a `command` argument list.               |
| `watch.include` | list of paths         | —         | Extra directories to watch (in addition to `.` and replace paths). |
| `watch.exclude` | list of glob patterns | —         | Directory patterns to ignore.                                      |
| `goModFile`     | string                | `go.mod`  | `go.mod` file to parse for `replace` directives.                   |

## How it works

Nitrogen runs a small pipeline of goroutines:

1. **Watcher** — reports filesystem events for the configured directories.
2. **Debouncer** — collapses rapid changes into a single rebuild trigger after
   `debounceDelay`.
3. **Builder** — runs the configured `build` steps in order, stopping on the first
   failure.
4. **Runner** — on a successful build, stops the running services and starts them
   again. If a build fails, the services keep running unchanged until the next
   successful build.

Local dependency paths from `go.mod` `replace` directives are watched too, so
editing a local library rebuilds and restarts your services.

## Development

```bash
make build   # build bin/nitrogen with version/build-date ldflags
make run     # run directly via `go run`
```

### Testing

```bash
make setup         # install tooling (go-test-coverage)
make test          # run tests and write coverage.out
make coverage      # run tests, print per-func coverage, and enforce thresholds
make coverage-html # run tests and open the HTML coverage report
```

Coverage thresholds are enforced by
[`go-test-coverage`](https://github.com/vladopajic/go-test-coverage) via
`.testcoverage.yml` (100% file/package/total, excluding the `internal/nitrogen`
CLI entry point). Lines that are impractical to cover are annotated with
`// coverage-ignore` comments.

### Dependencies

- [`fsnotify/fsnotify`](https://github.com/fsnotify/fsnotify) — cross-platform file system notifications
- [`gobwas/glob`](https://github.com/gobwas/glob) — glob pattern matching for exclude rules
- [`goccy/go-yaml`](https://github.com/goccy/go-yaml) — YAML config parsing
- [`golang.org/x/mod`](https://pkg.go.dev/golang.org/x/mod) — parsing `go.mod` replace directives
