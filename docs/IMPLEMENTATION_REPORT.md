# UndoLang 0.1.0 Implementation Report

## What was built

UndoLang is a Go 1.27 standard-library-only filesystem transaction language and runtime. It includes a UTF-8 lexer, recursive-descent parser, semantic validator, capability-root path sandbox, streaming conditions, immutable planner, durable CRC32C journal, real reversible filesystem operations, source-order program runner, restartable rollback/recovery, versioned JSON CLI, static documentation site, and local release tooling.

The public commands are `check`, `plan`, `run`, `recover`, `history`, `inspect`, `version`, `capabilities`, `schema`, and `agent-guide`. Machine output uses `undo-cli/1`; language syntax uses `undo-dsl/1`.

## Architecture map

| Package | Ownership |
|---|---|
| `internal/lang/*` | Tokens, spans, lexer, parser, AST, and semantic validation |
| `internal/pathcap` | `os.Root` capability mapping and protected-path access |
| `internal/condition` | Streaming precondition/postcondition evaluation |
| `internal/plan` | Immutable current-state and deferred program plans |
| `internal/streamutil` | Bounded-memory contains, copy, and replace algorithms |
| `internal/fsop` | Prepare, apply, verify, classify, and inverse primitives |
| `internal/journal` | Framed append/sync/decode and semantic replay |
| `internal/state` | `.undo`, UUIDv7 metadata, statuses, history, and lock ownership |
| `internal/txn` | One transaction's durable execution lifecycle |
| `internal/program` | Selection and source-order fail-fast execution |
| `internal/recovery` | Journal-authoritative reverse rollback from a fresh process |
| `internal/report`, `internal/cli` | Stable human/JSON contracts and command UX |

## Build, install, and run

Source builds require Go 1.27.x:

```sh
go build -trimpath -buildvcs=false -o undo ./cmd/undo
./undo version
```

Use `./install.sh` on macOS/Linux or `./install.ps1` on Windows for local, non-administrator installation. Neither installer downloads anything. A prebuilt binary requires no Go installation.

Required runtime environment variables or API keys: none. `NO_COLOR` is optional. Release builds use only standard Go variables: `GOOS`, `GOARCH`, `CGO_ENABLED=0`, and the pinned `GOTOOLCHAIN=go1.27.0` selector.

Relative DSL paths resolve under `--root`, which defaults to the invocation working directory. Runtime state is stored at `<root>/.undo`; the DSL cannot address that reserved subtree. Additional absolute roots require explicit repeatable `--allow-path` arguments.

## Verification performed

The following gates pass on macOS arm64 with Go 1.27.0:

```sh
GOPROXY=off go test ./...
GOPROXY=off go vet ./...
GOPROXY=off go test -race ./...
go list -m all
UNDOLANG_STRESS=1 GOPROXY=off go test ./internal/fsop -run '^TestStressCopy256MiB$' -count=1
```

The test suite covers the parser and journal fuzz surfaces, path/symlink enforcement, planner conflicts, every filesystem operation and inverse, overwrite restoration, large/chunk-boundary streaming, structured JSON, history/inspection, failed assertions, and nine real subprocess kill points recovered by a newly started CLI process.

Five release binaries cross-build with `CGO_ENABLED=0`: Linux amd64/arm64, macOS amd64/arm64, and Windows amd64. Only macOS arm64 has behavioral evidence in this environment. See [SUPPORT_MATRIX.md](SUPPORT_MATRIX.md).

Dependency proof is in [`../deps-proof.txt`](../deps-proof.txt). Two canonical macOS arm64 builds produced the identical SHA-256 recorded in [`../reproducible-build.txt`](../reproducible-build.txt).

## Complete-flow test

```sh
go build -trimpath -buildvcs=false -o undo ./cmd/undo
./undo check examples/app-upgrade.undo --root /path/to/fixture
./undo plan examples/app-upgrade.undo --root /path/to/fixture
./undo run examples/app-upgrade.undo --root /path/to/fixture --yes
./undo history --root /path/to/fixture
```

If an execution process is interrupted, run `./undo recover --root /path/to/fixture --yes` from a fresh process. Never delete `.undo` to bypass an unresolved recovery condition.

## Known limitations

- Filesystem mutations are externally visible while a transaction runs; UndoLang provides rollback/recovery, not isolation or universal atomic visibility.
- Cross-platform rename replacement and directory-sync durability are filesystem/OS-qualified.
- Separate roots whose external allowed capabilities overlap can race.
- Symlink copy and special-file mutation are rejected. Symlink entries themselves can be moved/deleted and restored.
- Basic modes are preserved where supported. Ownership, ACLs, xattrs, timestamps, hard-link identity, sparse allocation, resource forks, and alternate streams are not preserved.
- Recovery handles one unresolved transaction and does not resume later program entries.
- Windows and Linux targets compile but have not run the behavioral or crash suite on a real host in this environment.

No implementation decision remains blocked on user input. The final adversarial-audit findings and platform qualifications are appended after Phase 10.
