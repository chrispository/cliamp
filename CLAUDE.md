# CLAUDE.md - cliamp-lite

A stripped-down fork of cliamp (Go + Bubbletea). It plays one radio station,
9128.live, and shows cliamp's built-in visualizers. `AGENTS.md` is a symlink
to this file. Edit `CLAUDE.md` only.

## Scope

Keep it minimal. The only features are the station (play and stop), the
visualizer list and the visualizers. The only keys are up, down, Enter or
Space, q and Ctrl+C. Do not add playlists, providers, plugins, config, IPC or
other shortcuts unless the user asks.

## Layout

| Path | Responsibility |
|------|----------------|
| `main.go` | The whole app: the station constants, the Bubbletea model, the tick that feeds the visualizer, and the view |
| `main_darwin.go` | Locks the main goroutine to the main OS thread on macOS with cgo |
| `player/` | The audio engine from upstream: decoding, the stream pipeline, ICY titles and the sample tap the visualizer reads |
| `ui/` | The visualizer engine: `visualizer.go`, the drivers in `vis_driver.go`, the mode table in `vis_registry.go` and one `vis_*.go` file per mode. `styles.go` holds the colors |
| `theme/`, `playlist/`, `applog/`, `internal/` | Upstream helpers that `player` and `ui` import |

To add a visualizer, add a const and a row in `ui/vis_registry.go` and a
`vis_*.go` file. The list in `main.go` reads `ui.VisModeNames()`, so it picks
the mode up by itself.

## Build and test

`mise.toml` pins Go. Run the targets through mise, for example
`mise exec -- make check`.

| Target | What it does |
|--------|--------------|
| `make build` | Builds `./cliamp-lite` |
| `make install` | Installs it into `~/.local/bin` |
| `make check` | `fmt`, `vet` and `test` |
| `make ci` | What CI runs: fmt-check, tidy-check, vet, staticcheck, govulncheck, race tests and `git diff --exit-code` |

Keep the `go` line in `go.mod`, the `go` pin in `mise.toml` and
`STATICCHECK_VERSION` in the `Makefile` in step.

On Linux the build needs the ALSA, FLAC, Vorbis, Ogg and mpg123 headers.

## Conventions

- Wrap errors with `fmt.Errorf("context: %w", err)`.
- Platform files use the GOOS suffix, or `_stub.go` / `_other.go` with a
  `//go:build` line.
- Favor table-driven tests, colocated with the source.
- No emojis in code or docs.
- Minimal diffs. No speculative abstractions.

## Commits

Every commit has a body. Subject: `area: what changed` (50 chars or less).
Body: what was wrong, what the fix does, and how it was verified.
