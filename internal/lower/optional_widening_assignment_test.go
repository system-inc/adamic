package lower

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

const optionalAssignmentSys = `interface Levels { Low: number; Medium: number; High: number }
function getCustomLevels(): Partial<Levels> | undefined {
let customLevels: Partial<Levels> | undefined;
setCustomLevel('Low'); return customLevels;
function setCustomLevel(level: keyof Levels): void {
(customLevels || (customLevels = {}))[level] = 1;
}
}
console.log(String(getCustomLevels()?.Low));
`

// The reader clears sys.ts's false refusal. The element write then lowers as a finite keyed write
// (#13p3y0c), whose arm meets the absent optional field {} left out, a separate gap, so require both
// this rule's nil and the exact next whole stop.
func TestOptionalAssignmentSysBehindElementWrite(t *testing.T) {
	t.Parallel()
	finder, _ := cycleIndexTypes(t, optionalAssignmentSys)
	assignments := 0
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken && ast.IsIdentifier(ast.SkipParentheses(node.AsBinaryExpression().Left)) {
			assignments++
			if err := finder.l.refuseOptionalWidening(node); err != nil {
				t.Fatal(err)
			}
		}
		return node.ForEachChild(visit)
	}
	finder.l.program.Files()[0].AsNode().ForEachChild(visit)
	if assignments != 1 {
		t.Fatalf("judged %d assignment handles, want 1", assignments)
	}
	for _, extension := range []string{".a", ".ts"} {
		_, err := lowerSourceAs(t, optionalAssignmentSys, extension)
		want := "main" + extension + ":6:1: stage 0 can't lower writing a possibly absent optional own field yet"
		var later *NotYet
		if !errors.As(err, &later) || !strings.HasSuffix(err.Error(), want) {
			t.Fatalf("got %v, want the next stop ending %q", err, want)
		}
	}
}

func TestOptionalAssignmentFreshHandleLowers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"undefined handle", "let c: { Low?: number } | undefined;\nconst p: { Low?: number } = (c = {});\nconsole.log(String(p.Low));\n"},
		{"nullish handle", "let c: { Low?: number } | null | undefined;\nconst p: { Low?: number } = (c = {});\nconsole.log(String(p.Low));\n"},
		{"nested handles", "interface L { Low?: number }\nlet a: L = {}; let b: L = {};\nconst p: L = (a = (b = {}));\nconsole.log(String(p.Low));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				lowersAndAgreesWithNodeAs(t, probe.source, extension)
			})
		}
	}
}

// Reassigning a primitive or a nullable constructor handle cannot change a field
// of the value just assigned. These are the boundary regressions from the first delta.
func TestOptionalAssignmentCopiedValuesLower(t *testing.T) {
	t.Parallel()
	constructor, err := os.ReadFile("../oracle/testdata/new_expression_class_cache_local.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct{ name, source string }{
		{"primitive literal", "let x: number = 0;\nconst p: 1 = (x = 1);\nx = 2; console.log(String(p));\n"},
		{"primitive union", "let x: string = '';\nconst p: 'a' | 'b' = (x = 'a');\nx = 'c'; console.log(p);\n"},
		{"defined return", "let x: string | undefined;\nfunction f(): string { return x = 'yes'; }\nconsole.log(f());\n"},
		{"constructor handle", string(constructor)},
		{"enum member result", "const enum Kind { Unknown = 0, Keyword = 128, Identifier = 80 }\nlet wide: Kind = Kind.Unknown;\nfunction keyword(): Kind.Keyword { return Kind.Keyword; }\nfunction assignKeyword(): Kind.Keyword { const value: Kind.Keyword = keyword(); return wide = value; }\nconsole.log(String(assignKeyword()));\nconsole.log(String(wide));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				lowersAndAgreesWithNodeAs(t, probe.source, extension)
			})
		}
	}
}

// Mutant run and restored: using the evolving let's declared any type makes
// 09_literal_cache.a stop at its assignment instead of agreeing with Node.
func TestOptionalAssignmentEvolvingLiteralCacheLowers(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../../stage3/fixtures/taste/09_literal_cache.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) { t.Parallel(); lowersAndAgreesWithNodeAs(t, string(source), extension) })
	}
}

const optionalAssignmentJoin = `const flag = Date.now() < 0;
let u;
const q: Partial<{ a: number }> = (u = {});
if (flag) u = { a: "x" };
u.a = "s";
console.log(typeof q.a);
`

// Language's join is held by H1, not by the evolving handle's assignment.
// Mutant run and restored: skipping a union member without the property makes
// this direct rule assertion return nil and loses the whole program's refusal.
func TestOptionalAssignmentEvolvingJoinRefused(t *testing.T) {
	t.Parallel()
	finder, _ := cycleIndexTypes(t, optionalAssignmentJoin)
	writes := 0
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken && node.AsBinaryExpression().Left.Kind == ast.KindPropertyAccessExpression {
			writes++
			err := finder.l.unionSlotWriteRefusal(node)
			want := "main.a:5:1: Adamic 0.1 refuses a write through a union to a slot {} doesn't name; narrow to one member before writing (union-write)"
			var refused *Refused
			if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
				t.Fatalf("H1 got %v, want %q", err, want)
			}
		}
		return node.ForEachChild(visit)
	}
	finder.l.program.Files()[0].AsNode().ForEachChild(visit)
	if writes != 1 {
		t.Fatalf("judged %d union writes, want 1", writes)
	}
	for _, extension := range []string{".a", ".ts"} {
		_, err := lowerSourceAs(t, optionalAssignmentJoin, extension)
		want := "main" + extension + ":5:1: Adamic 0.1 refuses a write through a union to a slot {} doesn't name; narrow to one member before writing (union-write)"
		var refused *Refused
		if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
			t.Fatalf("got %v, want %q", err, want)
		}
	}
}

func TestOptionalAssignmentEvolvingClosureRefused(t *testing.T) {
	t.Parallel()
	const source = `const flag = Date.now() < 0;
let type;
const q: Partial<{ x: number }> = (type = {});
if (flag) type = { x: 'x' };
const write = (): void => { type.x = 's'; };
write(); console.log(typeof q.x);
`
	for _, extension := range []string{".a", ".ts"} {
		_, err := lowerSourceAs(t, source, extension)
		want := "main" + extension + ":5:29: Adamic 0.1 refuses a write through a union to a slot {} doesn't name; narrow to one member before writing (union-write)"
		var refused *Refused
		if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
			t.Fatalf("got %v, want %q", err, want)
		}
	}
}

// An annotation, an initializer, or an assigned any cannot borrow the evolving
// let's known RHS type. Judge this reader directly behind unrelated any stops.
func TestOptionalAssignmentOpaqueHandlesNotYet(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"let type: any; const p: { value: string } = (type = { value: 'a' });",
		"function f(input: any): void { let type = input; const p: { value: string } = (type = { value: 'a' }); } f({});",
		"function f(input: any): void { let type; const p: { value: string } = (type = input); } f({});",
	} {
		finder, _ := cycleIndexTypes(t, source)
		assignments := 0
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
				assignments++
				err := finder.l.refuseOptionalWidening(node)
				var later *NotYet
				if !errors.As(err, &later) || later.What != "an assignment used as a value without a known target write type" {
					t.Fatalf("got %v, want opaque handle NotYet", err)
				}
			}
			return node.ForEachChild(visit)
		}
		finder.l.program.Files()[0].AsNode().ForEachChild(visit)
		if assignments != 1 {
			t.Fatalf("judged %d assignment handles, want 1", assignments)
		}
	}
}

// The evolving-any declaration itself has a separate lowering gap in this small
// shape. Judge the assignment directly to isolate the opaque handle's primitive road.
func TestOptionalAssignmentEvolvingPrimitiveHandle(t *testing.T) {
	t.Parallel()
	finder, _ := cycleIndexTypes(t, "let x; const p: 1 = (x = 1); console.log(String(p));")
	assignments := 0
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
			assignments++
			if err := finder.l.refuseOptionalWidening(node); err != nil {
				t.Fatal(err)
			}
		}
		return node.ForEachChild(visit)
	}
	finder.l.program.Files()[0].AsNode().ForEachChild(visit)
	if assignments != 1 {
		t.Fatalf("judged %d assignments, want 1", assignments)
	}
}

const optionalAssignmentInvariant = "which can write number | undefined where string | undefined is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"

// #7n03qzg: the assignment publishes through its handles, including the setter's
// write type and every chained target. The guarded reads keep these accepted by tsc.
func TestOptionalAssignmentUnsafeHandlesRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, where string }{
		{"adamic", "let x: { Low?: string } = {};\nconst p: Partial<{ Low: number }> = (x = {});\nx.Low = 's';\nif (p.Low !== undefined) console.log(p.Low.toFixed());\n", ":2:38:"},
		{"chain", "let b: { Low?: string } = {};\nlet a: { Low?: number } = {};\nconst p: Partial<{ Low: number }> = (a = b = {});\nb.Low = 's';\nif (p.Low !== undefined) console.log(p.Low.toFixed());\n", ":3:38:"},
		{"setter", "class Box {\nget f(): { Low?: number } { return {}; }\nset f(value: { Low?: string }) {}\n}\nconst box = new Box();\nconst p: Partial<{ Low: number }> = (box.f = {});\nconsole.log(String(p.Low));\n", ":6:38:"},
		{"field", "const box: { f: { Low?: string } } = { f: {} };\nconst p: Partial<{ Low: number }> = (box.f = {});\nconsole.log(String(p.Low));\n", ":2:38:"},
		{"array index", "const values: { Low?: string }[] = [{}];\nconst p: Partial<{ Low: number }> = (values[0] = {});\nconst handle = values[0]; if (handle !== undefined) handle.Low = 's';\nif (p.Low !== undefined) console.log(p.Low.toFixed());\n", ":2:38:"},
		{"record index", "const values: Record<string, { Low?: string }> = {};\nconst p: Partial<{ Low: number }> = (values['x'] = {});\nconsole.log(String(p.Low));\n", ":2:38:"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				want := "main" + extension + probe.where + " Adamic 0.1 refuses a value of type { Low?: string; } seen as Partial<{ Low: number; }>, " + optionalAssignmentInvariant
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
					t.Fatalf("got %v, want a refusal ending %q", err, want)
				}
			})
		}
	}
}

func TestOptionalAssignmentHiddenFieldAndLogicalControls(t *testing.T) {
	t.Parallel()
	for _, operator := range []string{"=", "||=", "??=", "&&="} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(operator+extension, func(t *testing.T) {
				t.Parallel()
				source := "interface L { a?: number }\nconst held = { a: 'x' }; const wide: {} = held; let t: {} = {};\nconst p: Partial<L> = (t " + operator + " wide);\nconsole.log(String(p.a));\n"
				_, err := lowerSourceAs(t, source, extension)
				want := "main" + extension + ":3:24: Adamic 0.1 refuses optional property a in Partial<L> absent from structural source {}, which can hide fields; declare a on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
					t.Fatalf("got %v, want a refusal ending %q", err, want)
				}
			})
		}
	}
}

func TestOptionalAssignmentDestructuringValueNotYet(t *testing.T) {
	t.Parallel()
	const source = "let inner: { Low?: string } = {};\nconst p: { a?: { Low?: number } } = ({ a: inner } = { a: {} });\ninner.Low = 's';\nif (p.a?.Low !== undefined) console.log(p.a.Low.toFixed());\n"
	for _, extension := range []string{".a", ".ts"} {
		_, err := lowerSourceAs(t, source, extension)
		want := "main" + extension + ":2:38: stage 0 can't lower an assignment to a destructuring pattern used as a value yet"
		var later *NotYet
		if !errors.As(err, &later) || !strings.HasSuffix(err.Error(), want) {
			t.Fatalf("got %v, want a NotYet ending %q", err, want)
		}
	}
}

// Mutants run and restored: judge = as its right operand, and judge only the
// outermost handle (ignoring inner assignment sites too). Both witnesses then
// lower. The empty-literal witnesses above lose their handle refusals as well,
// but their later optional-field writes independently remain NotYet. Explicit
// Low: undefined avoids that mask; readonly receiving views avoid an independent
// mutable-view refusal of the literal's narrower Low: undefined type.
func TestOptionalAssignmentUnmaskedHandleWitnesses(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, where string }{
		{"one handle", "let x: { Low?: string | undefined } = { Low: 'old' };\nconst p: Readonly<Partial<{ Low: number | undefined }>> = (x = { Low: undefined });\nx.Low = 's';\nif (p.Low !== undefined) console.log(p.Low.toFixed());\n", ":2:60:"},
		{"inner handle", "let b: { Low?: string | undefined } = { Low: 'old' };\nlet a: { readonly Low?: number | undefined } = { Low: 0 };\nconst p: Readonly<Partial<{ Low: number | undefined }>> = (a = b = { Low: undefined });\nb.Low = 's';\nif (p.Low !== undefined) console.log(p.Low.toFixed());\n", ":3:60:"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				want := "main" + extension + probe.where + " stage 0 can't lower an assignment used as a value whose target write type { Low?: string | undefined; } cannot be read as Readonly<Partial<{ Low: number | undefined; }>> yet"
				var later *NotYet
				if !errors.As(err, &later) || !strings.HasSuffix(err.Error(), want) {
					t.Fatalf("got %v, want a NotYet ending %q", err, want)
				}
			})
		}
	}
}
