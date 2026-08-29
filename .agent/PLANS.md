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

Phase 2 path capabilities and non-mutating condition evaluation.

## Completed

- Phase 0: repository/bootstrap plan recorded.
- Phase 1: Go 1.27 module, source spans, UTF-8 lexer, recursive-descent parser, ordered AST, semantic validator, spelling suggestions, and initial `check`/`version` entrypoint.

## In Progress

- Implement traversal-resistant roots, resolved path types, reserved-state rules, and streamed conditions.

## Next

Phase 3: immutable planner and production read-only CLI contracts.

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
- reproducible-build verification: not run
