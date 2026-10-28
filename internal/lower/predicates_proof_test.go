package lower

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Behavior cannot observe proof summary flags or the checked-field admission handoff.
// Bodies of the positive probes come from TypeScript 6.0.3 core.ts:1769
// and factory/nodeTests.ts:318. These test proof summaries, not admission:
// open interface summaries still require the checked-view lowering handoff.
func TestPredicateBodyProof(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source, failure string
		tagged                bool
	}{
		{"condition assertion", `function fail(message?: string): never { throw new Error(message ?? "failed"); } function assert(expression: unknown, message?: string): asserts expression { if (!expression) { fail(message); } }`, "", false},
		{"returning fail", `function fail(message?: string): void { return; } function assert(expression: unknown, message?: string): asserts expression { if (!expression) { fail(message); } }`, "normal return", false},
		{"empty condition assertion", `function assert(expression: unknown): asserts expression {}`, "normal return", false},
		{"condition mutation", `function fail(): never { throw new Error("failed"); } function assert(expression: unknown): asserts expression { if (!expression) fail(); expression = false; }`, "mutation", false},
		{"diagnostic work on failure", `function fail(message?: string): never { const error = new Error(message ?? "failed"); throw error; } function assert(expression: unknown, message?: string): asserts expression { if (!expression) { message = "failed"; fail(message); } }`, "", false},
		{"typeof", `function isString(text: unknown): text is string { return typeof text === "string"; }`, "", false},
		{"kind", `const SyntaxKind = { Identifier: 80 } as const; interface Node { readonly kind: number } interface Identifier extends Node { readonly kind: typeof SyntaxKind.Identifier; readonly escapedText: string } export function isIdentifier(node: Node): node is Identifier { return node.kind === SyntaxKind.Identifier; }`, "", true},
		{"kind helpers fixed point", `interface Node { readonly kind: number } interface Identifier extends Node { readonly kind: 80; readonly escapedText: string } interface StringLiteral extends Node { readonly kind: 11; readonly text: string } type ModuleName = Identifier | StringLiteral; function isModuleName(node: Node): node is ModuleName { return isIdentifier(node) || isStringLiteral(node); } function isIdentifier(node: Node): node is Identifier { return node.kind === 80; } function isStringLiteral(node: Node): node is StringLiteral { return node.kind === 11; }`, "", true},
		{"helper", `function isString(text: unknown): text is string { return typeof text === "string"; } function guard(x: unknown): x is string { return isString(x); }`, "", false},
		{"false branch lies", `function guard(x: unknown): x is number { return typeof x === "number" && x > 0; }`, "false return", false},
		{"true branch lies", `function guard(x: unknown): x is number { return typeof x === "string"; }`, "true return", false},
		{"mutation", `interface Node { kind: number } interface Identifier extends Node { kind: 80; readonly escapedText: string } function guard(node: Node): node is Identifier { const yes = node.kind === 80; node.kind = 11; return yes; }`, "mutation", false},
		{"alias mutation", `interface Node { kind: number } interface Identifier extends Node { kind: 80; readonly escapedText: string } function guard(node: Node): node is Identifier { const alias = node; const yes = node.kind === 80; alias.kind = 11; return yes; }`, "mutation", false},
		{"recursive", `function guard(x: unknown): x is number { return guard(x); }`, "unconverged", false},
		{"default changes input", `function guard(x: unknown = 42): x is number { return typeof x === "number"; }`, "parameter initialization", false},
		{"empty assertion", `function guard(x: unknown): asserts x is number {}`, "normal return", false},
		{"assertion", `import { panic } from 'adamic'; function guard(x: unknown): asserts x is number { if (typeof x !== "number") panic("number required"); }`, "", false},
		{"disabled assertion", `import { panic } from 'adamic'; function guard(x: unknown, enabled: boolean): asserts x is number { if (enabled && typeof x !== "number") panic("number required"); }`, "normal return", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(path, []byte(probe.source), 0644); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			l := &lowering{program: program, checker: checked}
			var predicate *ast.Node
			var visit ast.Visitor
			visit = func(n *ast.Node) bool {
				if n.Kind == ast.KindTypePredicate && (probe.name != "kind helpers fixed point" || n.Parent.Name().Text() == "isModuleName") {
					predicate = n
				}
				n.ForEachChild(visit)
				return false
			}
			file.AsNode().ForEachChild(visit)
			proof, err := l.provePredicate(predicate)
			if probe.failure != "" {
				if err == nil || !strings.Contains(err.Error(), probe.failure) || !strings.Contains(err.Error(), "return a boolean and narrow at the caller") {
					t.Fatalf("want %q refusal with caller fix, got %+v, %v", probe.failure, proof, err)
				}
			} else if err != nil || proof.TaggedView != probe.tagged {
				t.Fatalf("want proof tagged=%v, got %+v, %v", probe.tagged, proof, err)
			}
			if probe.failure == "" && probe.tagged {
				l.result = &ir.Program{}
				if admission := l.predicateRefusal(predicate); admission != nil {
					t.Fatalf("want checked-view admission, got %v", admission)
				}
				target := checked.GetTypeAtLocation(predicate.AsTypePredicateNode().Type)
				members := []*checker.Type{target}
				if target.Flags()&checker.TypeFlagsUnion != 0 {
					members = target.Types()
				}
				for _, member := range members {
					if !l.result.CheckedFields[ir.CheckedFieldKey(int(member.Id()), "kind")] {
						t.Fatal("kind admission lost receiver-specific checked fields")
					}
				}

			}
		})
	}
}

func TestConditionAssertionAdmission(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, `function fail(message?: string): never { throw new Error(message ?? "failed"); }
function assert(expression: unknown, message?: string): asserts expression {
 if (!expression) { fail(message); }
}
assert('present'); console.log('present');`)
}

func TestPredicateCallbackContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		refused      bool
	}{
		{"inferred arrow", `function apply(callback: (value: number) => value is 1): boolean { return callback(1); } console.log(String(apply((value: number) => value === 1)));`, false},
		{"inferred opaque", `function lie(value: number): value is 1; function lie(value: number): boolean { return true; } function apply(callback: (value: number) => value is 1): boolean { return callback(2); } apply((value: number) => lie(value));`, true},
		{"inferred named", `function apply(callback: (value: number) => value is 1): boolean { return callback(1); } function guard(value: number) { return value === 1; } console.log(String(apply(guard)));`, false},
		{"arrow", `function apply(callback: (value: number) => value is number): boolean { return callback(1); } console.log(String(apply((value: number): value is number => typeof value === "number")));`, false},
		{"named", `function apply(callback: (value: number) => value is number): boolean { return callback(1); } function isNumber(value: number): value is number { return typeof value === "number"; } console.log(String(apply(isNumber)));`, false},
		{"unproven", `function apply(callback: (value: number) => value is number): boolean { return callback(1); } apply((value: number): value is number => true);`, true},
		{"lying", `function apply(callback: (value: number | string) => value is number): boolean { return callback(1); } apply((value: number | string): value is number => typeof value === "number" && value > 0);`, true},
		{"reassigned", `function apply(callback: (value: number) => value is number): boolean { return callback(1); } let guard = (value: number): value is number => typeof value === "number"; guard = (value: number): value is number => true; apply(guard);`, true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if !probe.refused {
				lowersAndAgreesWithNode(t, probe.source)
				return
			}
			_, err := lowerSource(t, probe.source)
			if probe.refused {
				if err == nil || !strings.Contains(err.Error(), "unproven predicate argument for parameter callback") {
					t.Fatalf("want pinned callback argument refusal, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEveryNeedsCallbackEffects(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/predicates/every_alias.ts")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "every.a")
	if err = os.WriteFile(path, source, 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{program: program, checker: checked, result: &ir.Program{}}
	err = l.refuse(file)
	if err == nil || !strings.Contains(err.Error(), "type predicate") {
		t.Fatalf("want array proof refusal without callback effects, got %v", err)
	}
}

// Behavior cannot observe the proof analysis's per-call use-direction facts.
func TestPredicateUseRegions(t *testing.T) {
	t.Parallel()
	prefix := `function some<T>(xs: readonly T[] | undefined): xs is readonly T[];
function some<T>(xs: readonly T[] | undefined): boolean { return xs !== undefined && xs.length > 0; }
`
	for _, probe := range []struct {
		name, body string
		want       predicateTruth
	}{
		{"unused", `function use(xs: readonly number[] | undefined): void { if (some(xs)) console.log("yes"); else console.log("no"); }`, 0},
		{"true", `function use(xs: readonly number[] | undefined): void { if (some(xs)) console.log(xs.length.toString()); }`, predicateTrue},
		{"false", `function use(xs: readonly number[] | undefined): void { if (some(xs)) {} else console.log(xs === undefined ? "no" : "yes"); }`, predicateFalse},
		{"both", `function use(xs: readonly number[] | undefined): void { if (some(xs)) console.log(xs.length.toString()); else console.log(xs === undefined ? "no" : "yes"); }`, predicateEither},
		{"return continuation", `function use(xs: readonly number[] | undefined): void { if (!some(xs)) return; console.log(xs.length.toString()); }`, predicateTrue},
		{"throw continuation", `function use(xs: readonly number[] | undefined): void { if (!some(xs)) throw new Error("no"); console.log(xs.length.toString()); }`, predicateTrue},
		{"false continuation", `function use(xs: readonly number[] | undefined): void { if (some(xs)) return; console.log(xs === undefined ? "no" : "yes"); }`, predicateFalse},
		{"const result alias", `function use(xs: readonly number[] | undefined): void { const present = some(xs); if (present) console.log(xs.length.toString()); }`, predicateTrue},
		{"compound result alias", `function use(xs: readonly number[] | undefined, flag: boolean): void { const present = some(xs) && flag; if (present) console.log(xs.length.toString()); }`, predicateTrue},
		{"short circuit", `function use(xs: readonly number[] | undefined): void { some(xs) && console.log(xs.length.toString()); }`, predicateTrue},
		{"closure read", `function use(xs: readonly number[] | undefined): void { if (!some(xs)) return; const later = () => xs.length.toString(); console.log(later()); }`, predicateTrue},
		{"indexed spelling", `function use(box: { readonly xs: readonly number[] | undefined }): void { if (!some(box.xs)) return; console.log(box["xs"].length.toString()); }`, predicateTrue},
		{"indexed argument", `function use(box: { readonly xs: readonly number[] | undefined }): void { if (!some(box["xs"])) return; console.log(box.xs.length.toString()); }`, predicateTrue},
		{"computed index", `function use(arrays: readonly (readonly number[] | undefined)[], index: number): void { if (!some(arrays[index])) return; console.log(arrays[index].length.toString()); }`, predicateTrue},
		{"loop read", `function use(xs: readonly number[] | undefined): void { while (some(xs)) { console.log(xs.length.toString()); break; } }`, predicateTrue},
		{"later read", `function use(xs: readonly number[] | undefined): void { if (!some(xs)) return; console.log("later"); console.log(xs.length.toString()); }`, predicateTrue},
		{"replacement", `function use(xs: readonly number[] | undefined): void { if (!some(xs)) return; xs = [1]; console.log(xs.length.toString()); }`, 0},
		{"false comparison", `function use(xs: readonly number[] | undefined): void { if (some(xs) === false) console.log(xs === undefined ? "no" : "yes"); }`, predicateFalse},
		{"loose false comparison", `function use(xs: readonly number[] | undefined): void { if (some(xs) == false) console.log(xs === undefined ? "no" : "yes"); }`, predicateFalse},
		{"true inequality", `function use(xs: readonly number[] | undefined): void { if (some(xs) !== true) console.log(xs === undefined ? "no" : "yes"); }`, predicateFalse},
		{"reversed false comparison", `function use(xs: readonly number[] | undefined): void { if (false === some(xs)) console.log(xs === undefined ? "no" : "yes"); }`, predicateFalse},
		{"switch true default", `function use(xs: readonly number[] | undefined): void { switch (true) { case some(xs): return; default: console.log(xs === undefined ? "no" : "yes"); } }`, predicateFalse},
		{"comma continuation", `function use(xs: readonly number[] | undefined): void { if ((console.log("tick"), some(xs))) return; console.log(xs === undefined ? "no" : "yes"); }`, predicateFalse},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "uses.a")
			if err := os.WriteFile(path, []byte(prefix+probe.body), 0644); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			l := &lowering{program: program, checker: checked, result: &ir.Program{}}
			var call *ast.CallExpression
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindCallExpression && node.Expression().Kind == ast.KindIdentifier && node.Expression().Text() == "some" {
					call = node.AsCallExpression()
				}
				node.ForEachChild(visit)
				return false
			}
			file.AsNode().ForEachChild(visit)
			if call == nil {
				t.Fatal("missing probe call")
			}
			if got := l.predicateUseDirections(call); got != probe.want {
				t.Fatalf("directions %d, want %d", got, probe.want)
			}
		})
	}
}

// Behavior cannot observe which call owns each proof use-direction fact.
func TestPredicateUsesBelongToEachCall(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "separate.a")
	source := `function some<T>(xs: readonly T[] | undefined): xs is readonly T[];
function some<T>(xs: readonly T[] | undefined): boolean { return xs !== undefined && xs.length > 0; }
function use(xs: readonly number[] | undefined): void {
 if (some(xs)) console.log(xs.length.toString());
 if (some(xs)) {} else { const later = () => xs === undefined ? "no" : "yes"; console.log(later()); }
}`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{program: program, checker: checked}
	var calls []*ast.CallExpression
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression && node.Expression().Kind == ast.KindIdentifier && node.Expression().Text() == "some" {
			calls = append(calls, node.AsCallExpression())
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if len(calls) != 2 {
		t.Fatalf("calls %d", len(calls))
	}
	for index, want := range []predicateTruth{predicateTrue, predicateFalse} {
		if got := l.predicateUseDirections(calls[index]); got != want {
			t.Fatalf("call %d: directions %d, want %d", index+1, got, want)
		}
	}
}

// A lowering collects every assigned symbol of the program once, and assignedSymbol answers as a fresh lowering does:
// two files in one program, a name assigned and a name only read in each, every name asked in order and then reversed
// (#pv935yw). Mutants: the walk reading the first file alone (the second file's assignment is missed), a kept answer
// read as false, or every name answered alike, each go red here; a missed assignment also reddens
// TestPredicateArgumentProofs/reassigned.
func TestAssignedSymbolAnswersAsAFreshLowering(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	first, second := filepath.Join(directory, "main.a"), filepath.Join(directory, "other.a")
	for path, source := range map[string]string{
		first:  "let assignedHere = 1; assignedHere = 2; const readHere = 3; console.log((assignedHere + readHere).toString());\n",
		second: "let assignedThere = 1; assignedThere += 1; const readThere = 4; console.log((assignedThere + readThere).toString());\n",
	} {
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	program, err := load.Load([]string{first, second})
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Files()) < 2 {
		t.Fatalf("the program holds %d files, want both roots", len(program.Files()))
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	want := map[string]bool{"assignedHere": true, "readHere": false, "assignedThere": true, "readThere": false}
	var names []*ast.Node
	for _, file := range program.Files() {
		var visit ast.Visitor
		visit = func(n *ast.Node) bool {
			if n.Kind == ast.KindVariableDeclaration {
				if _, ok := want[n.Name().Text()]; ok {
					names = append(names, n.Name())
				}
			}
			n.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
	}
	if len(names) != len(want) {
		t.Fatalf("found %d of the %d names", len(names), len(want))
	}
	fresh := func() *lowering { return &lowering{program: program, checker: checked} }
	for _, name := range names {
		if got := fresh().assignedSymbol(fresh().symbol(name)); got != want[name.Text()] {
			t.Fatalf("%s alone: assigned=%t, want %t", name.Text(), got, want[name.Text()])
		}
	}
	kept := fresh()
	reversed := slices.Clone(names)
	slices.Reverse(reversed)
	for _, name := range append(slices.Clone(names), reversed...) {
		if got := kept.assignedSymbol(kept.symbol(name)); got != want[name.Text()] {
			t.Fatalf("%s from the memo: assigned=%t, want %t", name.Text(), got, want[name.Text()])
		}
	}
}
