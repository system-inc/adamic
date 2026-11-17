package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// tsc's core.ts arrayFrom (:1337 at the stage 3 pin), its two overloads over one generic implementation
// that builds a fresh array in a for-of and returns it. Its body is certificate A's shape (#5fws6q8).
const overloadCertificateArrayFrom = `function arrayFrom<T, U>(iterator: Iterable<T>, map: (t: T) => U): U[];
function arrayFrom<T>(iterator: Iterable<T>): T[];
function arrayFrom<T, U>(iterator: Iterable<T>, map?: (t: T) => U): (T | U)[] {
    const result: (T | U)[] = [];
    for (const value of iterator) {
        result.push(map ? map(value) : value);
    }
    return result;
}
`

// overloadCertificateGateWords is what the gate says before an entry's form.
const overloadCertificateGateWords = "an overload certified by its body, entered other than by a direct call: "

// Both overloads called directly lower and agree with Node on the JavaScript backend and natively. The
// census accepts overload 1, U[] over (T | U)[], by certificate A where it refused it before
// (census_small.go's result refusal). Overload 2 it already accepted by relation, its implementation's
// extra binder U witnessed as T, so (T | T)[] is T[] (Correctness's scout lists only overload 1 refused).
// The iterable is a Map's values iterator, as tsc hands arrayFrom map.keys(): an Iterable<T> slot lowers
// only where every value reaching it is a collection's iterator (iterable_views.go).
func TestOverloadCertificateArrayFromLowers(t *testing.T) {
	t.Parallel()
	source := overloadCertificateArrayFrom + `const counts = new Map<string, number>([["a", 3], ["b", 1], ["c", 2]]);
const plain = arrayFrom(counts.values());
const doubled = arrayFrom(counts.values(), x => x * 2);
console.log(plain.join(","));
console.log(doubled.join(","));
`
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "3,1,2\n6,2,4\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want what Node prints", code, stdout, stderr)
	}
}

// Correctness's array negatives (stage3/census/overload-body-proof-scout/counterexamples on
// typescript/overload-core-evidence), each still refused at its declaration for its result. Her programs
// loop over a literal, which the certificate refuses before it reaches the hazard each one shows, so each
// is here twice: as she wrote it, and with only its loop moved onto a parameter, so the row stands on
// its own hazard. array-default already loops over a parameter. Her Map and fold negatives are
// certificates B and C's, not this one's.
func TestOverloadCertificateArrayNegativesStayRefused(t *testing.T) {
	t.Parallel()
	const observe = "console.log(`expected string, got ${typeof actual}`);\n"
	for _, probe := range []struct{ name, source string }{
		{"index write", `function bad(): string[];
function bad(): (string | number)[] {
    const out: (string | number)[] = [];
    for (const value of ["ok"]) out.push(value);
    out[0] = 0;
    return out;
}
const actual = bad()[0];
` + observe},
		{"index write over a parameter", `function bad(values: string[]): string[];
function bad(values: string[]): (string | number)[] {
    const out: (string | number)[] = [];
    for (const value of values) out.push(value);
    out[0] = 0;
    return out;
}
const actual = bad(["ok"])[0];
` + observe},
		{"alias", `function bad(): string[];
function bad(): (string | number)[] {
    const out: (string | number)[] = [];
    const alias = out;
    for (const value of ["ok"]) out.push(value);
    alias.push(0);
    return out;
}
const actual = bad()[1];
` + observe},
		{"alias over a parameter", `function bad(values: string[]): string[];
function bad(values: string[]): (string | number)[] {
    const out: (string | number)[] = [];
    const alias = out;
    for (const value of values) out.push(value);
    alias.push(0);
    return out;
}
const actual = bad(["ok"])[1];
` + observe},
		{"capture", `function bad(): string[];
function bad(): (string | number)[] {
    const out: (string | number)[] = [];
    const later = () => out.push(0);
    for (const value of ["ok"]) out.push(value);
    later();
    return out;
}
const actual = bad()[1];
` + observe},
		{"capture over a parameter", `function bad(values: string[]): string[];
function bad(values: string[]): (string | number)[] {
    const out: (string | number)[] = [];
    const later = () => out.push(0);
    for (const value of values) out.push(value);
    later();
    return out;
}
const actual = bad(["ok"])[1];
` + observe},
		{"getter", `function bad(): string[];
function bad(): (string | number)[] {
    const out: (string | number)[] = [];
    let reads = 0;
    const source = { get value(): string | number { return ++reads === 1 ? "ok" : 0; } };
    for (const unused of [0]) {
        if (typeof source.value === "string") out.push(source.value);
    }
    return out;
}
const actual = bad()[0];
` + observe},
		{"getter over a parameter", `function bad(values: number[]): string[];
function bad(values: number[]): (string | number)[] {
    const out: (string | number)[] = [];
    let reads = 0;
    const source = { get value(): string | number { return ++reads === 1 ? "ok" : 0; } };
    for (const unused of values) {
        if (typeof source.value === "string") out.push(source.value);
    }
    return out;
}
const actual = bad([0])[0];
` + observe},
		{"default", `function bad(values: number[], map?: (v: number) => string): string[];
function bad(values: number[], map: (v: number) => string | number = v => v): (string | number)[] {
    const out: (string | number)[] = [];
    for (const value of values) out.push(map(value));
    return out;
}
const actual = bad([0])[0];
` + observe},
		{"prototype", `Object.defineProperty(Array.prototype, "push", {
    value: function(this: (string | number)[]) { this[0] = 0; return 1; },
});
function bad(): string[];
function bad(): (string | number)[] {
    const out: (string | number)[] = [];
    for (const value of ["ok"]) out.push(value);
    return out;
}
const actual = bad()[0];
` + observe},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, "overload 1 of bad result string[] ") || !strings.Contains(refused.What, "cannot be served by implementation result (string | number)[]") {
				t.Fatalf("want overload 1 of bad refused at its declaration for its result, got %v", err)
			}
		})
	}
}

// Her prototype negative with its loop on a parameter passes the certificate, since its push is the
// library's own by every declaration, and the rewrite of Array.prototype's push that would break it
// stops where it's written, as the scout asks the certificate to keep (library_object.go).
func TestOverloadCertificateArrayPrototypeRewriteStaysRefused(t *testing.T) {
	t.Parallel()
	source := `Object.defineProperty(Array.prototype, "push", {
    value: function(this: (string | number)[]) { this[0] = 0; return 1; },
});
function bad(values: string[]): string[];
function bad(values: string[]): (string | number)[] {
    const out: (string | number)[] = [];
    for (const value of values) out.push(value);
    return out;
}
const actual = bad(["ok"])[0];
console.log(` + "`expected string, got ${typeof actual}`" + `);
`
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != "Object.defineProperty" {
		t.Fatalf("want the prototype rewrite refused where it's written, got %v", err)
	}
}

// overloadCertificateProbe loads modules from main.ts and hands back a lowering over main.ts's checker,
// the first call, new or tagged template in main.ts the checker resolves to an overload of arrayFrom,
// and that overload. The node is found by what it resolves to; the gate under test reads only symbols.
func overloadCertificateProbe(t *testing.T, modules map[string]string) (*lowering, *ast.Node, *ast.Node) {
	t.Helper()
	directory := t.TempDir()
	for name, source := range modules {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	var file *ast.SourceFile
	for _, candidate := range program.Files() {
		if strings.HasSuffix(program.FileName(candidate), "main.ts") {
			file = candidate
		}
	}
	if file == nil {
		t.Fatal("no main.ts among the program's files")
	}
	checked, release := program.Checker(context.Background(), file)
	t.Cleanup(release)
	l := &lowering{program: program, checker: checked, result: &ir.Program{}}
	var found, overload *ast.Node
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression || node.Kind == ast.KindTaggedTemplateExpression {
			if resolved := checked.GetResolvedSignature(node); resolved != nil {
				declaration := resolved.Declaration()
				if declaration != nil && declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() == nil && declaration.Name() != nil && declaration.Name().Text() == "arrayFrom" {
					found, overload = node, declaration
					return true
				}
			}
		}
		return node.ForEachChild(visit)
	}
	file.AsNode().ForEachChild(visit)
	if found == nil {
		t.Fatal("no entry into arrayFrom in main.ts")
	}
	return l, found, overload
}

// The gate keys erasure to the call: each row enters overload 1, the one certificate A accepts, and its
// entry is answered by overloadCertificateEntry on its own node. Lowered whole, a non-direct entry is
// NotYet at the call in the gate's words, naming its form, before its callee is lowered, so it never
// leans on overloadSetValueStop. A direct entry never says the gate's words; whether a direct row
// lowers past the gate is that road's own question, logged. tsc
// refuses (0, arrayFrom)(it) itself (a comma's left side with no effect, 2695, exempt only before an
// access), so the comma row's left side counts; and new of a function returning an array is tsc's
// error 2350, so there's no new row. The direct rows hand arrayFrom a Map's values iterator, the one
// iterable its Iterable<T> slot lowers; the non-direct rows keep an array, since the gate stops them
// before any argument is lowered. Mutant: overloadCertificateEntry answering true for every call
// expression, new and tagged template, and the comma, spread and tagged rows go red.
func TestOverloadCertificateArrayGate(t *testing.T) {
	t.Parallel()
	const items = "const items: number[] = [3, 1, 2];\n"
	const counts = "const counts = new Map<string, number>([[\"a\", 3], [\"b\", 1], [\"c\", 2]]);\n"
	exported := strings.ReplaceAll(overloadCertificateArrayFrom, "function arrayFrom", "export function arrayFrom")
	for _, probe := range []struct {
		name    string
		modules map[string]string
		direct  bool
		form    string
	}{
		{"parenthesized callee", map[string]string{"main.ts": overloadCertificateArrayFrom + counts + "const got = (arrayFrom)(counts.values(), x => x * 2);\nconsole.log(got.join(\",\"));\n"}, true, ""},
		{"optional call", map[string]string{"main.ts": overloadCertificateArrayFrom + counts + "const got = arrayFrom?.(counts.values(), x => x * 2);\nconsole.log(got.join(\",\"));\n"}, true, ""},
		{"import alias", map[string]string{"library.ts": exported, "main.ts": "import { arrayFrom as af } from \"./library\";\n" + counts + "const got = af(counts.values(), x => x * 2);\nconsole.log(got.join(\",\"));\n"}, true, ""},
		{"namespace member", map[string]string{"main.ts": "namespace ts {\n" + exported + "}\n" + counts + "const got = ts.arrayFrom(counts.values(), x => x * 2);\nconsole.log(got.join(\",\"));\n"}, true, ""},
		{"comma callee", map[string]string{"main.ts": overloadCertificateArrayFrom + items + "let touched = 0;\nconst got = (touched++, arrayFrom)(items, x => x * 2);\nconsole.log(got.join(\",\"));\n"}, false, "a comma expression callee"},
		{"spread argument", map[string]string{"main.ts": overloadCertificateArrayFrom + items + "const args: [number[], (x: number) => number] = [items, x => x * 2];\nconst got = arrayFrom(...args);\nconsole.log(got.join(\",\"));\n"}, false, "a spread argument"},
		{"tagged template", map[string]string{"main.ts": overloadCertificateArrayFrom + "const got = arrayFrom`a${(t: string) => t.length}`;\nconsole.log(got.join(\",\"));\n"}, false, "a tagged template"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			l, node, overload := overloadCertificateProbe(t, probe.modules)
			if erase := l.overloadCertificateEntry(node, overload); erase != probe.direct {
				t.Fatalf("overloadCertificateEntry: got %v, want %v", erase, probe.direct)
			}
			_, err := lowerModules(t, probe.modules)
			if probe.direct {
				if err != nil && strings.Contains(err.Error(), overloadCertificateGateWords) {
					t.Fatalf("a direct entry stopped at the gate: %v", err)
				}
				t.Logf("lowered whole: %v", err)
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != overloadCertificateGateWords+probe.form {
				t.Fatalf("want NotYet %q at the call, got %v", overloadCertificateGateWords+probe.form, err)
			}
		})
	}
}

// The decision is the call's, never the declaration's: asked of a node that isn't a call, a bare
// reference to the set, the gate answers false, whatever stops that reference elsewhere.
func TestOverloadCertificateArrayEntryRefusesANonCall(t *testing.T) {
	t.Parallel()
	l, call, overload := overloadCertificateProbe(t, map[string]string{"main.ts": overloadCertificateArrayFrom + "const items: number[] = [3, 1, 2];\nconst reference = arrayFrom;\nconsole.log(reference(items, x => x * 2).join(\",\"));\n"})
	if call.Kind != ast.KindCallExpression {
		t.Fatalf("want the call through reference, got %v", call.Kind)
	}
	var identifier *ast.Node
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclaration && ast.IsIdentifier(node.Name()) && node.Name().Text() == "reference" {
			identifier = node.Initializer()
			return true
		}
		return node.ForEachChild(visit)
	}
	ast.GetSourceFileOfNode(call).AsNode().ForEachChild(visit)
	if identifier == nil || !ast.IsIdentifier(identifier) {
		t.Fatal("no identifier reference to arrayFrom")
	}
	if l.overloadCertificateEntry(identifier, overload) {
		t.Fatal("overloadCertificateEntry erased a non-call use of the set")
	}
}
