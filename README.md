# ezmux

ezmux is a small tmux popup navigator for macOS and Linux. It shows sessions, their windows, and a live read-only capture of the selected window's active pane.

## Local POC

Requirements: Go 1.26 or newer and tmux 3.x. Run these commands from the repository:

```sh
go build -o bin/ezmux ./cmd/ezmux
./bin/ezmux
```

The standalone command expects to run inside tmux. Close it with `q` or `Esc`.

To test it as a popup without changing your tmux configuration permanently:

```sh
go build -o bin/ezmux ./cmd/ezmux
tmux display-popup -E -w 85% -h 80% -T ezmux "$PWD/bin/ezmux"
```

The popup controls are `h/l` or left/right for sessions, `j/k` or up/down for windows, `0-9` for a tmux window index, `Enter` to switch, and `Esc`/`q` to close.

## Key Binding

The repository includes `ezmux.tmux`, which binds prefix+w to the binary relative to the plugin directory. Build the binary first, then source it for the current tmux server:

```sh
go build -o bin/ezmux ./cmd/ezmux
bash ./ezmux.tmux
```

This does not modify `~/.tmux.conf`. For TPM later, the same script can be sourced by the plugin manager.

## Limitations

This POC does not create or delete sessions/windows, support mouse clicks, or provide configuration or TPM installation automation. Capture output depends on tmux's ANSI output and is intentionally limited to the last 100 lines.
