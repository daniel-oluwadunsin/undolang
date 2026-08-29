package cli

import (
	"crypto/sha256"
	"encoding/hex"
	json "encoding/json/v2"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/daniel-oluwadunsin/undolang/internal/buildinfo"
	"github.com/daniel-oluwadunsin/undolang/internal/lang/diag"
	"github.com/daniel-oluwadunsin/undolang/internal/lang/frontend"
	"github.com/daniel-oluwadunsin/undolang/internal/lang/validate"
	"github.com/daniel-oluwadunsin/undolang/internal/pathcap"
	"github.com/daniel-oluwadunsin/undolang/internal/plan"
	"github.com/daniel-oluwadunsin/undolang/internal/report"
)

type App struct{ Stdout, Stderr io.Writer }

func (a App) Run(args []string) int {
	if a.Stdout == nil {
		a.Stdout = os.Stdout
	}
	if a.Stderr == nil {
		a.Stderr = os.Stderr
	}
	if len(args) == 0 {
		fmt.Fprintln(a.Stderr, usage)
		return 1
	}
	switch args[0] {
	case "check":
		return a.check(args[1:])
	case "plan":
		return a.plan(args[1:])
	case "version":
		return a.version(args[1:])
	case "capabilities":
		return a.capabilities(args[1:])
	case "schema":
		return a.schema(args[1:])
	case "agent-guide":
		fmt.Fprintln(a.Stdout, agentGuide)
		return 0
	case "help", "--help", "-h":
		fmt.Fprintln(a.Stdout, usage)
		return 0
	default:
		return a.failure(false, &report.Error{Code: "E_USAGE", Message: "unknown command " + args[0]}, 1)
	}
}

const usage = "usage: undo <check|plan|version|capabilities|schema|agent-guide> [options]"
const agentGuide = "Agent workflow: capabilities --json -> schema --json -> check FILE --json -> plan FILE --json. Mutation commands are intentionally unavailable until the transaction engine is wired."

type stringList []string

func (s *stringList) String() string     { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error { *s = append(*s, v); return nil }

type common struct {
	root, transaction string
	allowed           stringList
	json, noColor     bool
	file              string
}

func parseCommon(name string, args []string, transaction bool) (common, error) {
	var c common
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&c.root, "root", "", "transaction root")
	flags.Var(&c.allowed, "allow-path", "additional capability root")
	if transaction {
		flags.StringVar(&c.transaction, "transaction", "", "transaction name")
	}
	flags.BoolVar(&c.json, "json", false, "machine JSON")
	flags.BoolVar(&c.noColor, "no-color", false, "disable color")
	reordered, err := reorder(args, map[string]bool{"root": true, "allow-path": true, "transaction": transaction, "json": false, "no-color": false})
	if err != nil {
		return c, err
	}
	if err = flags.Parse(reordered); err != nil {
		return c, err
	}
	if flags.NArg() != 1 {
		return c, fmt.Errorf("exactly one FILE is required")
	}
	c.file = flags.Arg(0)
	return c, nil
}

func reorder(args []string, values map[string]bool) ([]string, error) {
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			positional = append(positional, arg)
			continue
		}
		name := strings.TrimLeft(strings.SplitN(arg, "=", 2)[0], "-")
		takesValue, known := values[name]
		if !known {
			flags = append(flags, arg)
			continue
		}
		flags = append(flags, arg)
		if takesValue && !strings.Contains(arg, "=") {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag --%s requires a value", name)
			}
			i++
			flags = append(flags, args[i])
		}
	}
	return append(flags, positional...), nil
}

func load(c common) (validate.Program, []byte, string, *pathcap.Set, error) {
	src, err := os.ReadFile(c.file)
	if err != nil {
		return validate.Program{}, nil, "", nil, err
	}
	program, diagnostic := frontend.Parse(src)
	if diagnostic != nil {
		return validate.Program{}, nil, "", nil, diagnostic
	}
	abs, err := filepath.Abs(c.file)
	if err != nil {
		return validate.Program{}, nil, "", nil, err
	}
	caps, err := pathcap.Open(c.root, c.allowed)
	if err != nil {
		return validate.Program{}, nil, "", nil, err
	}
	return program, src, abs, caps, nil
}

func (a App) check(args []string) int {
	c, err := parseCommon("check", args, false)
	if err != nil {
		return a.err(jsonRequested(args), err)
	}
	program, _, path, caps, err := load(c)
	if err != nil {
		return a.err(c.json, err)
	}
	defer caps.Close()
	if err = plan.ValidatePaths(program, caps); err != nil {
		return a.err(c.json, err)
	}
	result := report.CheckResult{ScriptPath: path, Transactions: validate.Names(program)}
	if c.json {
		return a.success(result)
	}
	fmt.Fprintf(a.Stdout, "valid UndoLang program: %s\n", path)
	for _, name := range result.Transactions {
		fmt.Fprintf(a.Stdout, "transaction: %s\n", name)
	}
	return 0
}

func (a App) plan(args []string) int {
	c, err := parseCommon("plan", args, true)
	if err != nil {
		return a.err(jsonRequested(args), err)
	}
	program, src, path, caps, err := load(c)
	if err != nil {
		return a.err(c.json, err)
	}
	defer caps.Close()
	sum := sha256.Sum256(src)
	result, err := plan.Build(program, caps, plan.Options{ScriptPath: path, ScriptSHA256: hex.EncodeToString(sum[:]), SelectedName: c.transaction})
	if err != nil {
		return a.err(c.json, err)
	}
	if c.json {
		a.success(result)
		if !result.SafeToStart {
			return 2
		}
		return 0
	}
	fmt.Fprintf(a.Stdout, "program: %s\nroot: %s\nmode: %s\nsafe to start: %t\n", path, caps.Primary(), result.Mode, result.SafeToStart)
	for _, tx := range result.Transactions {
		fmt.Fprintf(a.Stdout, "transaction %s: %s\n", tx.Name, tx.Readiness)
		if tx.Plan != nil {
			for _, op := range tx.Plan.Operations {
				fmt.Fprintf(a.Stdout, "  %d. %s %s (%s)\n", op.Index+1, op.Kind, op.Target.Original, op.Effect)
			}
		}
	}
	if !result.SafeToStart {
		return 2
	}
	return 0
}

func jsonRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--json" || arg == "-json" || strings.HasPrefix(arg, "--json=") {
			return true
		}
	}
	return false
}

func boolFlag(args []string) (bool, error) {
	flags := flag.NewFlagSet("output", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	value := flags.Bool("json", false, "machine JSON")
	if err := flags.Parse(args); err != nil {
		return false, err
	}
	if flags.NArg() != 0 {
		return false, fmt.Errorf("unexpected arguments")
	}
	return *value, nil
}

func (a App) version(args []string) int {
	j, err := boolFlag(args)
	if err != nil {
		return a.err(false, err)
	}
	result := map[string]string{"version": buildinfo.Version, "dsl_version": buildinfo.DSLVersion, "api_version": buildinfo.APIVersion}
	if j {
		return a.success(result)
	}
	fmt.Fprintf(a.Stdout, "UndoLang %s\ndsl %s\napi %s\n", buildinfo.Version, buildinfo.DSLVersion, buildinfo.APIVersion)
	return 0
}

func (a App) capabilities(args []string) int {
	j, err := boolFlag(args)
	if err != nil {
		return a.err(false, err)
	}
	result := map[string]any{"cli_version": buildinfo.Version, "dsl_version": buildinfo.DSLVersion, "api_version": buildinfo.APIVersion, "operations": []string{"mkdir", "copy", "move", "write", "replace", "delete"}, "conditions": []string{"exists", "not_exists", "is_file", "is_dir", "contains", "sha256"}, "commands": []string{"check", "plan", "version", "capabilities", "schema", "agent-guide"}, "path_model": "relative paths bind to --root/cwd; external absolute paths require --allow-path"}
	if j {
		return a.success(result)
	}
	fmt.Fprintln(a.Stdout, "UndoLang capabilities: check and plan; six filesystem operations; six conditions")
	return 0
}

func (a App) schema(args []string) int {
	j, err := boolFlag(args)
	if err != nil {
		return a.err(false, err)
	}
	result := map[string]any{"dsl_version": buildinfo.DSLVersion, "transaction": "transaction STRING { require* mutation* assert* }", "operations": map[string]string{"mkdir": "mkdir PATH", "copy": "copy SOURCE -> TARGET [overwrite]", "move": "move SOURCE -> TARGET [overwrite]", "write": "write PATH = STRING", "replace": "replace PATH OLD -> NEW", "delete": "delete PATH"}, "conditions": map[string]string{"exists": "exists PATH", "not_exists": "not_exists PATH", "is_file": "is_file PATH", "is_dir": "is_dir PATH", "contains": "contains PATH TEXT", "sha256": "sha256 PATH = HEX"}}
	if j {
		return a.success(result)
	}
	fmt.Fprintln(a.Stdout, "UndoLang schema undo-dsl/1; use --json for the complete machine schema")
	return 0
}

func (a App) success(result any) int {
	data, err := json.Marshal(&report.Envelope{APIVersion: buildinfo.APIVersion, OK: true, Result: result})
	if err != nil {
		fmt.Fprintln(a.Stderr, err)
		return 7
	}
	fmt.Fprintln(a.Stdout, string(data))
	return 0
}

func (a App) err(jsonMode bool, err error) int {
	r := &report.Error{Code: "E_IO", Message: err.Error()}
	code := 3
	var diagnostic *diag.Error
	var pathError *pathcap.Error
	var planError *plan.Error
	switch {
	case errors.As(err, &diagnostic):
		r.Code, r.Message, r.Line, r.Column, code = diagnostic.Code, diagnostic.Message, diagnostic.Span.Start.Line, diagnostic.Span.Start.Column, 1
	case errors.As(err, &pathError):
		r.Code, r.Message, r.Path, r.Root, code = pathError.Code, pathError.Message, pathError.Path, pathError.Root, 3
	case errors.As(err, &planError):
		r.Code, r.Message, code = planError.Code, planError.Message, 2
	case errors.Is(err, os.ErrPermission):
		r.Code = "E_PERMISSION"
	}
	return a.failure(jsonMode, r, code)
}

func (a App) failure(jsonMode bool, result *report.Error, code int) int {
	if jsonMode {
		data, err := json.Marshal(&report.Envelope{APIVersion: buildinfo.APIVersion, OK: false, Error: result})
		if err == nil {
			fmt.Fprintln(a.Stdout, string(data))
			return code
		}
	}
	fmt.Fprintf(a.Stderr, "%s: %s\n", result.Code, result.Message)
	return code
}
