# Sourced by the other scripts. Every path and port is resolved in one place so
# the three of them cannot drift apart.

PREFIX="${OLLAMA_PREFIX:-$HOME/.local}"
OLLAMA="$PREFIX/bin/ollama"

export OLLAMA_HOST="${OLLAMA_HOST:-127.0.0.1:11434}"
export OLLAMA_MODELS="${OLLAMA_MODELS:-$HOME/.ollama/models}"

require_binary() {
  if [ ! -x "$OLLAMA" ]; then
    echo "ollama is not installed at $OLLAMA — run install.sh first" >&2
    exit 1
  fi
}
