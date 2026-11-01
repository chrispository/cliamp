# cliamp-lite

A stripped-down fork of [cliamp](https://github.com/bjarneo/cliamp): one
radio station, [9128.live](https://9128.live), and all of cliamp's built-in
terminal visualizers. No playlists, providers, plugins, config or IPC.

## Keys

| Key | Action |
|-----|--------|
| `↑` / `↓` | Move through the station and the visualizer list. The visualizer changes as you move. |
| `Enter` / `Space` | Play or stop the station (on the station row) |
| `q` / `Ctrl+C` | Quit |

## Build

`mise.toml` pins Go. On Linux the build needs the ALSA, FLAC, Vorbis, Ogg and
mpg123 headers (`alsa-lib flac libvorbis libogg mpg123` on Arch,
`libasound2-dev libflac-dev libvorbis-dev libogg-dev libmpg123-dev` on Debian
and Ubuntu). On PipeWire or PulseAudio, install `pipewire-alsa` or
`pulseaudio-alsa`.

```sh
mise exec -- make build     # ./cliamp-lite
mise exec -- make install   # ~/.local/bin/cliamp-lite
mise exec -- make check     # fmt, vet, test
```

## Layout

- `main.go` - the whole app: the Bubbletea model, the station and the list.
- `player/` - the audio engine from cliamp.
- `ui/` - the visualizers from cliamp.
- `theme/`, `playlist/`, `applog/`, `internal/` - helpers the player and ui
  packages import.

Licensed under the MIT license, as upstream. See `LICENSE`.
