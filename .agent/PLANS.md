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

Phase 0 bootstrap for continuous implementation through Phase 5.

## Completed

None.

## In Progress

- Record the locked module path and phase sequence.
- Preserve user-authored instruction/prompt edits in the bootstrap commit.
- Implement and verify Phases 1 through 5 continuously, with a commit after each coherent milestone.

## Next

Phase 1: Go bootstrap and complete language front-end.

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
- `go test ./...`: not run
- `go vet ./...`: not run
- zero-dependency verification: not run
- reproducible-build verification: not run
