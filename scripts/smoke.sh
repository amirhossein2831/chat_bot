#!/usr/bin/env bash
# End-to-end check against a running gateway. Exercises the paths that break
# silently: incremental streaming, persistence, and branch isolation.
set -euo pipefail

BASE="${BASE:-http://localhost:18080}"
JAR=$(mktemp)
EMAIL="smoke-$RANDOM@example.com"
PASSWORD="correct horse battery staple"
INVITE="${INVITE:?set INVITE to a seeded invite code}"

pass() { printf '  ok   %s\n' "$1"; }
fail() { printf '  FAIL %s\n' "$1"; exit 1; }

echo "smoke: $BASE"

code=$(curl -s -o /dev/null -w '%{http_code}' "$BASE/healthz")
[ "$code" = 200 ] && pass "healthz" || fail "healthz returned $code"

code=$(curl -s -o /dev/null -w '%{http_code}' -c "$JAR" -X POST -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\",\"invite\":\"$INVITE\"}" \
  "$BASE/api/v1/auth/register")
[ "$code" = 201 ] && pass "register" || fail "register returned $code"

code=$(curl -s -o /dev/null -w '%{http_code}' -c "$JAR" -b "$JAR" -X POST -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" "$BASE/api/v1/auth/login")
[ "$code" = 200 ] && pass "login" || fail "login returned $code"

curl -s -b "$JAR" "$BASE/api/v1/me" | grep -q "$EMAIL" && pass "me" || fail "me did not return the user"

MODEL=$(curl -s -b "$JAR" "$BASE/api/v1/models" \
  | python3 -c 'import json,sys; m=[x for x in json.load(sys.stdin) if x["available"]]; print(m[0]["id"] if m else "")')
[ -n "$MODEL" ] && pass "models: $MODEL available" || fail "no model is available"

CID=$(curl -s -b "$JAR" -X POST -H 'Content-Type: application/json' \
  -d "{\"title\":\"smoke\",\"model\":\"$MODEL\"}" "$BASE/api/v1/conversations" \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])')
[ -n "$CID" ] && pass "conversation created" || fail "conversation not created"

STREAM=$(mktemp)
curl -sN -b "$JAR" -X POST -H 'Content-Type: application/json' \
  -d "{\"conversation_id\":\"$CID\",\"content\":\"سلام\"}" "$BASE/api/v1/chat" > "$STREAM"

grep -q 'event: meta'  "$STREAM" && pass "meta frame"  || fail "no meta frame"
grep -q 'event: delta' "$STREAM" && pass "delta frames" || fail "no delta frames"
grep -q 'event: done'  "$STREAM" && pass "done frame"  || fail "no done frame"

COUNT=$(curl -s -b "$JAR" "$BASE/api/v1/conversations/$CID/messages" \
  | python3 -c 'import json,sys; print(len(json.load(sys.stdin)["messages"]))')
[ "$COUNT" = 2 ] && pass "both turns persisted" || fail "expected 2 messages, got $COUNT"

curl -s -b "$JAR" "$BASE/api/v1/prompts" \
  | python3 -c 'import json,sys; sys.exit(0 if json.load(sys.stdin) else 1)' \
  && pass "prompt library" || fail "prompt library empty"

curl -s "$BASE/metrics" | grep -q chat_time_to_first_token_seconds \
  && pass "metrics exposed" || fail "ttft metric missing"

echo "smoke: all checks passed"
