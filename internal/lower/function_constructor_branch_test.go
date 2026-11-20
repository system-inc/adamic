package lower

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// functionConstructorVerbose is tsc's node constructor writing end on one path, small (#837zjkt's slice 2): a
// module let it reads, set to verbose, and a body writing kind and pos on every path and end only when verbose.
func functionConstructorVerbose(verbose string) string {
	return "let verbose = " + verbose + ";\n" +
		"function NodeObject(this: Nd, kind: number, pos: number, end: number): void {\n" +
		"    this.kind = kind;\n" +
		"    this.pos = pos;\n" +
		"    if (verbose) {\n" +
		"        this.end = end;\n" +
		"    }\n" +
		"}\n"
}

// A function constructor writing a field on one path is eligible (#837zjkt's slice 2), and its window lays every
// field Absent (#w8nanzw (B)). With verbose set, the path writes end, and the program agrees with Node on the
// JavaScript backend and natively.
func TestFunctionConstructorOnePathWrittenAgrees(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNodeAs(t, functionConstructorProgram(functionConstructorVerbose("true"), "", ""), ".ts")
	code, out, errs := brandedNative(t, program)
	if code != 0 || out != "7 -1 -1\n9\n" {
		t.Fatalf("native: exit %d stdout %q stderr %q; want exit 0 and Node's %q", code, out, errs, "7 -1 -1\n9\n")
	}
}

// With verbose unset, the path skips end. Node reads it as undefined; Adamic's object is born a checked view of
// Nd over a slot left Absent, so the read of end stops with exit 70, before the line that reads it prints, on
// both backends, as the lie does (TestFunctionConstructorLieStopsAtTheRead).
// Mutant: windowLayout keeping the fields present lets the read print end's zero (function_constructor.mutants.md).
func TestFunctionConstructorOnePathSkippedStopsAtTheRead(t *testing.T) {
	t.Parallel()
	path := writeAgreementSource(t, "main.ts", functionConstructorProgram(functionConstructorVerbose("false"), "", ""))
	node := runAgreementNode(t, path)
	if node.code != 0 || string(node.stdout) != "7 -1 undefined\n9\n" {
		t.Fatalf("source Node: exit %d stdout %q, want exit 0 and the skipped read", node.code, node.stdout)
	}
	program := lowerAgreementEntry(t, path)
	generated := filepath.Join(t.TempDir(), "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	backend := runAgreementNode(t, generated)
	if backend.code != 70 || string(backend.stdout) != "" || !strings.HasPrefix(string(backend.stderr), "adamic: panic: ") {
		t.Fatalf("JavaScript backend: exit %d stdout %q stderr %q; want exit 70 at the checked read of end, before anything prints", backend.code, backend.stdout, backend.stderr)
	}
	code, out, errs := brandedNative(t, program)
	if code != 70 || out != "" || !strings.HasPrefix(errs, "adamic: panic: ") {
		t.Fatalf("native: exit %d stdout %q stderr %q; want exit 70 at the checked read of end, before anything prints", code, out, errs)
	}
}

// functionConstructorUnread is functionConstructorProgram with its first line never reading end.
func functionConstructorUnread(constructor, extra, tail string) string {
	return strings.Replace(functionConstructorProgram(constructor, extra, tail), " + ' ' + String(first.end)", "", 1)
}

// With verbose unset and end never read, the skipped slot is invisible as Node's is: Object.keys reads kind and
// pos, and in doesn't see end, on the JavaScript backend and natively.
// Mutant: windowLayout keeping the fields present lets keys and in see end (function_constructor.mutants.md).
func TestFunctionConstructorOnePathSkippedKeysAgree(t *testing.T) {
	t.Parallel()
	tail := "console.log(Object.keys(first).join(','));\n" +
		"console.log(String('end' in first));\n"
	const want = "7 -1\n9\nkind,pos\nfalse\n"
	program := lowersAndAgreesWithNodeAs(t, functionConstructorUnread(functionConstructorVerbose("false"), "", tail), ".ts")
	code, out, errs := brandedNative(t, program)
	if code != 0 || out != want {
		t.Fatalf("native: exit %d stdout %q stderr %q; want exit 0 and Node's %q", code, out, errs, want)
	}
}

// tsc's Signature shape, small: an if whose two branches write different fields, in different orders. Each
// instance's Object.keys, JSON.stringify and in read the order its own path's writes ran in, V8's order, on the
// JavaScript backend and natively. An if writing the same field on both sides, after fields every path writes
// in the layout's order, keeps the layout present and agrees too.
// Mutant: windowLayout keeping the fields present turns the different fields row red (function_constructor.mutants.md).
func TestFunctionConstructorBranchesAgree(t *testing.T) {
	t.Parallel()
	for _, row := range []struct {
		name, source, want string
	}{
		{"different fields on each side", typeConstructorProgram("function TypeObject(this: Tp, flags: number, checker: number): void {\n    if (flags > 1) {\n        this.checker = checker;\n        this.flags = flags;\n    } else {\n        this.flags = flags;\n        this.id = 0;\n    }\n}\n", "", "const checked = createType(3);\nconst plain = createType(0);\nconsole.log(Object.keys(checked).join(',') + ' ' + Object.keys(plain).join(','));\nconsole.log(String(JSON.stringify(checked)) + ' ' + String(JSON.stringify(plain)));\nconsole.log(String('id' in checked) + ' ' + String('checker' in plain) + ' ' + String(plain.flags));\n"), "checker,flags flags,id\n{\"checker\":7,\"flags\":3} {\"flags\":0,\"id\":0}\nfalse false 0\n"},
		{"one field on both sides", functionConstructorProgram("function NodeObject(this: Nd, kind: number, pos: number, end: number): void {\n    this.pos = pos;\n    this.end = end;\n    if (kind > 7) this.kind = kind;\n    else {\n        this.kind = 0;\n    }\n}\n", "", "console.log(Object.keys(first).join(','));\n"), "0 -1 -1\n9\npos,end,kind\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, row.source, ".ts")
			code, out, errs := brandedNative(t, program)
			if code != 0 || out != row.want {
				t.Fatalf("native: exit %d stdout %q stderr %q; want exit 0 and Node's %q", code, out, errs, row.want)
			}
		})
	}
}

// Language's rows (#837zjkt's slice 2): a write inside an expression, a comma's, a chained assignment's, a
// logical's or a logical assignment's, stays refused, and so does an if whose condition reads this, so each
// getter's cast keeps its refusal.
func TestFunctionConstructorBranchRoadsStop(t *testing.T) {
	t.Parallel()
	const refused = "a cast the runtime can't check"
	constructor := func(body string) string {
		return "let verbose = true;\nfunction NodeObject(this: Nd, kind: number, pos: number, end: number): void {\n" + body + "}\n"
	}
	for _, probe := range []struct{ name, constructor string }{
		{"w01 a write inside a comma", constructor("    this.kind = (this.pos = pos, kind);\n    this.end = end;\n")},
		{"w02 a chained write", constructor("    this.kind = this.pos = kind;\n    this.end = end;\n")},
		{"w03 a write inside a logical", constructor("    this.kind = kind;\n    this.pos = pos;\n    verbose && (this.end = end);\n")},
		{"w04 a logical assignment", constructor("    this.kind = kind;\n    this.pos = pos;\n    this.end ??= end;\n")},
		{"a condition reading this", constructor("    this.kind = kind;\n    this.pos = pos;\n    if (this.kind > 0) {\n        this.end = end;\n    }\n")},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, functionConstructorProgram(probe.constructor, "", ""), ".ts")
			if err == nil || !strings.Contains(err.Error(), refused) {
				t.Fatalf("want a stop holding %q, got %v", refused, err)
			}
		})
	}
}

// The region seam (WebAssembly's audit): the native emitter hands a region every field value of a class's own
// constructor emitted inside one, Absent ones included (internal/native/emit_objects.go). A function constructor
// whose window skips end on one path, made in F_new's region version and handed to a consumer that keeps nothing,
// agrees with Node and lets go of everything it made, under the sanitizer and the leak check.
func TestFunctionConstructorRegionAbsentField(t *testing.T) {
	t.Parallel()
	extra := "function weigh(node: Nd): number {\n    return node.kind + node.pos;\n}\n"
	tail := "console.log(String(weigh(createBaseNode(3))));\n" +
		"console.log(Object.keys(createBaseNode(4)).join(','));\n"
	program := lowersAndAgreesWithNodeAs(t, functionConstructorUnread(functionConstructorVerbose("false"), extra, tail), ".ts")
	code := native.C(program)
	version := regexp.MustCompile(`adamic_function_\w*NodeObject_allocate_[0-9a-f]{32}_in\b`)
	for _, line := range strings.Split(code, "\n") {
		if version.MatchString(line) || strings.Contains(line, "adamic_region ") {
			t.Logf("C: %s", strings.TrimSpace(line))
		}
	}
	binary := filepath.Join(t.TempDir(), "native")
	if err := native.Build(code, binary, native_runtime.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
	var out, errs bytes.Buffer
	command.Stdout, command.Stderr = &out, &errs
	if err := command.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatal(err)
		}
	}
	const want = "7 -1\n9\n2\nkind,pos\n"
	if out.String() != want || errs.Len() != 0 {
		t.Fatalf("native: stdout %q stderr %q; want Node's %q", out.String(), errs.String(), want)
	}
	if report := leakcheck.Report(t, code, binary); report != "" {
		t.Fatal(report)
	}
}
