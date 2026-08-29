# UndoLang

UndoLang is a Go 1.27, standard-library-only filesystem transaction language. The current implementation covers the language front-end, traversal-resistant path capabilities, streamed conditions, non-mutating planning, the durable journal/state foundation, and internal reversible filesystem primitives. Transaction execution is intentionally not exposed until the durable runner and recovery engine are wired in Phase 6.

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
- Filesystem primitives remain internal and cannot be reached through a production CLI path before journal orchestration exists.
- Planning, conditions, and filesystem operations use `os.Root` capability handles; mutation parents are rechecked and symlink parents are rejected.
- Symlink entries may be moved or deleted and restored. Copying symlinks, dereferencing them for mutation, and special filesystem objects are unsupported.
- Regular-file contents are copied, searched, hashed, and replaced with bounded buffers. Recursive trees preserve basic permission modes and empty directories.
- Backups preserve contents, directories, symlink targets, and basic modes. ACLs, ownership, xattrs, sparse allocation, resource forks, alternate data streams, and hard-link identity are not preserved.
- Files and temporary replacements are synced before installation and containing directories are synced where the host supports it. UndoLang does not claim atomic visibility, universal rename semantics, or identical durability guarantees across operating systems and filesystems.
- Cross-capability moves use verified copy-then-delete. Same-capability rename falls back only for an explicitly recognized cross-device error; this fallback is compile-tested but requires a suitable multi-filesystem fixture for behavioral testing.

## Verification

```sh
GOPROXY=off go test ./...
GOPROXY=off go vet ./...
GOPROXY=off go test -race ./...
go list -m all
```
