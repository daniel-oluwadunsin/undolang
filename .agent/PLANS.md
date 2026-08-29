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

Phase 10 adversarial security/correctness audit and final acceptance gates.

## Completed

- Phase 0: repository/bootstrap plan recorded.
- Phase 1: Go 1.27 module, source spans, UTF-8 lexer, recursive-descent parser, ordered AST, semantic validator, spelling suggestions, and initial `check`/`version` entrypoint.
- Phase 2: `os.Root` capability set, most-specific absolute mapping, reserved/escape policy, safe stat/open behavior, and bounded-memory contains/SHA-256 evaluation.
- Phase 3: exact/deferred program planner, conservative conflict analysis, rollback estimates, `undo-cli/1` JSON envelopes, and production read-only CLI commands.
- Phase 4: restrictive `.undo` layout, UUIDv7 metadata, exclusive active lock, validated state transitions, synced atomic status writes, CRC32C framed journal, strict replay, torn-tail detection, and history/inspection foundations.
- Phase 5: serializable prepare/apply/verify/undo metadata; verified backups; mkdir, copy, move, write, literal streaming replace, and delete primitives; complete-entry overwrite restoration; root-safe temporary installation; basic mode preservation; explicit symlink/special-file policies; and fail-closed inverse verification.
- Phase 6: crash-classifiable prepared metadata, strict journal state replay, source-order execution, post-lock revalidation, postconditions, journal-driven reverse rollback, idempotent recovery, real `run`/`recover`, and nine real-process kill/recover checkpoints.
- Phase 7: complete help and approval UX, whole-program JSON results, stable exit classes, history/inspect, expanded schema/capabilities, compiled-binary contract tests, and local shell/PowerShell installers.
- Phase 8: static newsprint landing/docs site, canonical runtime guides, five parser-validated examples, downloadable agent context, responsive navigation, and automated snippet/link/command correctness tests.
- Phase 9: version 0.1.0, MIT license, Track F metadata, 15-entry stdlib ledger, dependency receipts, byte-identical build proof, local installers, and five-target cross-build/checksum tooling.

## In Progress

- Audit hostile inputs, races, interrupted recovery, resource bounds, output disclosure, platform gaps, and all final acceptance evidence.

## Next

- Phase 10: adversarial audit and final acceptance gates.

## Decisions Made During Implementation

- Public module path: `github.com/daniel-oluwadunsin/undolang`.
- Go version: 1.27.0, standard library only, with no `require` block.
- The user explicitly authorized continuous work through Prompt 5, overriding per-prompt stop points.
- Prompt 5 ends at real reversible filesystem primitives; production `run`/recovery orchestration remains Phase 6.
- Quoted strings implement `\\uXXXX` and reject surrogate code points.
- Empty `contains` needles are semantic errors.
- Symlink copy is rejected in v1; symlink move/delete operate on the link entry.
- JSON uses Go 1.27 `encoding/json/v2`; transaction IDs use standard-library `uuid.NewV7`.
- The user authorized continuous implementation through Phase 10 with phase-specific verification and commits.
- Recovery classifies each prepared operation against durable before/after descriptors. Any state matching neither descriptor is ambiguous and fails closed with backups retained.
- A torn final journal frame may be truncated only to a fully validated prefix; complete-frame corruption is never repaired.
- Mutating JSON/noninteractive CLI use requires `--yes`; interactive confirmation is added in Phase 7.

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
- Phase 6 `GOPROXY=off GOTOOLCHAIN=go1.27.0 go test ./...`: pass.
- Phase 6 `GOPROXY=off GOTOOLCHAIN=go1.27.0 go vet ./...`: pass.
- Phase 6 `GOPROXY=off GOTOOLCHAIN=go1.27.0 go test -race ./...`: pass on macOS arm64.
- Phase 6 crash matrix: nine externally killed subprocess checkpoints recovered by a freshly built CLI process; pass on macOS arm64.
- Phase 6 module proof: only `github.com/daniel-oluwadunsin/undolang`.
- Phase 7 offline tests, vet, and race detector: pass on macOS arm64.
- Phase 7 compiled-binary help/version/schema/capabilities tests and run/history/inspect JSON tests: pass.
- Phase 7 installer shell syntax validation: pass; PowerShell execution is untested on this host.
- Phase 8 parser validation for every repository example and public UndoLang snippet: pass.
- Phase 8 documented-command help, local-site-link, and zero-external-asset tests: pass.
- Phase 8 desktop/mobile browser review, responsive navigation, and copy-to-agent interaction: pass.
- Phase 8 offline tests, vet, and race detector: pass on macOS arm64.
- Phase 8 module proof: only `github.com/daniel-oluwadunsin/undolang`.
- Phase 9 dependency proof: `go.mod` has no `require`, `go.sum` is absent, module list contains only the main module, and offline tests/build pass.
- Phase 9 reproducibility proof: two canonical macOS arm64 builds both hashed `c6b9e15fb6ff95aaa1a5c5b6ec42b88efcbde0496d43a2a3b079892738141343`.
- Phase 9 cross-builds: Linux amd64/arm64, macOS amd64/arm64, and Windows amd64 all pass with CGO disabled; only macOS arm64 behavior is claimed tested.
- Phase 9 offline tests, vet, and race detector: pass on macOS arm64.
