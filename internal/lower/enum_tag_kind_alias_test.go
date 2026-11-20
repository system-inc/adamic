package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// #s8gy89r U7: a switch over a const alias of the reference's field is that field's switch. Each
// program uses kindPredicateTypes, whose F is a one-member enum, so A's flag is an open tag B lacks
// and every narrowed read of N reaches the walk of enum_tag_flow.go.
const kindAliasValues = "const a: A = { kind: K.A, flag: F.X, a: 1 };\nconst b: B = { kind: K.B, b: 'x' };\n"

// kindAliasRecognized asks kindAliasField directly of the last switch over kind in source, with the
// reference the first n.a read inside it, so no later lowering stop can hide the recognizer's answer.
func kindAliasRecognized(t *testing.T, source string) string {
	t.Helper()
	l, statements := enumGuardProof(t, source)
	var statement, reference *ast.Node
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindSwitchStatement {
			if switched := ast.SkipParentheses(n.AsSwitchStatement().Expression); switched.Kind == ast.KindIdentifier && switched.Text() == "kind" {
				statement, reference = n, nil
			}
		}
		if statement != nil && reference == nil && n.Kind == ast.KindPropertyAccessExpression && n.Name().Text() == "a" && n.Expression().Kind == ast.KindIdentifier && n.Pos() >= statement.Pos() && n.End() <= statement.End() {
			reference = n.Expression()
		}
		n.ForEachChild(visit)
		return false
	}
	for _, n := range statements {
		visit(n)
	}
	if statement == nil || reference == nil {
		t.Fatal("missing alias switch or its read")
	}
	walk := &enumTagWalk{l: l, reference: reference, seen: map[*ast.FlowNode]bool{}, firstProof: -1, exempt: map[*ast.Node]bool{}}
	return walk.kindAliasField(statement)
}

// tsc's checkExpressionWorker in miniature: the alias, an unrelated call in a switch over it, then the
// switch whose cases read the narrowed node. .ts and .a both lower it and agree with Node.
func TestKindAliasWorkerAgreesWithNode(t *testing.T) {
	t.Parallel()
	const source = kindPredicateTypes + `let checks = 0;
function cancelled(): void { checks++; }
function check(n: N, token: boolean): string {
 const kind = n.kind;
 if (token) { switch (kind) { case K.A: cancelled(); } }
 switch (kind) { case K.A: return ` + "`a ${n.a + 1}`; default: return `b ${n.b}`;" + ` }
}
` + kindAliasValues + "console.log(check(a, true)); console.log(check(b, true)); console.log(String(checks));\n"
	if field := kindAliasRecognized(t, source); field != "kind" {
		t.Fatalf("kind-alias worker: recognized %q", field)
	}
	for _, extension := range []string{".ts", ".a"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, source, extension)
			if code, out, errs := brandedNative(t, program); code != 0 || out != "a 2\nb x\n1\n" {
				t.Fatalf("kind-alias worker %s: exit=%d stdout=%q stderr=%q", extension, code, out, errs)
			}
		})
	}
}

// Each row keeps the walk's answer of main: the recognizer reads no field, and a row the checker
// narrows by its alias stays refused as family 1. The checker narrows nothing by a let alias or by
// another reference's alias (TestKindAliasCheckerIgnoresLetAndOtherReference), so those two rows read
// n.a under their own comparison, which proves it on main and here alike.
func TestKindAliasFallbacksKeepFamilyOne(t *testing.T) {
	t.Parallel()
	const getters = `enum K { A = 1, B = 2 }
enum F { X = 1 }
class GA { get kind(): K.A { return K.A; } readonly flag: F = F.X; readonly a: number = 1 }
class GB { get kind(): K.B { return K.B; } readonly b: string = 'x' }
type N = GA | GB;
const a = new GA();
const b = new GB();
`
	const cases = "switch (kind) { case K.A: return `a ${n.a}`; default: return 'b'; }"
	for _, p := range []struct{ name, source, node string }{
		{"let", kindPredicateTypes + kindAliasValues + "function show(n: N): string { let kind = n.kind; switch (kind) { case K.A: return n.kind === K.A ? `a ${n.a}` : 'x'; default: return 'b'; } }\nconsole.log(show(a));\n", "a 1\n"},
		{"after", kindPredicateTypes + kindAliasValues + "function show(n: N): string { const read = (): string => { " + cases + " }; const kind = n.kind; return read(); }\nconsole.log(show(a));\n", ""},
		{"other_function", kindPredicateTypes + kindAliasValues + "function show(n: N): string { const kind = n.kind; const read = (): string => { " + cases + " }; return read(); }\nconsole.log(show(a));\n", ""},
		{"loop", kindPredicateTypes + kindAliasValues + "function show(n: N, m: N): string { const kind = n.kind; let out = ''; for (let i = 0; i < 2; i++) { switch (kind) { case K.A: out += `a ${n.a};`; } n = m; } return out; }\nconsole.log(show(a, b));\n", ""},
		{"other_reference", kindPredicateTypes + kindAliasValues + "function show(n: N, m: N): string { const kind = m.kind; switch (kind) { case K.A: return n.kind === K.A ? `a ${n.a}` : 'x'; default: return 'b'; } }\nconsole.log(show(a, a));\n", "a 1\n"},
		{"getter", getters + "function show(n: N): string { const kind = n.kind; " + cases + " }\nconsole.log(show(a)); console.log(show(b));\n", ""},
		{"assigned", kindPredicateTypes + kindAliasValues + "function show(n: N, m: N): string { const kind = n.kind; n = m; " + cases + " }\nconsole.log(show(a, b));\n", ""},
		{"destructured", kindPredicateTypes + kindAliasValues + "function show(n: N, m: N): string { const kind = n.kind; [n] = [m]; " + cases + " }\nconsole.log(show(a, b));\n", ""},
		{"shorthand", kindPredicateTypes + kindAliasValues + "function show(n: N, m: N): string { const kind = n.kind; ({ n } = { n: m }); " + cases + " }\nconsole.log(show(a, b));\n", ""},
	} {
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()
			if field := kindAliasRecognized(t, p.source); field != "" {
				t.Fatalf("kind-alias fallback %s was recognized as %q", p.name, field)
			}
			if p.node != "" {
				program := lowersAndAgreesWithNodeAs(t, p.source, ".ts")
				if code, out, errs := brandedNative(t, program); code != 0 || out != p.node {
					t.Fatalf("kind-alias %s own comparison: exit=%d stdout=%q stderr=%q", p.name, code, out, errs)
				}
				return
			}
			_, err := lowerSourceAs(t, p.source, ".ts")
			var refused *Refused
			if !errors.As(err, &refused) || refused.What != kindPredicateRefusal {
				t.Fatalf("kind-alias fallback %s must keep family 1: %v", p.name, err)
			}
		})
	}
}

// The checker itself narrows nothing by a let alias or by an alias of another reference
// (getCandidateDiscriminantPropertyAccess asks for a const whose initializer reads the reference), so a
// read only those could narrow never type-checks and never reaches the walk.
func TestKindAliasCheckerIgnoresLetAndOtherReference(t *testing.T) {
	t.Parallel()
	for _, p := range []struct{ name, show string }{
		{"let", "function show(n: N): string { let kind = n.kind; switch (kind) { case K.A: return `a ${n.a}`; default: return 'b'; } }\n"},
		{"other_reference", "function show(n: N, m: N): string { const kind = m.kind; switch (kind) { case K.A: return `a ${n.a}`; default: return 'b'; } }\n"},
	} {
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()
			_, err := load.Load([]string{writeAgreementSource(t, "main.ts", kindPredicateTypes+p.show)})
			if err == nil || !strings.Contains(err.Error(), "error TS2339: Property 'a' does not exist on type 'N'.") {
				t.Fatalf("kind-alias %s: the checker narrowed by it: %v", p.name, err)
			}
		})
	}
}

// Rule 3350 still judges every read: a module-level let written by a function called between the alias
// and the read is stopped with 3350's words, as it is under a direct switch, though the alias itself is
// recognized.
func TestKindAliasKeepsRule3350(t *testing.T) {
	t.Parallel()
	const prefix = kindPredicateTypes + kindAliasValues + "function first(): N { return a; }\nlet current: N = first();\nfunction swap(): void { current = b; }\n"
	source := prefix + "function show(): string { const kind = current.kind; swap(); switch (kind) { case K.A: return `a ${current.a}`; default: return 'b'; } }\nconsole.log(show());\n"
	if observed := runAgreementNode(t, writeAgreementSource(t, "main.ts", source)); observed.code != 0 || string(observed.stdout) != "a undefined\n" {
		t.Fatalf("kind-alias 3350 Node: %+v", observed)
	}
	if field := kindAliasRecognized(t, source); field != "kind" {
		t.Fatalf("kind-alias 3350 row: recognized %q", field)
	}
	_, err := lowerSourceAs(t, source, ".ts")
	var stop *NotYet
	if !errors.As(err, &stop) || stop.What != "a read of current narrowed to A where a call may have rewritten it (current is assigned inside swap)" {
		t.Fatalf("kind-alias rule 3350: %v", err)
	}
}
