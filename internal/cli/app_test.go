package cli

import (
	"bytes"
	json "encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func invoke(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := (App{Stdout: &stdout, Stderr: &stderr}).Run(args)
	return code, stdout.String(), stderr.String()
}

func TestCheckJSONAndInterspersedFlags(t *testing.T) {
	root := t.TempDir()
	scriptDir := t.TempDir()
	script := filepath.Join(scriptDir, "test.undo")
	if err := os.WriteFile(script, []byte(`transaction "a" { mkdir "new" } transaction "b" {}`), 0o600); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := invoke("check", script, "--root", root, "--json")
	if code != 0 || stderr != "" {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	var envelope struct {
		APIVersion string `json:"api_version"`
		OK         bool   `json:"ok"`
		Result     struct {
			Transactions []string `json:"transactions"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatal(err)
	}
	if !envelope.OK || envelope.APIVersion != "undo-cli/1" || len(envelope.Result.Transactions) != 2 {
		t.Fatalf("envelope=%#v", envelope)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("check mutated root: %v", entries)
	}
}

func TestPlanJSONNoMutation(t *testing.T) {
	root := t.TempDir()
	script := filepath.Join(t.TempDir(), "test.undo")
	if err := os.WriteFile(filepath.Join(root, "source"), []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(script, []byte(`transaction "x" { copy "source" -> "target" }`), 0o600); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := invoke("plan", "--json", script, "--root", root)
	if code != 0 || stderr != "" {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	var envelope map[string]any
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope["ok"] != true {
		t.Fatalf("output=%v", envelope)
	}
	if _, err := os.Stat(filepath.Join(root, "target")); !os.IsNotExist(err) {
		t.Fatal("plan mutated target")
	}
}

func TestAgentContracts(t *testing.T) {
	for _, command := range []string{"capabilities", "schema", "version"} {
		code, stdout, stderr := invoke(command, "--json")
		if code != 0 || stderr != "" {
			t.Fatalf("%s code=%d stderr=%s", command, code, stderr)
		}
		var envelope map[string]any
		if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
			t.Fatalf("%s: %v", command, err)
		}
		if envelope["api_version"] != "undo-cli/1" || envelope["ok"] != true {
			t.Fatalf("%s: %v", command, envelope)
		}
	}
}

func TestJSONErrorDoesNotPolluteStdout(t *testing.T) {
	code, stdout, stderr := invoke("check", "missing.undo", "--json")
	if code == 0 || stderr != "" {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	var envelope map[string]any
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope["ok"] != false {
		t.Fatalf("%v", envelope)
	}
}
