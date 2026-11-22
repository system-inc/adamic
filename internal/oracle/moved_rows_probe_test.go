//go:build movedprobe

package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// The rows #vks2w2y moves, held to Node before they're pinned again: a generic function's type parameter
// read from the checker's resolved signature when inference left it unbound (internal/lower/generic.go)
// turns internal/lower's stops into lowerings. Each row's source is copied from its test at 6d3751a8,
// prelude and all, and written as main with that test's extension, then held to Node as
// inferred_binding_test.go holds its programs: release, sanitized native and the JavaScript backend.
// TestBrandedSupertypeStaysNotYet's assertion at predicates_branded_test.go:268 with Readonly<T> &
// TypeElement is generic_readonly_argument's own, so five rows moved, and the last two rows are that
// test's other generic rows, measured beside them.
// A probe, never landed: it prints a line per row and pins nothing.

// movedAccessTokenSource is enum_indexed_access_source_test.go's accessTokenSource.
const movedAccessTokenSource = `enum E { A, B, C }
interface SyntaxNode {
	readonly kind: E;
}
interface Leaf extends SyntaxNode {
	readonly kind: E.B;
}
function token<K extends E>(kind: K): K {
	return kind;
}
function make<T extends SyntaxNode>(kind: T["kind"]) {
	return token(kind);
}
console.log(String(make<Leaf>(E.B)));
console.log(String(make<SyntaxNode>(E.C)));
`

// movedAccessUnionSource is enum_indexed_access_source_test.go's accessUnionSource.
const movedAccessUnionSource = `enum E { A, B, C }
interface SyntaxNode {
	readonly kind: E;
}
interface Named extends SyntaxNode {
	readonly kind: E.A;
	readonly text: string;
}
interface Token<K extends E> extends SyntaxNode {
	readonly kind: K;
}
function finish<T extends SyntaxNode>(node: T): T {
	return node;
}
function makeNamed(): Named {
	return { kind: E.A, text: 'n' };
}
function token<K extends E>(kind: K): Token<K> {
	const made: Token<K> = { kind };
	return made;
}
function missing<T extends SyntaxNode>(kind: T["kind"], named: boolean): SyntaxNode {
	const result = named ? makeNamed() : token(kind);
	return finish(result);
}
console.log(String(missing<Named>(E.A, true).kind));
console.log(String(missing<SyntaxNode>(E.C, false).kind));
`

// movedBrandedPrelude is predicates_branded_test.go's brandedPrelude.
const movedBrandedPrelude = `interface Node { readonly kind: number; readonly pos: number; }
interface Declaration extends Node { _declarationBrand: undefined; }
interface TypeElement extends Declaration { _typeElementBrand: undefined; readonly name?: string; }
interface PropertySignature extends TypeElement { readonly kind: 171; readonly name: string; readonly typeText: string; }
interface MethodSignature extends TypeElement { readonly kind: 173; readonly name: string; readonly count: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
function kindOf(n: Node): number { return n.kind; }
function isTypeElement(n: Node): n is TypeElement { const k = kindOf(n); return k === 171 || k === 173; }
function isTypeElementLie(n: Node): n is TypeElement { return kindOf(n) !== 0; }
`

// movedCallLie is TestBrandedSupertypeEveryCallPathIsJudged's lie, written after the prelude.
const movedCallLie = `const i: Identifier = { kind: 80, pos: 2, escapedText: "e" };
`

// movedRow is one moved row: its test and row name, its file's extension, and its whole source.
type movedRow struct{ name, extension, source string }

var movedRows = []movedRow{
	{"TestIndexedAccessSourceIntoWholeEnumSlotLowers/accessTokenSource", ".a", movedAccessTokenSource},
	{"TestIndexedAccessSourceInUnionMemberLowers/accessUnionSource", ".a", movedAccessUnionSource},
	{"TestBrandedSupertypeEveryCallPathIsJudged/generic_claim", ".ts", movedBrandedPrelude + movedCallLie + `function isT<T extends Node>(n: Node): n is T { return kindOf(n) !== 0; }
function show(n: Node): void { if (isTypeElementLie(n)) { console.log("element"); } }
function g<U extends Node>(x: PropertySignature | Identifier): void { if (isT<U>(x)) { console.log(String(x.pos)); } }
g<TypeElement>(i);
show(i);`},
	{"TestBrandedSupertypeEveryCallPathIsJudged/generic_branded_supertype", ".ts", movedBrandedPrelude + movedCallLie + `interface Tagged<V> { readonly kind: number; readonly pos: number; _taggedBrand: undefined; readonly value: V }
interface Sub extends Tagged<number> { readonly kind: 7; readonly extra: string; }
function isTagged(n: Node): n is Tagged<number> { return kindOf(n) !== 0; }
function isTaggedOf<U extends number>(n: Node): n is Tagged<U> { return kindOf(n) !== 0; }
function show(n: Node): void { if (isTagged(n)) { console.log(String(n.pos)); } }
function g<U extends number>(x: Sub | Identifier): void { if (isTaggedOf<U>(x)) { console.log(String(x.pos)); } }
show(i);
g<number>(i);`},
	{"TestBrandedSupertypeStaysNotYet/generic_readonly_argument", ".ts", movedBrandedPrelude + `function show<T extends Node>(n: Readonly<T>): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; const v: Node = w; show(v);`},
	{"TestBrandedSupertypeStaysNotYet/generic_argument_through_a_supertype", ".ts", movedBrandedPrelude + `function show<T extends Declaration>(n: T): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5, _declarationBrand: undefined }; show(w);`},
	{"TestBrandedSupertypeStaysNotYet/generic_anonymous_argument", ".ts", movedBrandedPrelude + `function show<K extends number>(n: { readonly kind: K; readonly pos: number }): void { if (isTypeElementLie(n)) { console.log(String(n.name)); } }
const w = { kind: 80, pos: 2, escapedText: "e", name: 5 }; show(w);`},
}

// Every row prints, whatever it meets: a stage whose helper calls t.Fatal ends in its own subtest,
// whose log holds the helper's words, and the row prints the stage as fatal.
func TestMovedRowsProbe(t *testing.T) {
	t.Parallel()
	for _, row := range movedRows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			fmt.Printf("ROW %s %s\n", row.name, probeMovedRow(t, row))
		})
	}
}

// probeMovedRow writes row's source, runs it on Node, lowers it and runs each backend, and says what each
// stage met, as the fields after the row's name.
func probeMovedRow(t *testing.T, row movedRow) string {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, "main"+row.extension)
	if err := os.WriteFile(path, []byte(row.source), 0o644); err != nil {
		return "write=" + err.Error()
	}
	var truth run
	if !movedStage(t, "node", func(t *testing.T) { truth = onNode(t, path) }) {
		return "node=fatal(see its node subtest)"
	}
	fields := []string{fmt.Sprintf("node=%d:%s", truth.exitCode, strings.ReplaceAll(string(truth.stdout), "\n", "|"))}
	if len(truth.stderr) != 0 {
		fields = append(fields, fmt.Sprintf("node_stderr=%.300q", truth.stderr))
	}
	// lowered stops the test at a source tsc refuses, so the checker's words are read here first.
	if _, err := load.Load([]string{path}); err != nil {
		return strings.Join(append(fields, "lowered=load:"+strings.ReplaceAll(err.Error(), directory+string(os.PathSeparator), "")), " ")
	}
	var program *ir.Program
	var stop error
	if !movedStage(t, "lowered", func(t *testing.T) { program, stop = lowered(t, path) }) {
		return strings.Join(append(fields, "lowered=fatal(see its lowered subtest)"), " ")
	}
	if stop != nil {
		return strings.Join(append(fields, "lowered="+strings.ReplaceAll(stop.Error(), directory+string(os.PathSeparator), "")), " ")
	}
	fields = append(fields, "lowered=ok")
	var sanitized, release, javaScript run
	sanitizedRan := movedStage(t, "sanitized", func(t *testing.T) { sanitized, _ = natively(t, program) })
	releaseRan := movedStage(t, "release", func(t *testing.T) { release = released(t, program) })
	javaScriptRan := movedStage(t, "javascript", func(t *testing.T) { javaScript = onJavaScriptBackend(t, program) })
	return strings.Join(append(fields,
		"release="+movedVerdict(truth, release, releaseRan),
		"sanitized="+movedVerdict(truth, sanitized, sanitizedRan),
		"javascript="+movedVerdict(truth, javaScript, javaScriptRan),
	), " ")
}

// movedStage runs one stage as a subtest of its own, so a helper's t.Fatal ends that stage alone, and says
// whether it ran to its end.
func movedStage(t *testing.T, name string, body func(t *testing.T)) bool {
	t.Helper()
	finished := false
	t.Run(name, func(t *testing.T) {
		// Not parallel: the row reads what this stage returns as soon as t.Run does.
		body(t)
		finished = true
	})
	return finished
}

// movedVerdict is agree when got agrees with Node, and otherwise how it differs, with both runs.
func movedVerdict(truth, got run, ran bool) string {
	if !ran {
		return "fatal(see its subtest)"
	}
	difference := disagreement(truth, got)
	if difference == "" {
		return "agree"
	}
	return fmt.Sprintf("%q(Node:%d:%q,got:%d:%q,stderr:%.300q)", difference, truth.exitCode, truth.stdout, got.exitCode, got.stdout, got.stderr)
}
