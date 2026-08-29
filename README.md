# UndoLang

UndoLang is a Go 1.27, standard-library-only filesystem transaction language. The current implementation covers the language front-end, traversal-resistant path capabilities, streamed conditions, and non-mutating planning. Transaction execution is intentionally not exposed until the durable runner and recovery engine are wired in Phase 6.

## Build

```sh
go build -trimpath -buildvcs=false -o undo ./cmd/undo
```

No runtime environment variables or API keys are required. `NO_COLOR` is accepted by convention; current output contains no ANSI color.

## Read-only workflow

```sh
./undo check migration.undo --root /target
./undo plan migration.undo --root /target
./undo plan migration.undo --transaction upgrade --root /target --json
./undo capabilities --json
./undo schema --json
```

Relative DSL paths resolve against `--root`, or the invocation working directory when omitted. The script's own directory never becomes the root implicitly. Absolute paths outside the root require repeatable `--allow-path` directory capabilities.

A file may contain multiple named transactions. Whole-program planning preflights the first transaction and marks later state-sensitive work deferred. A selected transaction receives an exact current-state plan.

## Current limitations

- `run`, rollback orchestration, and crash recovery are not exposed yet.
- Planning and condition reads use `os.Root` to prevent symlink traversal outside declared roots.
- Symlink copying and special filesystem objects are unsupported.
- UndoLang does not claim atomic visibility or universal rename/durability behavior across platforms.

## Verification

```sh
GOPROXY=off go test ./...
GOPROXY=off go vet ./...
GOPROXY=off go test -race ./...
go list -m all
```
