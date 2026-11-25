package lower

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// A parenthesized optional callee, `(x?.m)()`, ends the chain in JavaScript: the call calls what x?.m gives,
// undefined when x is missing, and throws a TypeError (#492fbpd, Seats' tail-i2 find, live on main). Every
// binder that guards an optional receiver asks the callee as written (optionalCallee), so none of these lowers
// as x?.m(), which skips the call and prints undefined. Each receiver is a property the checker narrowed to
// present and a call put back to missing, so the program type-checks as Seats' does. Each must stop: refused at
// compile time, or at exit 70 with nothing printed on the JavaScript backend and natively under the
// sanitizers, where Node throws. Mutants: optionalCallee taking the callee through its parentheses, as main
// did, and the string, map and join rows print undefined; class.go's stop dropped, and the class method row
// prints undefined. The RegExp row stops at main's own NotYet for an optional RegExp call either way. The
// spread and continuing rows (J-g1zw2ta) are a library method whose call optionalIntrinsic once declined
// before reaching its stop: mutant, the spread and continuing returns moved back above the stop, and push,
// unshift and join lower and dereference the missing array natively, and toFixed fails at clang.
func TestParenthesizedOptionalCalleeStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"string intrinsic", "interface Box { name: string | undefined }\nconst box: Box = { name: \"entry\" };\nfunction clear(): void { box.name = undefined; }\nif (box.name !== undefined) {\n    clear();\n    console.log(String((box.name?.toUpperCase)()));\n}\n"},
		{"regexp test", "interface Box { pattern: RegExp | undefined }\nconst box: Box = { pattern: /a/ };\nfunction clear(): void { box.pattern = undefined; }\nif (box.pattern !== undefined) {\n    clear();\n    console.log(String((box.pattern?.test)(\"abc\")));\n}\n"},
		{"map get", "interface Box { table: Map<string, number> | undefined }\nconst box: Box = { table: new Map([[\"k\", 1]]) };\nfunction clear(): void { box.table = undefined; }\nif (box.table !== undefined) {\n    clear();\n    console.log(String((box.table?.get)(\"k\")));\n}\n"},
		{"array join", "interface Box { items: string[] | undefined }\nconst box: Box = { items: [\"a\", \"b\"] };\nfunction clear(): void { box.items = undefined; }\nif (box.items !== undefined) {\n    clear();\n    console.log(String((box.items?.join)(\",\")));\n}\n"},
		{"class method", "class Counter { count(): number { return 1; } }\ninterface Box { counter: Counter | undefined }\nconst box: Box = { counter: new Counter() };\nfunction clear(): void { box.counter = undefined; }\nif (box.counter !== undefined) {\n    clear();\n    console.log(String((box.counter?.count)()));\n}\n"},
		{"array push spread", "interface Box { items: number[] | undefined }\nconst box: Box = { items: [1] };\nconst more: number[] = [2, 3];\nfunction clear(): void { box.items = undefined; }\nif (box.items !== undefined) {\n    clear();\n    console.log(String((box.items?.push)(...more)));\n}\n"},
		{"array unshift spread", "interface Box { items: number[] | undefined }\nconst box: Box = { items: [1] };\nconst more: number[] = [2, 3];\nfunction clear(): void { box.items = undefined; }\nif (box.items !== undefined) {\n    clear();\n    console.log(String((box.items?.unshift)(...more)));\n}\n"},
		{"array join continues", "interface Box { items: string[] | undefined }\nconst box: Box = { items: [\"a\", \"b\"] };\nfunction clear(): void { box.items = undefined; }\nif (box.items !== undefined) {\n    clear();\n    console.log(String((box.items?.join)(\",\").length));\n}\n"},
		{"number toFixed continues", "interface Box { n: number | undefined }\nconst box: Box = { n: 1.5 };\nfunction clear(): void { box.n = undefined; }\nif (box.n !== undefined) {\n    clear();\n    console.log((box.n?.toFixed)(1).length.toString());\n}\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				var refused *Refused
				var pending *NotYet
				if !errors.As(err, &refused) && !errors.As(err, &pending) {
					t.Fatalf("want a stop, refused or NotYet, got %v", err)
				}
				return
			}
			directory := t.TempDir()
			binary := filepath.Join(directory, "native")
			if err := native.Build(native.C(program), binary, native_runtime.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			runner, err := filepath.Abs("../../oracle/node.mjs")
			if err != nil {
				t.Fatal(err)
			}
			generated := filepath.Join(directory, "generated.mjs")
			if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
				t.Fatal(err)
			}
			for _, command := range []*exec.Cmd{exec.Command(binary), exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated)} {
				command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
				var out, errs bytes.Buffer
				command.Stdout, command.Stderr = &out, &errs
				code := 0
				if err := command.Run(); err != nil {
					var exit *exec.ExitError
					if !errors.As(err, &exit) {
						t.Fatal(err)
					}
					code = exit.ExitCode()
				}
				if code != 70 || out.Len() != 0 {
					t.Fatalf("%s: exit %d stdout %q stderr %q; want 70 and nothing printed, where Node throws", command.Path, code, out.String(), errs.String())
				}
			}
		})
	}
}

// The same calls written without parentheses are the optional roads optionalCallee keeps: each skips the
// call when the receiver is missing and agrees with Node on the JavaScript backend and natively under the
// sanitizers. An optional RegExp call is main's NotYet with or without parentheses, so it has no row here.
// Mutant: optionalCallee answering nil, and the rows lose their optional roads.
func TestOptionalCalleeWithoutParenthesesAgrees(t *testing.T) {
	t.Parallel()
	const source = "interface Box { name: string | undefined; items: string[] | undefined; table: Map<string, number> | undefined }\n" +
		"function show(box: Box): void {\n    console.log(String(box.name?.toUpperCase()));\n    console.log(String(box.items?.join(\",\")));\n    console.log(String(box.table?.get(\"k\")));\n}\n" +
		"show({ name: \"entry\", items: [\"a\", \"b\"], table: new Map([[\"k\", 1]]) });\n" +
		"show({ name: undefined, items: undefined, table: undefined });\n"
	const stdout = "ENTRY\na,b\n1\nundefined\nundefined\nundefined\n"
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, out, stderr := brandedNative(t, program); code != 0 || out != stdout {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, out, stderr, stdout)
	}
}
