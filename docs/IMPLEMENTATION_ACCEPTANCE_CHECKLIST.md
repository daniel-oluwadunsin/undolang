# UndoLang Final Acceptance Checklist

## Hackathon dependency compliance

- [ ] Go 1.27 pinned/documented.
- [ ] `go.mod` has no `require` block.
- [ ] No `go.sum` needed for runtime project.
- [ ] No `golang.org/x/...` imports.
- [ ] No vendored/copied third-party library source.
- [ ] Runtime invokes no external executable.
- [ ] `GOPROXY=off go test ./...` passes.
- [ ] `GOPROXY=off go build ...` passes.
- [ ] `go list -m all` shows only main module.

## DSL

- [ ] one or more named transactions/file.
- [ ] duplicate transaction names rejected.
- [ ] `run FILE` executes all in source order.
- [ ] `run FILE --transaction NAME` executes only selected transaction.
- [ ] entire source validates before any mutation.
- [ ] whole-program execution is fail-fast; prior commits persist; later transactions skip.
- [ ] UTF-8 lexer.
- [ ] quoted/raw strings.
- [ ] comments.
- [ ] line/column diagnostics.
- [ ] require phase.
- [ ] all six operations.
- [ ] assert phase.
- [ ] all six conditions.
- [ ] invalid/unknown syntax fails safely.
- [ ] no shell/exec/import/plugin path.

## Paths/security

- [ ] default root = cwd.
- [ ] `--root` works.
- [ ] script location independent.
- [ ] absolute path inside root works.
- [ ] external absolute path denied by default.
- [ ] `--allow-path` works.
- [ ] os.Root enforcement.
- [ ] `..` escape blocked.
- [ ] symlink escape blocked.
- [ ] `.undo` reserved.
- [ ] secret content not logged.

## Planner

- [ ] no mutations.
- [ ] operation effects accurate.
- [ ] conflicts detected.
- [ ] dangerous overwrite explicit.
- [ ] backup estimate.
- [ ] human output.
- [ ] JSON output/version.

## Transaction engine

- [ ] durable state directory.
- [ ] exclusive active lock.
- [ ] journal frames + CRC.
- [ ] journal sync ordering.
- [ ] backups before destructive changes.
- [ ] all real operations.
- [ ] large files streamed.
- [ ] preconditions revalidated.
- [ ] postconditions trigger rollback.
- [ ] real filesystem/I/O/verification error triggers rollback independent of assertions.
- [ ] operation error triggers rollback.
- [ ] runtime derives rollback/recovery; no user-authored rollback required.
- [ ] rollback reverse order.
- [ ] backup cleanup only after safe finalization.

## Crash recovery

- [ ] real child process kill tested.
- [ ] interrupted execute recovered.
- [ ] interrupted rollback recovered/idempotent.
- [ ] torn tail handled.
- [ ] mid-journal CRC corruption fails closed.
- [ ] ambiguous state keeps backups.
- [ ] unresolved lock blocks new run.

## CLI / AI agents

- [ ] check.
- [ ] plan.
- [ ] run.
- [ ] recover.
- [ ] history.
- [ ] inspect.
- [ ] capabilities --json.
- [ ] schema --json.
- [ ] agent-guide.
- [ ] stable error codes.
- [ ] documented exit codes.
- [ ] `--json` never polluted with prose.
- [ ] noninteractive mutation requires explicit approval.

## Distribution/docs

- [ ] static marketing folder, zero package dependencies.
- [ ] README matches implementation.
- [ ] language docs match parser.
- [ ] security/limitations honest.
- [ ] support matrix says tested vs cross-built.
- [ ] prebuilt release strategy documented.
- [ ] install.sh/install.ps1 convenience only.
- [ ] no runtime environment variables/API keys.

## Bonuses/submission

- [ ] STDLIB.md >=10 real substitutions.
- [ ] build twice byte-identical.
- [ ] both SHA-256 hashes published.
- [ ] deps-proof.txt current.
- [ ] reproducible-build.txt current.
- [ ] `.zero-dep.toml` Track F.
- [ ] public license.
- [ ] 5-minute real demo ready.
