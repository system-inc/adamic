package lower

import (
	"context"
	"errors"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// dispatchScan loads source as main.ts, the way a dispatch program is compiled, and runs ruling 17's scan over it on
// a lowering built by hand (numeric_keys_test.go's way), with the lowering, so a test can ask its conditions.
func dispatchScan(t *testing.T, source string) (*lowering, *dispatchCastRoads) {
	t.Helper()
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	t.Cleanup(release)
	l := &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1}
	return l, l.dispatchCasts()
}

// dispatchDeclineReasons is every declined candidate's reason, sorted, for a failure's message.
func dispatchDeclineReasons(roads *dispatchCastRoads) []string {
	reasons := []string{}
	for _, reason := range roads.declines {
		reasons = append(reasons, reason)
	}
	sort.Strings(reasons)
	return reasons
}

// dispatchTwin is source with its one-based line replaced by text, or, with insert, text put after that line, as the
// brief's twins of t00.ts are written.
func dispatchTwin(t *testing.T, source string, line int, text string, insert bool) string {
	t.Helper()
	lines := strings.Split(source, "\n")
	if line < 1 || line > len(lines) {
		t.Fatalf("line %d isn't in the source", line)
	}
	if insert {
		lines = slices.Insert(lines, line, text)
	} else {
		lines[line-1] = text
	}
	return strings.Join(lines, "\n")
}

// dispatchFreshCandidate is a fresh dispatch site for the program's one candidate lookup, none of its conditions yet
// asked, so a test can ask one condition by itself.
func dispatchFreshCandidate(t *testing.T, l *lowering) *dispatchSite {
	t.Helper()
	var site *dispatchSite
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if site = l.dispatchCandidate(node); site != nil {
			return true
		}
		return node.ForEachChild(visit)
	}
	l.program.Files()[0].AsNode().ForEachChild(visit)
	if site == nil {
		t.Fatal("no dispatch candidate")
	}
	return site
}

// Every program the rulings pin is judged by the scan: admitted, or declined with the reason of the first condition
// in dispatch_cast.go's order that fails, so a decline is pinned to its own condition and not to an earlier one that
// happens to fail too (t03.ts's lesson). Each program has one candidate lookup, the brief's site. The twins are
// t00.ts with one line changed or inserted, as the brief writes them, and t25.ts is t24.ts with its lookup's view
// and its call shortened to one argument. Mutant, red: any one condition in dispatchUseReason's list answering "",
// so its rows admit; the generic slot's row turns red with dispatchSlotAtCall returning the signature uninstantiated.
func TestDispatchCastsAdmitOnlyTheRuledShape(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name   string
		source func(t *testing.T) string
		want   string
	}{
		{"t00 the honest dispatch", func(*testing.T) string { return dispatchT00 }, ""},
		{"t05 a declaration shared by two keys", func(*testing.T) string { return dispatchT05 }, ""},
		{"t22 two members under one key", func(*testing.T) string { return dispatchT22 }, ""},
		{"t24 the slot's trailing argument", func(*testing.T) string { return dispatchT24 }, ""},
		{"a generic slot inferred at the call", func(*testing.T) string { return dispatchGenericSlot }, ""},
		{"an optional call", func(t *testing.T) string {
			return dispatchTwin(t, dispatchT00, 13, "\treturn fn === undefined ? \"none\" : fn?.(node);", false)
		}, ""},
		{"t01 a let repointed", func(*testing.T) string { return dispatchT01 }, "condition 1: x is neither a const nor a never-assigned parameter"},
		{"t02 a separately bound assertion", func(*testing.T) string { return dispatchT02 }, "the proof: the table is bound by a separate assertion"},
		{"t04 a method signature", func(t *testing.T) string {
			return dispatchTwin(t, dispatchT00, 5, "type Fn<T> = { run(node: T): string }[\"run\"];", false)
		}, "function type: key 0's call signature comes from a method signature"},
		{"c06 an entry typed any", func(*testing.T) string { return dispatchC06 }, "no any: key 0's entry has a parameter typed any"},
		{"c08 an unconstrained overload", func(*testing.T) string { return dispatchC08 }, "entries: key 0's entry binds an overloaded function declaration"},
		{"c09 a constrained overload", func(*testing.T) string { return dispatchC09 }, "entries: key 0's entry binds an overloaded function declaration"},
		{"c10 an any overload", func(*testing.T) string { return dispatchC10 }, "entries: key 0's entry binds an overloaded function declaration"},
		{"c11 an overload at a one-member key", func(*testing.T) string { return dispatchC11 }, "entries: key 0's entry binds an overloaded function declaration"},
		{"c12-trailing an extra parameter", func(*testing.T) string { return dispatchC12Trailing }, "entries count: key 0's entry declares 2 parameters, its call signature 1"},
		{"c14-trailing-one-member its one-member twin", func(*testing.T) string { return dispatchC14TrailingOneMember }, "entries count: key 0's entry declares 2 parameters, its call signature 1"},
		{"g05 a union of function types", func(*testing.T) string { return dispatchG05 }, "single signature: key 301's property is a union, not one function type"},
		{"t19 string keys", func(*testing.T) string { return dispatchT19 }, "numeric keys: a key of the table's type isn't a number"},
		{"r00 a table mapped over an enum's keys", func(*testing.T) string { return dispatchR00 }, "shape: the table's declared type is a mapped type over an enum's keys"},
		{"t25 too few arguments", func(t *testing.T) string {
			shortened := dispatchTwin(t, dispatchT24, 12, "\tconst fn = (table as Record<Kind, (node: any) => string>)[node.kind];", false)
			return dispatchTwin(t, shortened, 13, "\treturn fn === undefined ? \"none\" : fn(node);", false)
		}, "the third clause: a call passes 1 argument, key 0's call signature requires 2"},
		{"t10 passed", func(t *testing.T) string {
			return dispatchTwin(t, dispatchT00, 10, "console.log(String(Object.keys(table).length));", true)
		}, "the proof: the table is passed"},
		{"t11 stored", func(t *testing.T) string { return dispatchTwin(t, dispatchT00, 10, "const kept = table;", true) }, "the proof: the table is stored"},
		{"t12 re-exported", func(t *testing.T) string { return dispatchTwin(t, dispatchT00, 10, "export { table };", true) }, "the proof: the table is re-exported"},
		{"t13 spread", func(t *testing.T) string { return dispatchTwin(t, dispatchT00, 10, "const copy = { ...table };", true) }, "the proof: the table is spread"},
		{"t14 read without the assertion", func(t *testing.T) string {
			return dispatchTwin(t, dispatchT00, 10, "console.log(table[Kind.A]({ kind: Kind.A, a: \"y\" }));", true)
		}, "the proof: the table is read without the assertion"},
		{"t15 an exported table", func(t *testing.T) string {
			return dispatchTwin(t, dispatchT00, 7, "export const table: Table = {", false)
		}, "the proof: the table is exported"},
		{"t16 a let table", func(t *testing.T) string { return dispatchTwin(t, dispatchT00, 7, "let table: Table = {", false) }, "the proof: the table isn't a const"},
		{"t17 an accessor entry", func(t *testing.T) string {
			return dispatchTwin(t, dispatchT00, 8, "\tget [Kind.A](): Fn<NodeA> { return (node: NodeA) => \"a:\" + node.a.toUpperCase(); },", false)
		}, "entries: an accessor entry"},
		{"t18 a method entry", func(t *testing.T) string {
			return dispatchTwin(t, dispatchT00, 8, "\t[Kind.A](node: NodeA): string { return \"a:\" + node.a.toUpperCase(); },", false)
		}, "entries: a method entry"},
		{"t21 a spread entry", func(t *testing.T) string {
			annotated := strings.Replace(dispatchG04, "const table: { [K in Kind]: Fn<Extract<Node, { kind: K }>> } = {", "const table: Table = {", 1)
			return dispatchTwin(t, annotated, 5, "type Table = { [N in Node as N[\"kind\"]]: Fn<N> };", true)
		}, "entries: a spread element"},
		{"a plain call no presence test guards", func(t *testing.T) string {
			return dispatchTwin(t, dispatchT00, 13, "\treturn fn(node);", false)
		}, "escapes: a call of fn no presence test guards"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, roads := dispatchScan(t, probe.source(t))
			if len(roads.sites)+len(roads.declines) != 1 {
				t.Fatalf("want one candidate, got %d admitted and %d declined (%q)", len(roads.sites), len(roads.declines), dispatchDeclineReasons(roads))
			}
			got := ""
			for _, reason := range roads.declines {
				got = reason
			}
			if got != probe.want {
				t.Fatalf("got %q, want %q", got, probe.want)
			}
		})
	}
}

// Seats' twins of t00.ts are written from the brief's words; each must differ from t00.ts by exactly its one line, so
// a twin whose edit missed its line fails here, never as a row that passes for the wrong reason.
func TestDispatchTwinsChangeOneLine(t *testing.T) {
	t.Parallel()
	replaced := dispatchTwin(t, dispatchT00, 7, "let table: Table = {", false)
	if strings.Count(replaced, "\n") != strings.Count(dispatchT00, "\n") || !strings.Contains(replaced, "\nlet table: Table = {\n\t[Kind.A]") {
		t.Fatalf("line 7 not replaced in place:\n%s", replaced)
	}
	inserted := dispatchTwin(t, dispatchT00, 10, "const kept = table;", true)
	if strings.Count(inserted, "\n") != strings.Count(dispatchT00, "\n")+1 || !strings.Contains(inserted, "\n};\nconst kept = table;\nfunction visit(") {
		t.Fatalf("line not inserted after line 10:\n%s", inserted)
	}
}

// An admitted lookup's site holds what the lowering will read: the table, its literal and its keys in the literal's
// order, each key's members from the checker's mapped property (t22.ts's key 1 is ImportAttributes | AssertClause,
// her 20:46:47Z rule, never the entry's ImportAttributes), x, fn, and fn's one presence test and one call. An optional
// call is marked. The scan is held for the lowering. Mutant, red: dispatchEntries reading members from the entry's own
// first parameter, so t22.ts's key 1 holds ImportAttributes alone.
func TestDispatchSiteHoldsItsParts(t *testing.T) {
	t.Parallel()
	l, roads := dispatchScan(t, dispatchT00)
	if l.dispatchCasts() != roads {
		t.Fatal("a second scan wasn't the held one")
	}
	var site *dispatchSite
	for _, admitted := range roads.sites {
		site = admitted
	}
	if site == nil {
		t.Fatalf("t00.ts declined: %q", dispatchDeclineReasons(roads))
	}
	if site.table.Name != "table" || site.subject.Name != "node" || site.callee.Name != "fn" || site.literal == nil || site.literal.Kind != ast.KindObjectLiteralExpression {
		t.Fatalf("table %q, x %q, fn %q, literal %v", site.table.Name, site.subject.Name, site.callee.Name, site.literal)
	}
	if len(site.presenceTests) != 1 || site.presenceTests[0].Kind != ast.KindBinaryExpression || len(site.calls) != 1 || site.calls[0].optional {
		t.Fatalf("want one presence test and one plain call, got %d and %d", len(site.presenceTests), len(site.calls))
	}
	if len(site.keys) != 2 {
		t.Fatalf("want two keys, got %d", len(site.keys))
	}
	for index, want := range []struct {
		name    string
		value   float64
		members string
	}{{"0", 0, "NodeA"}, {"1", 1, "NodeB"}} {
		key := site.keys[index]
		if key.name != want.name || key.value != want.value || key.entry.Kind != ast.KindPropertyAssignment || key.function.Kind != ast.KindArrowFunction || key.signature == nil || l.checker.TypeToString(key.members) != want.members {
			t.Fatalf("key %d: %q %v %v %v, members %s", index, key.name, key.value, key.entry.Kind, key.function.Kind, l.checker.TypeToString(key.members))
		}
	}
	l, roads = dispatchScan(t, dispatchT22)
	for _, admitted := range roads.sites {
		site = admitted
	}
	if len(roads.sites) != 1 || len(site.keys) != 2 || site.keys[1].name != "1" || site.keys[1].members == nil {
		t.Fatalf("t22.ts: %d admitted, declines %q", len(roads.sites), dispatchDeclineReasons(roads))
	}
	names := []string{}
	for _, member := range flattenedMembers(site.keys[1].members) {
		names = append(names, member.Symbol().Name)
	}
	sort.Strings(names)
	if strings.Join(names, ",") != "AssertClause,ImportAttributes" {
		t.Fatalf("t22.ts's key 1 passes x as %s, want ImportAttributes | AssertClause", l.checker.TypeToString(site.keys[1].members))
	}
	_, roads = dispatchScan(t, dispatchTwin(t, dispatchT00, 13, "\treturn fn === undefined ? \"none\" : fn?.(node);", false))
	for _, admitted := range roads.sites {
		site = admitted
	}
	if len(roads.sites) != 1 || len(site.calls) != 1 || !site.calls[0].optional {
		t.Fatalf("the optional call: %d admitted, declines %q", len(roads.sites), dispatchDeclineReasons(roads))
	}
}

// Each of Adamic's two 22:50:42Z conditions is held by itself on c12-trailing.ts and c14-trailing-one-member.ts, which
// break both: the entries count names the entry's extra parameter, and the call condition, asked on its own once the
// entries and the escapes hold, names the extra argument. t24.ts, the honest call with the slot's trailing argument,
// passes both. Mutant, red: Adamic's first-argument-only version, the count and the assignability after x dropped
// from dispatchCallCondition, turns the call column of both trailing rows red while t24.ts's stays green.
func TestDispatchTrailingArgumentsEachCheck(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source, count, call string
	}{
		{"c12-trailing", dispatchC12Trailing, "entries count: key 0's entry declares 2 parameters, its call signature 1", "the call condition: a call passes 2 arguments, key 0's call signature declares 1"},
		{"c14-trailing-one-member", dispatchC14TrailingOneMember, "entries count: key 0's entry declares 2 parameters, its call signature 1", "the call condition: a call passes 2 arguments, key 0's call signature declares 1"},
		{"t24", dispatchT24, "", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			l, _ := dispatchScan(t, probe.source)
			site := dispatchFreshCandidate(t, l)
			if reason := l.dispatchEntries(site); reason != "" {
				t.Fatalf("entries: %q", reason)
			}
			if reason := l.dispatchEntriesCount(site); reason != probe.count {
				t.Fatalf("entries count %q, want %q", reason, probe.count)
			}
			if reason := l.dispatchEscapes(site); reason != "" {
				t.Fatalf("escapes: %q", reason)
			}
			if reason := l.dispatchCallCondition(site); reason != probe.call {
				t.Fatalf("call condition %q, want %q", reason, probe.call)
			}
		})
	}
}

// t05.ts's visitBOrC is an identifier entry bound to one function declaration with a body and no overloads, the
// signature tsc related at the initializer being the body's own, so Adamic's entries condition passes it at both its
// keys, B and C. Mutant, red: dispatchIdentifierEntry answering every identifier as overloaded.
func TestDispatchIdentifierEntryOfOneDeclaration(t *testing.T) {
	t.Parallel()
	l, _ := dispatchScan(t, dispatchT05)
	site := dispatchFreshCandidate(t, l)
	if reason := l.dispatchEntries(site); reason != "" {
		t.Fatalf("entries: %q", reason)
	}
	shared := 0
	for _, key := range site.keys {
		if key.function.Kind == ast.KindFunctionDeclaration && key.function.Name().Text() == "visitBOrC" {
			shared++
		}
	}
	if shared != 2 {
		t.Fatalf("visitBOrC bound at %d keys, want 2", shared)
	}
}

// refuseAnyRelation leaves an admitted lookup's assertion to the proof, and only that: t00.ts's refusal walk records
// no any relation, and t01.ts's, declined by condition 1, records today's stop at its assertion, in its words.
// Mutant, red: dispatchCastOperand answering true for every assertion's operand, so t01.ts's row records nothing.
func TestDispatchCastLeavesItsAssertionToTheProof(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source, want string
	}{
		{"t00 admitted", dispatchT00, ""},
		{"t01 declined", dispatchT01, "a member typed any (0(0)) related to NodeA at the assertion site"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			l, _ := dispatchScan(t, probe.source)
			_ = l.refuse(l.program.Files()[0])
			if probe.want == "" {
				if l.refusal.deferred != nil {
					t.Fatalf("want no deferred stop, got %v", l.refusal.deferred)
				}
				return
			}
			var pending *NotYet
			if !errors.As(l.refusal.deferred, &pending) || !strings.Contains(pending.What, probe.want) {
				t.Fatalf("want the deferred stop %q, got %v", probe.want, l.refusal.deferred)
			}
		})
	}
}

// dispatchGenericSlot is a dispatch whose slot is generic in its own R, as upstream's ForEachChildFunction is in its
// T: the call's trailing arguments are related to the slot's parameters once the checker infers R from the call
// (number, from 5), so a show taking a number fits. Erased to its constraint, R would be unknown, and a function
// taking a number isn't assignable to one taking unknown.
const dispatchGenericSlot = `enum Kind { A, B }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
type Node = NodeA | NodeB;
type Fn<T> = <R>(node: T, seed: R, show: (seed: R) => string) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
const table: Table = {
	[Kind.A]: <R>(node: NodeA, seed: R, show: (seed: R) => string) => "a:" + node.a.toUpperCase() + show(seed),
	[Kind.B]: <R>(node: NodeB, seed: R, show: (seed: R) => string) => "b:" + String(node.b + 1) + show(seed),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node, 5, (seed: number) => seed.toFixed(1));
}
console.log(visit({ kind: Kind.A, a: "x" }));
console.log(visit({ kind: Kind.B, b: 1 }));
`

// dispatchT00 is t00.ts, the honest dispatch: the house's model of forEachChild's lookup.
const dispatchT00 = `enum Kind { A, B }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
type Node = NodeA | NodeB;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
const table: Table = {
	[Kind.A]: (node: NodeA) => "a:" + node.a.toUpperCase(),
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.A, a: "x" }));
console.log(visit({ kind: Kind.B, b: 1 }));
`

// dispatchT01 is t01.ts, condition 1's pin: a let repointed between the lookup and the call (Language's r01.ts in the ruled shape).
const dispatchT01 = `enum Kind { A, B }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
type Node = NodeA | NodeB;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
const table: Table = {
	[Kind.A]: (node: NodeA) => "a:" + node.a.toUpperCase(),
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(start: Node, next: Node): string {
	let node = start;
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	node = next;
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.A, a: "x" }, { kind: Kind.B, b: 1 }));
`

// dispatchT02 is t02.ts, the proof's pin: the table bound by a separate assertion and written through it (Language's r02.ts in the ruled shape).
const dispatchT02 = `enum Kind { A, B }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
type Node = NodeA | NodeB;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
const table: Table = {
	[Kind.A]: (node: NodeA) => "a:" + node.a.toUpperCase(),
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
const view = table as Record<Kind, Fn<any>>;
view[Kind.A] = table[Kind.B];
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.A, a: "x" }));
`

// dispatchT05 is t05.ts, a function declaration shared by two keys and declared for their sub-union.
const dispatchT05 = `enum Kind { A, B, C }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
interface NodeC { readonly kind: Kind.C; readonly b: number }
type Node = NodeA | NodeB | NodeC;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
function visitBOrC(node: NodeB | NodeC): string {
	return (node.kind === Kind.B ? "b:" : "c:") + String(node.b + 1);
}
const table: Table = {
	[Kind.A]: (node: NodeA) => "a:" + node.a.toUpperCase(),
	[Kind.B]: visitBOrC,
	[Kind.C]: visitBOrC,
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.A, a: "x" }));
console.log(visit({ kind: Kind.B, b: 1 }));
console.log(visit({ kind: Kind.C, b: 2 }));
`

// dispatchT19 is t19.ts, the numeric-keys pin: string keys, and a kind, "constructor", no entry names (Language's g03.ts in the ruled shape).
const dispatchT19 = `enum Kind { A = "a", B = "b", C = "constructor" }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
interface NodeC { readonly kind: Kind.C }
type Node = NodeA | NodeB | NodeC;
type Fn<T> = (node: T) => string;
type Table = { [N in NodeA | NodeB as N["kind"]]: Fn<N> };
const table: Table = {
	[Kind.A]: (node: NodeA) => "a:" + node.a.toUpperCase(),
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.A, a: "x" }));
console.log(visit({ kind: Kind.C }));
`

// dispatchT22 is t22.ts, the two-member pin: ImportAttributes and AssertClause under one key, as upstream's 301.
const dispatchT22 = `enum Kind { Name, ImportAttributes, AssertClause = ImportAttributes }
interface Name { readonly kind: Kind.Name; readonly text: string }
interface ImportAttributes { readonly kind: Kind.ImportAttributes; readonly elements: string }
interface AssertClause extends ImportAttributes {}
type Node = Name | ImportAttributes | AssertClause;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
const table: Table = {
	[Kind.Name]: (node: Name) => "name:" + node.text.toUpperCase(),
	[Kind.ImportAttributes]: (node: ImportAttributes) => "attributes:" + node.elements,
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
const clause: AssertClause = { kind: Kind.AssertClause, elements: "type" };
console.log(visit({ kind: Kind.Name, text: "x" }));
console.log(visit({ kind: Kind.ImportAttributes, elements: "a" }));
console.log(visit(clause));
`

// dispatchT24 is t24.ts, the honest call with the slot's trailing argument.
const dispatchT24 = `enum Kind { A, B }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
type Node = NodeA | NodeB;
type Fn<T> = (node: T, tag: string) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
const table: Table = {
	[Kind.A]: (node: NodeA, tag: string) => tag + "a:" + node.a.toUpperCase(),
	[Kind.B]: (node: NodeB, tag: string) => tag + "b:" + String(node.b + 1),
};
function visit(node: Node, tag: string): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node, tag);
}
console.log(visit({ kind: Kind.A, a: "x" }, ">"));
console.log(visit({ kind: Kind.B, b: 1 }, ">"));
`

// dispatchC06 is c06.ts, an entry whose parameter is typed any.
const dispatchC06 = `enum Kind { A, B, C = A }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
interface NodeC { readonly kind: Kind.C; readonly c: number }
type Node = NodeA | NodeB | NodeC;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
const table: Table = {
	[Kind.A]: (node: any) => "a:" + node.a.toUpperCase(),
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.C, c: 1 }));
`

// dispatchC08 is c08.ts, an identifier entry bound to an unconstrained <T> overload over an implementation for NodeA.
const dispatchC08 = `enum Kind { A, B, C = A }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
interface NodeC { readonly kind: Kind.C; readonly c: number }
type Node = NodeA | NodeB | NodeC;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
function visitA<T>(node: T): string;
function visitA(node: NodeA): string { return "a:" + node.a.toUpperCase(); }
const table: Table = {
	[Kind.A]: visitA,
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.C, c: 1 }));
`

// dispatchC09 is c09.ts, an identifier entry bound to a <T extends NodeA | NodeC> overload over an implementation for NodeA.
const dispatchC09 = `enum Kind { A, B, C = A }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
interface NodeC { readonly kind: Kind.C; readonly c: number }
type Node = NodeA | NodeB | NodeC;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
function visitA<T extends NodeA | NodeC>(node: T): string;
function visitA(node: NodeA): string { return "a:" + node.a.toUpperCase(); }
const table: Table = {
	[Kind.A]: visitA,
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.C, c: 1 }));
`

// dispatchC10 is c10.ts, an identifier entry bound to an any overload over an implementation for NodeA.
const dispatchC10 = `enum Kind { A, B, C = A }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
interface NodeC { readonly kind: Kind.C; readonly c: number }
type Node = NodeA | NodeB | NodeC;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
function visitA(node: any): string;
function visitA(node: NodeA): string { return "a:" + node.a.toUpperCase(); }
const table: Table = {
	[Kind.A]: visitA,
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.C, c: 1 }));
`

// dispatchC11 is c11.ts, the one-member-key twin: a <T extends Node> overload over an implementation for NodeB, at key A.
const dispatchC11 = `enum Kind { A, B }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
type Node = NodeA | NodeB;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
function visitB<T extends Node>(node: T): string;
function visitB(node: NodeB): string { return "b:" + node.b.toFixed(1); }
const table: Table = {
	[Kind.A]: visitB,
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.B, b: 1 }));
console.log(visit({ kind: Kind.A, a: "x" }));
`

// dispatchC12Trailing is c12-trailing.ts, an entry with an extra optional parameter, called with two arguments through the erased view.
const dispatchC12Trailing = `enum Kind { A, B, C = A }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
interface NodeC { readonly kind: Kind.C; readonly c: number }
type Node = NodeA | NodeB | NodeC;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
const table: Table = {
	[Kind.A]: (_node: NodeA | NodeC, other?: NodeA) => other === undefined ? "none" : "a:" + other.a.toUpperCase(),
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, (node: any, other?: Node) => string>)[node.kind];
	return fn === undefined ? "none" : fn(node, node);
}
console.log(visit({ kind: Kind.C, c: 1 }));
`

// dispatchC14TrailingOneMember is c14-trailing-one-member.ts, c12-trailing.ts's one-member-key twin.
const dispatchC14TrailingOneMember = `enum Kind { A, B }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
type Node = NodeA | NodeB;
type Fn<T> = (node: T) => string;
type Table = { [N in Node as N["kind"]]: Fn<N> };
const table: Table = {
	[Kind.A]: (_node: NodeA, other?: NodeB) => other === undefined ? "none" : "b:" + other.b.toFixed(1),
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, (node: any, other?: Node) => string>)[node.kind];
	return fn === undefined ? "none" : fn(node, node);
}
console.log(visit({ kind: Kind.A, a: "x" }));
`

// dispatchG04 is g04.ts, Language's spread entry over an enum-keyed table, the source of t21.ts.
const dispatchG04 = `enum Kind { A, B }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
type Node = NodeA | NodeB;
type Fn<T> = (node: T) => string;
const shared = { [Kind.B]: (node: NodeB) => "b:" + String(node.b + 1) };
const table: { [K in Kind]: Fn<Extract<Node, { kind: K }>> } = {
	[Kind.A]: (node: NodeA) => "a:" + node.a,
	...shared,
};
shared[Kind.B] = (node: NodeB) => "late " + String(node.b);
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.B, b: 1 }));
`

// dispatchG05 is g05.ts, Language's union of function types: a distributive Fn gives key 301 one function type per member.
const dispatchG05 = `enum Kind { A = 1, Shared = 301 }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface Assert { readonly kind: Kind.Shared; readonly assertOnly: string }
interface Attributes { readonly kind: Kind.Shared; readonly count: number }
type Node = NodeA | Assert | Attributes;
type Fn<T> = T extends unknown ? (node: T) => string : never;
type Table = { [TNode in Node as TNode["kind"]]: Fn<TNode> };
const table: Table = {
	[Kind.A]: (node: NodeA) => "a:" + node.a,
	[Kind.Shared]: (node: Assert) => "assert:" + node.assertOnly.toUpperCase(),
};
type Probe = Table[Kind.Shared];
function visit(node: Node): string {
	const fn = (table as Record<Kind, (node: any) => string>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.Shared, assertOnly: "x" }));
console.log(visit({ kind: Kind.Shared, count: 2 }));
`

// dispatchR00 is r00.ts, Language's honest dispatch over a table mapped over an enum's keys.
const dispatchR00 = `enum Kind { A, B }
interface NodeA { readonly kind: Kind.A; readonly a: string }
interface NodeB { readonly kind: Kind.B; readonly b: number }
type Node = NodeA | NodeB;
type Fn<T> = (node: T) => string;
type NodeOf<K extends Kind> = Extract<Node, { kind: K }>;
const table: { [K in Kind]: Fn<NodeOf<K>> } = {
	[Kind.A]: (node: NodeA) => "a:" + node.a.toUpperCase(),
	[Kind.B]: (node: NodeB) => "b:" + String(node.b + 1),
};
function visit(node: Node): string {
	const fn = (table as Record<Kind, Fn<any>>)[node.kind];
	return fn === undefined ? "none" : fn(node);
}
console.log(visit({ kind: Kind.A, a: "x" }));
console.log(visit({ kind: Kind.B, b: 1 }));
`
