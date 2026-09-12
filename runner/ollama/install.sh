#!/usr/bin/env bash
# Unpacks an Ollama release into $OLLAMA_PREFIX (default ~/.local), which needs
# no root. The binary finds its libraries at ../lib/ollama, so bin and lib must
# be unpacked under the same prefix.
set -euo pipefail
source "$(dirname "$0")/lib.sh"

ARCHIVE="${1:-$HOME/Downloads/ollama-linux-amd64.tar.zst}"

if [ ! -f "$ARCHIVE" ]; then
  echo "no archive at $ARCHIVE" >&2
  echo "download one from https://github.com/ollama/ollama/releases and pass its path" >&2
  exit 1
fi

mkdir -p "$PREFIX"

case "$ARCHIVE" in
  *.zst) tar --zstd -xf "$ARCHIVE" -C "$PREFIX" ;;
  *)     tar -xzf "$ARCHIVE" -C "$PREFIX" ;;
esac

"$OLLAMA" --version 2>&1 | grep -v "^Warning" || true
echo "installed to $PREFIX"
echo "add it to your shell: export PATH=\"$PREFIX/bin:\$PATH\""
