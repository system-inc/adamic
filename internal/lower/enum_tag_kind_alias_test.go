package lower

import (
	"errors"
	"strconv"
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
			_, err := lowerSourceAs(t, p.source, ".a")
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

// A closure that rebinds the reference, called between the alias and the switch: no name in that range
// writes it, so the alias is recognized, and tsc narrows the reference by the switch. Language's a01
// declares the closure in a parameter default, and Node throws reading .a of the B; Optimizer's two
// declare it before the alias, over a parameter and over a let. Over kindPredicateTypes' open tag
// (default_parameter, parameter_open, let_open) each read reaches the walk; over the closed NA | NB
// (parameter, let) only 3350 is asked. Rule 3350 alone holds each, in .ts and .a, and the never-rebound
// control lowers and prints yes natively.
func TestKindAliasClosureRebindKeepsRule3350(t *testing.T) {
	t.Parallel()
	const types = `enum K { A = 1, B = 2, C = 3 }
interface Node { readonly kind: K }
interface NA extends Node { readonly kind: K.A; readonly a: string }
interface NB extends Node { readonly kind: K.B; readonly b: number }
const na: NA = { kind: K.A, a: 'yes' };
const nb: NB = { kind: K.B, b: 7 };
`
	const body = "const kind = node.kind; reset(); switch (kind) { case K.A: return node.a; default: return 'x'; } }\nconsole.log(f(na, nb));\n"
	const openBody = "const kind = node.kind; reset(); switch (kind) { case K.A: return node.a.toFixed(0); default: return 'other'; } }\nconsole.log(show(a, b));\n"
	for _, p := range []struct{ name, source, node, stop string }{
		{"default_parameter", kindPredicateTypes + kindAliasValues + "function show(node: N, swap = (): void => { node = b; }): string { const kind = node.kind; swap(); switch (kind) { case K.A: return node.a.toFixed(0); default: return 'other'; } }\nconsole.log(show(a));\n", "TypeError", "a read of node narrowed to A where a call may have rewritten it (node is assigned inside a function)"},
		{"parameter", types + "function f(node: NA | NB, other: NA | NB): string { const reset = () => { node = other; }; " + body, "undefined\n", "a read of node narrowed to NA where a call may have rewritten it (node is assigned inside a function)"},
		{"let", types + "function f(start: NA | NB, other: NA | NB): string { let node = start; const reset = () => { node = other; }; " + body, "undefined\n", "a read of node narrowed to NA where a call may have rewritten it (node is assigned inside a function)"},
		{"parameter_open", kindPredicateTypes + kindAliasValues + "function show(node: N, other: N): string { const reset = (): void => { node = other; }; " + openBody, "TypeError", "a read of node narrowed to A where a call may have rewritten it (node is assigned inside a function)"},
		{"let_open", kindPredicateTypes + kindAliasValues + "function show(start: N, other: N): string { let node = start; const reset = (): void => { node = other; }; " + openBody, "TypeError", "a read of node narrowed to A where a call may have rewritten it (node is assigned inside a function)"},
		{"control", types + "function f(node: NA | NB): string { const kind = node.kind; switch (kind) { case K.A: return node.a; default: return 'x'; } }\nconsole.log(f(na));\n", "yes\n", ""},
	} {
		t.Run(p.name, func(t *testing.T) {
			t.Parallel()
			observed := runAgreementNode(t, writeAgreementSource(t, "main.ts", p.source))
			if p.node == "TypeError" && (observed.code == 0 || !strings.Contains(string(observed.stderr), "TypeError")) || p.node != "TypeError" && (observed.code != 0 || string(observed.stdout) != p.node) {
				t.Fatalf("kind-alias closure rebind %s Node: %+v", p.name, observed)
			}
			if field := kindAliasRecognized(t, p.source); field != "kind" {
				t.Fatalf("kind-alias closure rebind %s: recognized %q", p.name, field)
			}
			for _, extension := range []string{".ts", ".a"} {
				if p.stop == "" {
					program := lowersAndAgreesWithNodeAs(t, p.source, extension)
					if code, out, errs := brandedNative(t, program); code != 0 || out != "yes\n" {
						t.Fatalf("kind-alias closure rebind control %s native: exit=%d stdout=%q stderr=%q", extension, code, out, errs)
					}
					continue
				}
				_, err := lowerSourceAs(t, p.source, extension)
				var stop *NotYet
				if !errors.As(err, &stop) || stop.What != p.stop {
					t.Fatalf("kind-alias closure rebind %s rule 3350 %s: %v", p.name, extension, err)
				}
			}
		})
	}
}

// Platforms' premise pin. kindAliasRewritten guards the reference's binding, never the saved tag, so a
// write to node.kind between the alias and the switch would leave the alias stale. No program makes that
// write without a view whose tag is wider than the member's literal and writable: over a readonly tag,
// the readonly field made writable is refused at the call that hands node over (family 4), and a
// mutable tag seen as wider is family 3's. If either rule is ever narrowed, the alias road opens here first.
func TestKindAliasTagWriteThroughWiderViewIsRefused(t *testing.T) {
	t.Parallel()
	const show = "function retag(target: { kind: K }): void { target.kind = K.B; }\nfunction show(node: N): string {\nconst kind = node.kind;\nretag(node);\nswitch (kind) { case K.A: return node.a.toFixed(0); default: return 'other'; } }\nconsole.log(show(a));\n"
	source := kindPredicateTypes + kindAliasValues + show
	line := strconv.Itoa(strings.Count(source[:strings.Index(source, "retag(node);")], "\n") + 1)
	for _, extension := range []string{".ts", ".a"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, source, extension)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, "seen as { kind: K; }, whose readonly field kind becomes writable") || !strings.Contains(refused.Where, ":"+line+":") {
				t.Fatalf("kind-alias tag write %s: want the readonly-made-writable refusal at the retag call on line %s, got %v", extension, line, err)
			}
		})
	}
}
