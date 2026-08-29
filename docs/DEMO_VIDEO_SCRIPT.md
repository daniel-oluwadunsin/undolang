# UndoLang demo video script

This is a natural voiceover and recording plan for a **3:55–4:00 minute** submission
video. It starts at the landing page, explains the language, demonstrates the real
filesystem operations, and closes with the reproducible-build and standard-library
receipts required by the Zero Dependency Hackathon.

The commands below use a disposable `video-root` directory. Do not point the demo at
your home directory or an important project. Prepare the fixture before recording so
the terminal portion stays within the time budget.

## Before recording

1. Build the binary and confirm the working tree is clean enough for a demo:

   ```sh
   cd /Users/macbook/Desktop/hackathons/undolang
   make build
   export UNDO=/Users/macbook/Desktop/hackathons/undolang/dist/undo
   ```

2. Open the repository in the IDE, with `examples/fs.undo`,
   `go.mod`, `internal/`, `STDLIB.md`, and `Makefile` easy to reach.

3. Open a second terminal in the disposable example folder:

   ```sh
   cd /Users/macbook/Desktop/undolang-examples
   export UNDO=/Users/macbook/Desktop/hackathons/undolang/dist/undo
   export DEMO_ROOT="$PWD/video-root"
   mkdir -p "$DEMO_ROOT"
   ```

4. Have the transaction commands in the next section ready to paste. Their output is
   real; `|| true` only keeps the recording shell moving after an intentional failure.

## Timed recording and voiceover

### 0:00–0:20 — Start at the landing page

**Record:** Open `marketing/index.html` in a browser. Show the hero, then the
`PLAN → LOCK → PREPARE → JOURNAL → APPLY → VERIFY → COMMIT` diagram and the
zero-dependency claim.

**Say:**

> “Most file automation can leave a machine half changed when the fourth step fails.
> UndoLang is a small filesystem language and a crash-safe transaction runtime. It
> makes the plan visible first, journals destructive work, verifies the result, and
> either commits or rolls back. This is a local Go binary with no runtime packages,
> services, or API keys.”

### 0:20–0:45 — Walk through the documentation

**Record:** Open the static docs navigation. Visit Getting Started, Language,
Programs, Transactions and Recovery, Paths, AI Agents, and Security. Pause on the
support/limitations language rather than scrolling every page.

**Say:**

> “The documentation is part of the product, not a separate web application. These
> pages explain installation, the language, program-versus-transaction selection,
> capability roots, recovery, the agent JSON contract, and the limits. In particular,
> UndoLang does not pretend to provide isolation or universal atomic visibility.”

### 0:45–1:25 — Explain the syntax in the IDE

**Record:** Open `examples/fs.undo`. Highlight the transaction names,
then `require`, the mutation block, and `assert`. Scroll through `copy`, `replace`,
`write`, `move`, `delete`, `mkdir`, and the conditions. Show the nine transaction
names in the file.

**Say:**

> “An UndoLang file is a program containing one or more uniquely named transactions.
> A transaction has three strict phases: requirements, mutations, then assertions.
> Requirements are checked before anything changes. Mutations describe the forward
> work. Assertions describe the state that must be true afterward.
>
> “Paths are relative to the declared root, not the script’s directory. Strings can
> be quoted or raw backtick paths, and the front end reports UTF-8 source positions.
> The language deliberately stays small: create directories, copy, move, write,
> literal replace, delete, and inspect conditions such as `is_file`, `contains`, and
> `sha256`.
>
> “This one file also demonstrates selection. `run FILE` executes transactions in
> source order and stops at the first failure. `--transaction NAME` runs one named
> transaction, so the examples do not need separate files.”

### 1:25–1:50 — Show how it was built

**Record:** In the IDE, show the package tree: `lexer`, `parser`, `validate`,
`pathcap`, `plan`, `journal`, `fsop`, `txn`, and `recovery`. Briefly show `go.mod`,
then the terminal command `go list -m all` with its single module line.

**Say:**

> “The implementation is layered. A hand-written lexer and recursive-descent parser
> build the AST; semantic validation enforces the phase rules. Capability roots and
> `os.Root` protect path access. The planner is read-only. The journal is framed and
> checksummed. Filesystem operations use bounded streaming buffers, and transaction
> and recovery code rebuild their decisions from durable journal state.
>
> “Everything here is Go 1.27 standard library code. The module list has only the
> main module.”

### 1:50–2:55 — Run the real filesystem transactions

**Record:** Switch to the terminal in `~/Desktop/undolang-examples`. First show the
`check` output listing all nine transactions. Then paste the selected commands below
one at a time, keeping each result visible for a beat. Use `find "$DEMO_ROOT"` after
the successful sequence to show the files on disk.

```sh
"$UNDO" check fs.undo --root "$DEMO_ROOT"

"$UNDO" run fs.undo --root "$DEMO_ROOT" --transaction seed-fixture --yes
"$UNDO" run fs.undo --root "$DEMO_ROOT" --transaction first-commits --yes
"$UNDO" run fs.undo --root "$DEMO_ROOT" --transaction second-fails --yes || true
"$UNDO" run fs.undo --root "$DEMO_ROOT" --transaction upgrade-files --yes
"$UNDO" run fs.undo --root "$DEMO_ROOT" --transaction overwrite-current --yes
"$UNDO" run fs.undo --root "$DEMO_ROOT" --transaction overwrite-current --yes
"$UNDO" run fs.undo --root "$DEMO_ROOT" --transaction retire-old-layout --yes
"$UNDO" run fs.undo --root "$DEMO_ROOT" --transaction demonstrate-rollback --yes || true
"$UNDO" run fs.undo --root "$DEMO_ROOT" --transaction missing-source --yes || true

chmod 0555 "$DEMO_ROOT/read-only"
"$UNDO" run fs.undo --root "$DEMO_ROOT" --transaction permission-denied --yes || true
chmod 0755 "$DEMO_ROOT/read-only"

find "$DEMO_ROOT" -maxdepth 3 -print
"$UNDO" history --root "$DEMO_ROOT"
```

**Say:**

> “Now I’m using the same program against a disposable root. `seed-fixture` creates
> directories and files. `first-commits` proves an earlier transaction can commit.
> `second-fails` stops at its missing-file requirement, so its write never happens.
> `upgrade-files` copies and replaces content, and `overwrite-current` exercises
> backup-and-replace twice. `retire-old-layout` moves one file and deletes another.
>
> “The next transaction intentionally fails an assertion. UndoLang rolls its write
> back, so the configuration is restored. The missing-source transaction fails during
> preflight. Finally, a read-only directory produces a real permission failure. These
> are operating-system outcomes, not simulated success messages. `history` leaves an
> audit record for the work.”

### 2:55–3:15 — Show whole-program fail-fast and JSON

**Record:** Use a fresh disposable root. Do not delete an arbitrary path; reset only
`$DEMO_ROOT`, then run:

```sh
rm -rf "$DEMO_ROOT"
mkdir -p "$DEMO_ROOT"
"$UNDO" run fs.undo --root "$DEMO_ROOT" --yes --json
```

Zoom briefly on the ordered JSON entries: committed transactions first, the failing
transaction, and skipped entries after it. Show that
`first-transaction-committed.txt` remains.

**Say:**

> “With no selector, the entire file is a sequential, fail-fast program. The JSON
> result preserves every declared entry: commits stay committed, the failure is
> reported with a stable code, and later transactions are marked skipped. It is not an
> implicit super-transaction.”

### 3:15–3:35 — Reproducible-build bonus

**Record:** In the repository terminal, run:

```sh
make reproducible-build
```

Keep the labeled Build A and Build B SHA-256 lines and the byte-identical result on
screen.

**Say:**

> “For the reproducible-build bonus, this is one Make command. It builds the same
> pinned Go 1.27 artifact twice, hashes both binaries, and fails if the bytes differ.
> The two labeled hashes match exactly.”

### 3:35–3:55 — STDLIB receipt and close

**Record:** Open `STDLIB.md`. Scroll over the substitutions table, then briefly show
`go.mod` with no `require` block. End on the README or the landing-page closing
statement.

**Say:**

> “The other bonus receipt is `STDLIB.md`: real substitutions such as `flag.FlagSet`
> for a CLI framework, a hand-written parser instead of a parser generator,
> `os.Root` for path capabilities, streaming I/O for file utilities, the standard
> crypto and hash packages, and the testing and release tools. The manifest is empty.
>
> “UndoLang gives filesystem automation a visible plan, a durable journal, verified
> rollback, and explicit recovery—with no third-party runtime dependency.”

## Editing notes

- Keep the browser, IDE, and terminal captures at normal speed; trim only command
  wait time and repeated output. Do not speed up the spoken explanation.
- If the permission transaction is bypassed by a privileged account, show the command
  and say that the host’s privilege model bypassed mode bits; do not claim a failure
  that did not occur.
- The full crash-kill/recovery matrix is documented in
  [`DEMO_RUNBOOK.md`](DEMO_RUNBOOK.md). It is valuable evidence, but the four-minute
  cut above prioritizes the language, all filesystem operation scenarios, the
  fail-fast model, and both hackathon bonus receipts.
