#!/usr/bin/env bash
# Downloads every model the gateway catalogue assigns to the local provider.
# Safe to re-run: a model already present is left alone.
set -euo pipefail
source "$(dirname "$0")/lib.sh"
require_binary

if ! curl -sf -o /dev/null --max-time 5 "http://$OLLAMA_HOST/api/tags"; then
  echo "no server on $OLLAMA_HOST — start serve.sh first" >&2
  exit 1
fi

models=$(local_models)
if [ -z "$models" ]; then
  echo "no local models in $CONFIG" >&2
  exit 1
fi

for model in $models; do
  echo "pulling $model"
  "$OLLAMA" pull "$model"
done

"$OLLAMA" list
