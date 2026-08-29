# UndoLang Implementation Plan

This file is the living implementation plan for UndoLang.

Codex must:

- Read `AGENTS.md` first.
- Read `docs/HACKATHON_CONTEXT.md`.
- Read `docs/PROGRAM_EXECUTION_MODEL.md`.
- Read `docs/PRD.md`, `docs/DSL_LANGUAGE_SPEC.md`, and `docs/TECHNICAL_SPEC.md`.
- Update this plan before beginning each implementation phase.
- Mark completed work honestly.
- Record architectural decisions, blockers, unresolved questions, and verification results.
- Never introduce a third-party dependency.
- Never silently change product semantics.

## Current Phase

Phases 0-5 complete. Phase 6 is not authorized by this plan.

## Completed

- Phase 0: repository/bootstrap plan recorded.
- Phase 1: Go 1.27 module, source spans, UTF-8 lexer, recursive-descent parser, ordered AST, semantic validator, spelling suggestions, and initial `check`/`version` entrypoint.
- Phase 2: `os.Root` capability set, most-specific absolute mapping, reserved/escape policy, safe stat/open behavior, and bounded-memory contains/SHA-256 evaluation.
- Phase 3: exact/deferred program planner, conservative conflict analysis, rollback estimates, `undo-cli/1` JSON envelopes, and production read-only CLI commands.
- Phase 4: restrictive `.undo` layout, UUIDv7 metadata, exclusive active lock, validated state transitions, synced atomic status writes, CRC32C framed journal, strict replay, torn-tail detection, and history/inspection foundations.
- Phase 5: serializable prepare/apply/verify/undo metadata; verified backups; mkdir, copy, move, write, literal streaming replace, and delete primitives; complete-entry overwrite restoration; root-safe temporary installation; basic mode preservation; explicit symlink/special-file policies; and fail-closed inverse verification.

## In Progress

None.

## Next

Phase 6 (not authorized in this run): production transaction runner, rollback orchestration, and recovery.

## Decisions Made During Implementation

- Public module path: `github.com/daniel-oluwadunsin/undolang`.
- Go version: 1.27.0, standard library only, with no `require` block.
- The user explicitly authorized continuous work through Prompt 5, overriding per-prompt stop points.
- Prompt 5 ends at real reversible filesystem primitives; production `run`/recovery orchestration remains Phase 6.
- Quoted strings implement `\\uXXXX` and reject surrogate code points.
- Empty `contains` needles are semantic errors.
- Symlink copy is rejected in v1; symlink move/delete operate on the link entry.
- JSON uses Go 1.27 `encoding/json/v2`; transaction IDs use standard-library `uuid.NewV7`.

## Open Questions

None blocking Phases 0-5.

## Verification

- Go 1.27 toolchain availability: verified with `GOTOOLCHAIN=go1.27.0 go version`.
- Phase 1 `GOPROXY=off go test ./...`: pass.
- Phase 1 `GOPROXY=off go vet ./...`: pass.
- Phase 1 `GOPROXY=off go test -race ./...`: pass on macOS arm64.
- Phase 1 `go list -m all`: one module (`github.com/daniel-oluwadunsin/undolang`).
- Phase 2 offline tests, vet, and race detector: pass on macOS arm64.
- Phase 2 module proof: one main module.
- Phase 3 offline tests, vet, and race detector: pass on macOS arm64.
- Phase 3 `check`/`plan` mutation snapshots and structured JSON contract tests: pass.
- Phase 4 offline tests, vet, and race detector: pass on macOS arm64.
- Phase 4 journal corruption/torn-tail/reference tests and state lock lifecycle tests: pass.
- Phase 5 focused filesystem operation/inverse integration tests: pass on macOS arm64.
- `UNDOLANG_STRESS=1 GOPROXY=off GOTOOLCHAIN=go1.27.0 go test ./internal/fsop -run '^TestStressCopy256MiB$' -count=1`: pass.
- `GOPROXY=off GOTOOLCHAIN=go1.27.0 go test ./...`: pass.
- `GOPROXY=off GOTOOLCHAIN=go1.27.0 go vet ./...`: pass.
- `GOPROXY=off GOTOOLCHAIN=go1.27.0 go test -race ./...`: pass on macOS arm64.
- `CGO_ENABLED=0 GOOS=<target> GOARCH=amd64 GOPROXY=off GOTOOLCHAIN=go1.27.0 go test -run '^$' -exec=/usr/bin/true ./...` for linux, windows, and darwin: compilation/linking pass; foreign behavioral tests were not executed.
- `GOTOOLCHAIN=go1.27.0 go list -m all`: one module (`github.com/daniel-oluwadunsin/undolang`).
- Reproducible-build verification: deferred to the release phase.
