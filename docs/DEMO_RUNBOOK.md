# UndoLang five-minute demo runbook

This is a copyable runbook for the real 0.1.0 binary. It assumes a macOS or
Linux shell, Go 1.27.0, and a checkout of this repository. It uses synthetic
fixture files only; it never fakes a transaction failure. The crash section
terminates the real `undo` process from outside the program.

## 0:00–0:25 — Build and position

```sh
CGO_ENABLED=0 GOTOOLCHAIN=go1.27.0 GOPROXY=off \
  go build -trimpath -buildvcs=false -o /tmp/undo-demo ./cmd/undo
UNDO=/tmp/undo-demo
DEMO_DIR=$(mktemp -d /tmp/undolang-demo.XXXXXX)
```

Say: “UndoLang is a crash-safe filesystem transaction runtime and tiny DSL.
The file is the program; each named transaction is its own rollback boundary.”

## 0:25–1:10 — Real language, check, and plan

```sh
cat > "$DEMO_DIR/migration.undo" <<'EOF'
transaction "stage" {
  require is_file "release/app-v2"
  mkdir "staging/v2"
  copy "release/app-v2" -> "staging/v2/app"
  assert is_file "staging/v2/app"
}

transaction "activate" {
  require is_file "staging/v2/app"
  copy "staging/v2/app" -> "bin/app" overwrite
  write "VERSION" = "2"
  assert contains "VERSION" "2"
}

transaction "cleanup" {
  require is_dir "legacy"
  delete "legacy"
  assert not_exists "legacy"
}
EOF
mkdir -p "$DEMO_DIR/all/release" "$DEMO_DIR/all/bin" \
  "$DEMO_DIR/all/staging/v2" "$DEMO_DIR/all/legacy"
printf '%s\n' 'app version 2' > "$DEMO_DIR/all/release/app-v2"
printf '%s\n' 'app version 1' > "$DEMO_DIR/all/bin/app"
printf '%s\n' '1' > "$DEMO_DIR/all/VERSION"
printf '%s\n' 'old generated file' > "$DEMO_DIR/all/legacy/old.txt"

"$UNDO" check "$DEMO_DIR/migration.undo" --root "$DEMO_DIR/all"
"$UNDO" plan "$DEMO_DIR/migration.undo" --root "$DEMO_DIR/all"
```

Point out that `stage` is exact/ready and later state-sensitive transactions
are deferred in the whole-program plan.

## 1:10–1:55 — Whole-program execution and selected transaction

```sh
"$UNDO" run "$DEMO_DIR/migration.undo" --root "$DEMO_DIR/all" --yes
cat "$DEMO_DIR/all/VERSION"
test ! -e "$DEMO_DIR/all/legacy"

mkdir -p "$DEMO_DIR/selected/staging/v2" "$DEMO_DIR/selected/bin"
printf '%s\n' 'app version 2' > "$DEMO_DIR/selected/staging/v2/app"
printf '%s\n' 'app version 1' > "$DEMO_DIR/selected/bin/app"
printf '%s\n' '1' > "$DEMO_DIR/selected/VERSION"
"$UNDO" run "$DEMO_DIR/migration.undo" --transaction activate \
  --root "$DEMO_DIR/selected" --yes --json
```

The first run proves source-order `stage -> activate -> cleanup`; the second
proves `--transaction activate` executes only the named transaction while the
whole source still validates.

## 1:55–2:25 — Agent contract and empty manifest

```sh
"$UNDO" capabilities --json
"$UNDO" schema --json
cat go.mod
GOTOOLCHAIN=go1.27.0 GOPROXY=off go list -m all
```

Say: “The agent interface is the CLI: discover capabilities and schema, check,
plan, obtain approval, then run with `--yes`. The module graph is one line.”

## 2:25–3:50 — Real process kill and recovery

Create a second real transaction that takes long enough for an external kill:

```sh
mkdir -p "$DEMO_DIR/crash"
dd if=/dev/zero of="$DEMO_DIR/crash/large-source" bs=1048576 count=512
cat > "$DEMO_DIR/crash.undo" <<'EOF'
transaction "crash-demo" {
  require is_file "large-source"
  write "CRASH_MARKER" = "started"
  copy "large-source" -> "large-copy"
  write "AFTER_COPY" = "committed"
  assert contains "AFTER_COPY" "committed"
}
EOF

"$UNDO" run "$DEMO_DIR/crash.undo" --root "$DEMO_DIR/crash" --yes --json \
  > "$DEMO_DIR/crash-run.json" &
crash_pid=$!
until test -e "$DEMO_DIR/crash/CRASH_MARKER"; do sleep 0.01; done
kill -TERM "$crash_pid"
wait "$crash_pid" || true

"$UNDO" recover --root "$DEMO_DIR/crash" --yes --json
test ! -e "$DEMO_DIR/crash/CRASH_MARKER"
test ! -e "$DEMO_DIR/crash/large-copy"
test ! -e "$DEMO_DIR/crash/AFTER_COPY"
```

The `kill -TERM` is the failure. Recovery is a fresh process reading the
durable journal and restoring the pre-transaction state. If the copy finishes
before the kill on a fast machine, do not present that as a crash: repeat with
a larger source file and kill while `large-copy` is being created. A successful
demo must show `recover` handling an actually unresolved transaction.

## 3:50–4:30 — Receipts

```sh
./scripts/deps-proof.sh
./scripts/repro-build.sh
```

Show the one-module output, the offline test/build pass, and the matching A/B
SHA-256 values. `STDLIB.md` contains the 15 concrete substitutions behind the
empty manifest.

## 4:30–5:00 — Close

Say: “UndoLang gives filesystem automation a visible plan, durable journal,
verified rollback, and explicit crash recovery. It does not claim isolation or
universal atomic visibility, and it needs no runtime package, service, or API
key.”

Do not claim Windows/Linux behavioral testing from these local commands; the
support matrix distinguishes the macOS arm64 test host from cross-build-only
release targets.
