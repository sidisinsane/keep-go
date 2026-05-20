#!/usr/bin/env bash
# ---
# description: Generate Markdown documentation from all Keep JSON schemas and post-process the output.
# usage: scripts/generate-schema-docs.sh
# exits:
#   0: Documentation generated successfully.
#   1: yadox command not found or generation failed.
# requires: https://github.com/enpace/yadox
# ---
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$SCRIPT_DIR/.."
SCHEMA_DIR="$ROOT/schema"
DOCS_DIR="$ROOT/docs"

mkdir -p "$DOCS_DIR"

for schema in "$SCHEMA_DIR"/keep-*.schema.json; do
    base="$(basename "$schema" .schema.json)"
    name="${base#keep-}"
    output="$DOCS_DIR/${name}-schema.md"

    echo "Generating $output..."
    yadox generate --schemas "$schema" -o "$output" -w

    sed -i '' 's/<nil>/null/g' "$output"
done

echo "Done."
