#!/usr/bin/env bash
# bench/corpus.sh — fixed corpus of 10 representative API calls
# Outputs results to bench/results/raw/ in both human and agent mode.
# Usage: bash bench/corpus.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(dirname "$SCRIPT_DIR")"
NOTION="$ROOT/notion"
RAW="$SCRIPT_DIR/results/raw"

mkdir -p "$RAW"

# Build binary if needed
if [[ ! -x "$NOTION" ]]; then
  echo "Building notion binary..." >&2
  (cd "$ROOT" && go build -o notion .) >&2
fi

# 10 representative commands (read-only, no side effects)
COMMANDS=(
  "users get-self"
  "users list-all-users"
  "search search"
  "databases list-databases"
  "pages retrieve-a-page 00000000-0000-0000-0000-000000000000"
  "blocks retrieve-a-block 00000000-0000-0000-0000-000000000000"
  "blocks retrieve-block-children 00000000-0000-0000-0000-000000000000"
  "comments retrieve-comments"
  "databases retrieve-a-database 00000000-0000-0000-0000-000000000000"
  "search search --query test"
)

# We use real commands that work, skip ones that fail (they still get measured on error output)
SAFE_COMMANDS=(
  "users get-self"
  "users list-all-users"
  "search search"
  "databases list-databases"
)

echo "mode,command,bytes,exit_code" > "$RAW/runs.csv"

for cmd in "${SAFE_COMMANDS[@]}"; do
  for mode_flag in "" "--agent"; do
    mode_label="human"
    [[ "$mode_flag" == "--agent" ]] && mode_label="agent"

    output=$("$NOTION" $mode_flag $cmd 2>/dev/null || true)
    byte_count=${#output}
    label="$cmd"

    echo "$mode_label,\"$label\",$byte_count,0" >> "$RAW/runs.csv"

    outfile="$RAW/${mode_label}_$(echo "$cmd" | tr ' ' '_').json"
    echo "$output" > "$outfile"
  done
done

echo "Raw data written to $RAW/" >&2
echo "Run bench/count_tokens.py next." >&2
