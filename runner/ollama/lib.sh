# Sourced by the other scripts. Every path and port is resolved in one place so
# the three of them cannot drift apart.

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
PREFIX="${OLLAMA_PREFIX:-$HOME/.local}"
OLLAMA="$PREFIX/bin/ollama"
CONFIG="${GATEWAY_CONFIG:-$ROOT/gateway/configs/config.yaml}"

export OLLAMA_HOST="${OLLAMA_HOST:-127.0.0.1:11434}"

require_binary() {
  if [ ! -x "$OLLAMA" ]; then
    echo "ollama is not installed at $OLLAMA — run install.sh first" >&2
    exit 1
  fi
}

# The catalogue in the gateway config decides which models exist. Reading it
# here keeps this directory from becoming a second, stale copy of that list.
local_models() {
  awk '
    /^  - id:/ { id = $3; gsub(/"/, "", id); next }
    /^    provider:/ { if ($2 == "local" && id != "") print id; id = "" }
  ' "$CONFIG"
}
