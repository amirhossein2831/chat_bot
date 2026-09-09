#!/usr/bin/env bash
# Downloads the models listed below. Add a line to pull another one; the tag is
# part of the name and decides the quantisation and size.
# Safe to re-run: a model already present is left alone.
set -euo pipefail
source "$(dirname "$0")/lib.sh"
require_binary

MODELS=(
  qwen2.5:3b
)

if ! curl -sf -o /dev/null --max-time 5 "http://$OLLAMA_HOST/api/tags"; then
  echo "no server on $OLLAMA_HOST — start serve.sh first" >&2
  exit 1
fi

for model in "${MODELS[@]}"; do
  echo "pulling $model"
  "$OLLAMA" pull "$model"
done

"$OLLAMA" list
