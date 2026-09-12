#!/usr/bin/env bash
# Runs the model server in the foreground. The gateway reaches it over the
# OpenAI-compatible routes it serves at /v1, so providers.local.base_url in the
# gateway config must match OLLAMA_HOST.
set -euo pipefail
source "$(dirname "$0")/lib.sh"
require_binary

echo "serving on http://$OLLAMA_HOST"
exec "$OLLAMA" serve
