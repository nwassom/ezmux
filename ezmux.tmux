#!/usr/bin/env bash

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
tmux bind-key -T prefix w display-popup -E -w 85% -h 80% -T "ezmux" "$CURRENT_DIR/bin/ezmux"
