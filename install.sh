#!/bin/sh
set -eu

source_binary=${1:-./undo}
install_dir=${2:-"${HOME}/.local/bin"}
temporary_binary=

if [ ! -f "$source_binary" ]; then
  if [ "$source_binary" != "./undo" ] || [ ! -f ./go.mod ]; then
    echo "UndoLang binary not found: $source_binary" >&2
    exit 1
  fi
  temporary_binary=$(mktemp "${TMPDIR:-/tmp}/undolang-install.XXXXXX")
  trap 'rm -f "$temporary_binary"' EXIT HUP INT TERM
  echo "Building UndoLang from the local source tree..."
  GOPROXY=off GOTOOLCHAIN=local CGO_ENABLED=0 go build -trimpath -buildvcs=false -o "$temporary_binary" ./cmd/undo
  source_binary=$temporary_binary
fi

mkdir -p "$install_dir"
cp "$source_binary" "$install_dir/undo"
chmod 0755 "$install_dir/undo"
echo "Installed UndoLang at $install_dir/undo"

case ":${PATH}:" in
  *":${install_dir}:"*) ;;
  *) echo "Add $install_dir to PATH to invoke 'undo' directly." ;;
esac

