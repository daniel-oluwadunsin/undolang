package program

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/daniel-oluwadunsin/undolang/internal/lang/frontend"
	"github.com/daniel-oluwadunsin/undolang/internal/pathcap"
)

func parseProgram(t *testing.T, source string) interface{} {
	t.Helper()
	program, diagnostic := frontend.Parse([]byte(source))
	if diagnostic != nil {
		t.Fatal(diagnostic)
	}
	return program
}

func TestSourceOrderCommit(t *testing.T) {
	root := t.TempDir()
	caps, err := pathcap.Open(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer caps.Close()
	program, diagnostic := frontend.Parse([]byte(`
transaction "first" { write "one" = "1" assert contains "one" "1" }
transaction "second" { require exists "one" write "two" = "2" }
`))
	if diagnostic != nil {
		t.Fatal(diagnostic)
	}
	result, err := Run(program, caps, Options{ScriptPath: "test.undo", ScriptSHA256: "hash"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Transactions) != 2 || result.Transactions[0].Status != Committed || result.Transactions[1].Status != Committed {
		t.Fatalf("result=%#v", result)
	}
	for _, name := range []string{"one", "two"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLaterPreconditionFailureKeepsEarlierCommitAndSkipsLater(t *testing.T) {
	root := t.TempDir()
	caps, _ := pathcap.Open(root, nil)
	defer caps.Close()
	program, diagnostic := frontend.Parse([]byte(`
transaction "first" { write "committed" = "yes" }
transaction "second" { require exists "missing" write "never" = "x" }
transaction "third" { write "also-never" = "x" }
`))
	if diagnostic != nil {
		t.Fatal(diagnostic)
	}
	result, err := Run(program, caps, Options{ScriptPath: "test.undo", ScriptSHA256: "hash"})
	if err == nil {
		t.Fatal("expected precondition failure")
	}
	if got := result.Transactions; len(got) != 3 || got[0].Status != Committed || got[1].Status != FailedPreflight || got[2].Status != Skipped {
		t.Fatalf("result=%#v", result)
	}
	if _, err := os.Stat(filepath.Join(root, "committed")); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"never", "also-never"} {
		if _, err := os.Stat(filepath.Join(root, name)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("%s exists: %v", name, err)
		}
	}
}

func TestPostconditionFailureRollsBackCurrentTransaction(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "existing"), []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	caps, _ := pathcap.Open(root, nil)
	defer caps.Close()
	program, diagnostic := frontend.Parse([]byte(`transaction "bad" {
write "existing" = "after"
write "created" = "temporary"
assert contains "existing" "never"
}`))
	if diagnostic != nil {
		t.Fatal(diagnostic)
	}
	result, err := Run(program, caps, Options{ScriptPath: "test.undo", ScriptSHA256: "hash"})
	if err == nil || result.Transactions[0].Status != RolledBack {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	content, err := os.ReadFile(filepath.Join(root, "existing"))
	if err != nil || string(content) != "before" {
		t.Fatalf("existing=%q err=%v", content, err)
	}
	if _, err := os.Stat(filepath.Join(root, "created")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("created remains: %v", err)
	}
}

func TestSelectedTransactionOnly(t *testing.T) {
	root := t.TempDir()
	caps, _ := pathcap.Open(root, nil)
	defer caps.Close()
	program, diagnostic := frontend.Parse([]byte(`
transaction "one" { write "one" = "1" }
transaction "two" { write "two" = "2" }
`))
	if diagnostic != nil {
		t.Fatal(diagnostic)
	}
	result, err := Run(program, caps, Options{ScriptPath: "test.undo", ScriptSHA256: "hash", SelectedName: "two"})
	if err != nil || len(result.Transactions) != 1 || result.Transactions[0].Name != "two" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	if _, err := os.Stat(filepath.Join(root, "one")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unselected transaction ran: %v", err)
	}
}
