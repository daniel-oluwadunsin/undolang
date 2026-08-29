#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
build_dir=$(mktemp -d "${TMPDIR:-/tmp}/undolang-repro.XXXXXX")
trap 'rm -rf "$build_dir"' EXIT HUP INT TERM
cd "$repo_dir"

for output in undo-a undo-b; do
  CGO_ENABLED=0 GOTOOLCHAIN=go1.27.0 GOPROXY=off \
    go build -trimpath -buildvcs=false -o "$build_dir/$output" ./cmd/undo
done
go run ./tools/buildproof "$build_dir/undo-a" "$build_dir/undo-b"
printf '%s\n' "reproducible build: PASS"
