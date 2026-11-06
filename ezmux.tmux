#!/usr/bin/env bash

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
tmux bind-key -T prefix w display-popup -E -b rounded -s 'bg=#1e1e2e,fg=#cdd6f4' -S 'fg=#585b70' -e EZMUX_POPUP=1 -w 85% -h 80% -T "ezmux" "$CURRENT_DIR/bin/ezmux"
