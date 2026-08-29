package plan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daniel-oluwadunsin/undolang/internal/lang/frontend"
	"github.com/daniel-oluwadunsin/undolang/internal/pathcap"
)

func buildSource(t *testing.T, root, src, selected string) (ProgramPlan, error) {
	t.Helper()
	program, diagnostic := frontend.Parse([]byte(src))
	if diagnostic != nil {
		t.Fatal(diagnostic)
	}
	caps, err := pathcap.Open(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer caps.Close()
	return Build(program, caps, Options{ScriptPath: "test.undo", ScriptSHA256: "digest", SelectedName: selected})
}

func TestProgramPlanDefersLaterTransactions(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source"), []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	src := `transaction "first" { require exists "source" copy "source" -> "copy" }
transaction "second" { require exists "copy" delete "copy" }`
	result, err := buildSource(t, root, src, "")
	if err != nil {
		t.Fatal(err)
	}
	if !result.SafeToStart || len(result.Transactions) != 2 {
		t.Fatalf("result = %#v", result)
	}
	if result.Transactions[0].Readiness != Ready || result.Transactions[0].Plan.Operations[0].Effect != EffectCreate {
		t.Fatalf("first = %#v", result.Transactions[0])
	}
	if result.Transactions[1].Readiness != Deferred || result.Transactions[1].Operations[0].Effect != EffectDeferred {
		t.Fatalf("second = %#v", result.Transactions[1])
	}
	if _, err := os.Stat(filepath.Join(root, "copy")); !os.IsNotExist(err) {
		t.Fatal("planning mutated destination")
	}
}

func TestSelectedPlanAndPrecondition(t *testing.T) {
	root := t.TempDir()
	src := `transaction "a" {} transaction "b" { require not_exists "x" write "x" = "value" }`
	result, err := buildSource(t, root, src, "b")
	if err != nil {
		t.Fatal(err)
	}
	if result.Mode != "selected" || result.SelectedName != "b" || !result.SafeToStart {
		t.Fatalf("result = %#v", result)
	}
	if result.Transactions[0].Plan.Summary.Creates != 1 {
		t.Fatalf("summary = %#v", result.Transactions[0].Plan.Summary)
	}
}

func TestPlannerConflicts(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a"), []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []string{
		`transaction "x" { copy "a" -> "a" }`,
		`transaction "x" { copy "a" -> "b" copy "a" -> "b" }`,
		`transaction "x" { delete "missing" }`,
	}
	for _, src := range tests {
		if _, err := buildSource(t, root, src, ""); err == nil {
			t.Fatalf("expected conflict for %s", src)
		}
	}
}
