# UndoLang

**Version 0.1.0 · Zero Dependency Hackathon, Track F (Open/Wildcard)**

UndoLang is a Go 1.27, standard-library-only filesystem transaction runtime and small `.undo` DSL. It makes multi-file automation inspectable before execution and recoverable after known failure or process interruption.

```text
PLAN -> LOCK -> PREPARE -> JOURNAL -> APPLY -> VERIFY -> COMMIT
                                             failure -> ROLLBACK
```

A `.undo` file is a program containing one or more named transactions. Each transaction is its own recoverable boundary. Whole-file execution is sequential and fail-fast: prior commits remain committed, the failing transaction rolls back, and later entries are skipped.

## Build

For a guided list of build, verification, and CLI commands:

```sh
make help
make build
```

The canonical one-command build remains:

```sh
go build -trimpath -buildvcs=false -o undo ./cmd/undo
./undo version
```

The resulting binary has no runtime package, database, daemon, cloud, API-key, or environment-variable requirement. `NO_COLOR` is optional. The module contains no `require` block and no `go.sum`.

For a pinned offline release build, run `./scripts/build.sh`. Release maintainers can use `./scripts/release.sh` to cross-build the documented targets and print their SHA-256 hashes. Generated binaries stay under ignored `dist/`.

Local convenience installers are available as `./install.sh` for macOS/Linux and `./install.ps1` for Windows. They copy an existing local binary without needing Go. When building this checkout, they automatically download the official Go 1.27.0 archive only if a compatible Go 1.27.x is not already available, verify its SHA-256 checksum, and keep it in a user-local UndoLang toolchain directory. Use `--no-install-go` (or `-NoInstallGo`) for an entirely offline install; neither installer needs administrator access.

## First program

```undolang
transaction "upgrade" {
  require contains "VERSION" "1"
  copy "release/app" -> "bin/app" overwrite
  write "VERSION" = "2"
  assert contains "VERSION" "2"
}
```

```sh
./undo check migration.undo --root /target
./undo plan migration.undo --root /target
./undo run migration.undo --root /target
./undo run migration.undo --root /target --yes --json
```

`run FILE` executes every transaction in source order. Add `--transaction NAME` to `plan` or `run` to select one. The entire source always validates before target mutation. Relative paths bind to `--root`, or invocation cwd when omitted; script location is independent. External absolute paths require repeatable `--allow-path` capabilities.

If execution is interrupted:

```sh
./undo recover --root /target --yes
./undo history --root /target
./undo inspect TXID --root /target --json
```

Recovery validates the framed CRC32C journal and derives rollback from durable records. It repairs only an incomplete final frame after a valid prefix. Corrupt or ambiguous state fails closed and retains backups.

## Commands

| Command | Purpose |
|---|---|
| `check FILE` | Validate the complete program and path policy without target mutation. |
| `plan FILE` | Report exact first/selected effects and deferred later readiness. |
| `run FILE` | Execute selected/all transactions with journaling, verification, and rollback. |
| `recover` | Resume rollback or reconcile a finalized stale lock. |
| `history` / `inspect TXID` | Read local transaction audit metadata and validated journals. |
| `capabilities --json` / `schema --json` | Discover the stable `undo-cli/1` agent contract. |

Noninteractive and JSON mutation requires `--yes`; `--json` alone never grants approval.

## Documentation and examples

- [Language reference](docs/LANGUAGE.md)
- [Programs, transactions, and recovery](docs/TRANSACTIONS.md)
- [Security](docs/SECURITY.md)
- [AI agent guide](docs/AGENTS_GUIDE.md)
- [Support matrix](docs/SUPPORT_MATRIX.md)
- [Examples](examples/)
- [Static marketing/docs site](marketing/index.html)
- [Standard-library substitutions](STDLIB.md)
- [Five-minute demo runbook](docs/DEMO_RUNBOOK.md)
- [Simple testing guide](TESTING.md)

## Why Track F

UndoLang combines a purpose-built language, capability sandbox, streaming file algorithms, binary recovery journal, and crash-restartable rollback—the sort of local systems runtime normally assembled from a CLI framework, parser generator, UUID package, database, transaction library, and file helpers. Track F fits because the product crosses parser, storage, developer-tooling, and security boundaries while demonstrating that Go 1.27's standard library is sufficient. The detailed substitutions are recorded in [STDLIB.md](STDLIB.md).

## Current limitations

- UndoLang does not claim isolation or atomic visibility across a sequence of filesystem operations.
- Platform rename and directory-sync durability differs. macOS arm64 is behavior-tested here; other release targets are cross-built only.
- One active transaction is allowed per primary root. Separate roots with overlapping external capabilities can race.
- Symlink entries may be moved or deleted and restored. Copying symlinks, dereferencing them for mutation, and special filesystem objects are unsupported.
- Basic modes are preserved where supported; ownership, ACLs, xattrs, timestamps, sparse allocation, resource forks, alternate streams, and hard-link identity are not.
- Recovery handles the unresolved transaction only and does not automatically resume later program entries.

## Verification

```sh
GOPROXY=off go test ./...
GOPROXY=off go vet ./...
GOPROXY=off go test -race ./...
go list -m all
./scripts/deps-proof.sh
./scripts/repro-build.sh
```

See `docs/IMPLEMENTATION_REPORT.md` for final evidence and platform qualifications.
