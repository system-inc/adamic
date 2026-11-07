package lower

import (
	"context"
	"fmt"
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
func TestPredicateBodyProof(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		name, source, failure string
		tagged                bool
	}{
		{
			"condition assertion",
			`function fail(message?: string): never { throw new Error(message ?? "failed"); } function assert(expression: unknown, message?: string): asserts expression { if (!expression) { fail(message); } }`,
			"",
			false,
		},
		{
			"returning fail",
			`function fail(message?: string): void { return; } function assert(expression: unknown, message?: string): asserts expression { if (!expression) { fail(message); } }`,
			"normal return",
			false,
		},
		{
			"empty condition assertion",
			`function assert(expression: unknown): asserts expression {}`,
			"normal return",
			false,
		},
		{
			"condition mutation",
			`function fail(): never { throw new Error("failed"); } function assert(expression: unknown): asserts expression { if (!expression) fail(); expression = false; }`,
			"mutation",
			false,
		},
		{
			"diagnostic work on failure",
			`function fail(message?: string): never { const error = new Error(message ?? "failed"); throw error; } function assert(expression: unknown, message?: string): asserts expression { if (!expression) { message = "failed"; fail(message); } }`,
			"",
			false,
		},
		{"typeof", `function isString(text: unknown): text is string { return typeof text === "string"; }`, "", false},
		{
			"kind",
			`const SyntaxKind = { Identifier: 80 } as const; interface Node { readonly kind: number } interface Identifier extends Node { readonly kind: typeof SyntaxKind.Identifier; readonly escapedText: string } export function isIdentifier(node: Node): node is Identifier { return node.kind === SyntaxKind.Identifier; }`,
			"",
			true,
		},
		{
			"kind helpers fixed point",
			`interface Node { readonly kind: number } interface Identifier extends Node { readonly kind: 80; readonly escapedText: string } interface StringLiteral extends Node { readonly kind: 11; readonly text: string } type ModuleName = Identifier | StringLiteral; function isModuleName(node: Node): node is ModuleName { return isIdentifier(node) || isStringLiteral(node); } function isIdentifier(node: Node): node is Identifier { return node.kind === 80; } function isStringLiteral(node: Node): node is StringLiteral { return node.kind === 11; }`,
			"",
			true,
		},
		{
			"helper",
			`function isString(text: unknown): text is string { return typeof text === "string"; } function guard(x: unknown): x is string { return isString(x); }`,
			"",
			false,
		},
		{
			"false branch lies",
			`function guard(x: unknown): x is number { return typeof x === "number" && x > 0; }`,
			"false return",
			false,
		},
		{
			"true branch lies",
			`function guard(x: unknown): x is number { return typeof x === "string"; }`,
			"true return",
			false,
		},
		{
			"mutation",
			`interface Node { kind: number } interface Identifier extends Node { kind: 80; readonly escapedText: string } function guard(node: Node): node is Identifier { const yes = node.kind === 80; node.kind = 11; return yes; }`,
			"mutation",
			false,
		},
		{
			"alias mutation",
			`interface Node { kind: number } interface Identifier extends Node { kind: 80; readonly escapedText: string } function guard(node: Node): node is Identifier { const alias = node; const yes = node.kind === 80; alias.kind = 11; return yes; }`,
			"mutation",
			false,
		},
		{"recursive", `function guard(x: unknown): x is number { return guard(x); }`, "unconverged", false},
		{
			"default changes input",
			`function guard(x: unknown = 42): x is number { return typeof x === "number"; }`,
			"parameter initialization",
			false,
		},
		{"empty assertion", `function guard(x: unknown): asserts x is number {}`, "normal return", false},
		{
			"assertion",
			`import { panic } from 'adamic'; function guard(x: unknown): asserts x is number { if (typeof x !== "number") panic("number required"); }`,
			"",
			false,
		},
		{
			"disabled assertion",
			`import { panic } from 'adamic'; function guard(x: unknown, enabled: boolean): asserts x is number { if (enabled && typeof x !== "number") panic("number required"); }`,
			"normal return",
			false,
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(test.TempDir(), "main.a")
			if writeError := os.WriteFile(path, []byte(probe.source), 0644); writeError != nil {
				test.Fatal(writeError)
			}
			program, loadError := load.Load([]string{path})
			if loadError != nil {
				test.Fatal(loadError)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			lowering := &lowering{program: program, checker: checked}
			var predicate *ast.Node
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindTypePredicate &&
					(probe.name != "kind helpers fixed point" || node.Parent.Name().Text() == "isModuleName") {
					predicate = node
				}
				node.ForEachChild(visit)
				return false
			}
			file.AsNode().ForEachChild(visit)
			proof, provePredicateError := lowering.provePredicate(predicate)
			if probe.failure != "" {
				if provePredicateError == nil ||
					!strings.Contains(provePredicateError.Error(), probe.failure) ||
					!strings.Contains(provePredicateError.Error(), "return a boolean and narrow at the caller") {
					test.Fatalf(
						"want %q refusal with caller fix, got %+v, %v",
						probe.failure,
						proof,
						provePredicateError,
					)
				}
			} else if provePredicateError != nil || proof.TaggedView != probe.tagged {
				test.Fatalf("want proof tagged=%v, got %+v, %v", probe.tagged, proof, provePredicateError)
			}
			if probe.failure == "" && probe.tagged {
				lowering.result = &ir.Program{}
				if admission := lowering.predicateRefusal(predicate); admission != nil {
					test.Fatalf("want checked-view admission, got %v", admission)
				}
				target := checked.GetTypeAtLocation(predicate.AsTypePredicateNode().Type)
				members := []*checker.Type{target}
				if target.Flags()&checker.TypeFlagsUnion != 0 {
					members = target.Types()
				}
				for _, member := range members {
					if !lowering.result.CheckedFields[ir.CheckedFieldKey(int(member.Id()), "kind")] {
						test.Fatal("kind admission lost receiver-specific checked fields")
					}
				}

			}
		})
	}
}

func TestConditionAssertionAdmission(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(test, `function fail(message?: string): never { throw new Error(message ?? "failed"); }
function assert(expression: unknown, message?: string): asserts expression {
 if (!expression) { fail(message); }
}
assert('present'); console.log('present');`)
}

func TestPredicateCallbackContracts(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		name, source string
		refused      bool
	}{
		{
			"inferred arrow",
			`function apply(callback: (value: number) => value is 1): boolean { return callback(1); } console.log(String(apply((value: number) => value === 1)));`,
			false,
		},
		{
			"inferred opaque",
			`function lie(value: number): value is 1; function lie(value: number): boolean { return true; } function apply(callback: (value: number) => value is 1): boolean { return callback(2); } apply((value: number) => lie(value));`,
			true,
		},
		{
			"inferred named",
			`function apply(callback: (value: number) => value is 1): boolean { return callback(1); } function guard(value: number) { return value === 1; } console.log(String(apply(guard)));`,
			false,
		},
		{
			"arrow",
			`function apply(callback: (value: number) => value is number): boolean { return callback(1); } console.log(String(apply((value: number): value is number => typeof value === "number")));`,
			false,
		},
		{
			"named",
			`function apply(callback: (value: number) => value is number): boolean { return callback(1); } function isNumber(value: number): value is number { return typeof value === "number"; } console.log(String(apply(isNumber)));`,
			false,
		},
		{
			"unproven",
			`function apply(callback: (value: number) => value is number): boolean { return callback(1); } apply((value: number): value is number => true);`,
			true,
		},
		{
			"lying",
			`function apply(callback: (value: number | string) => value is number): boolean { return callback(1); } apply((value: number | string): value is number => typeof value === "number" && value > 0);`,
			true,
		},
		{
			"reassigned",
			`function apply(callback: (value: number) => value is number): boolean { return callback(1); } let guard = (value: number): value is number => typeof value === "number"; guard = (value: number): value is number => true; apply(guard);`,
			true,
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			if !probe.refused {
				lowersAndAgreesWithNode(test, probe.source)
				return
			}
			_, lowerSourceError := lowerSource(test, probe.source)
			if probe.refused {
				if lowerSourceError == nil ||
					!strings.Contains(lowerSourceError.Error(), "unproven predicate argument for parameter callback") {
					test.Fatalf("want pinned callback argument refusal, got %v", lowerSourceError)
				}
			} else if lowerSourceError != nil {
				test.Fatal(lowerSourceError)
			}
		})
	}
}

func TestEveryNeedsCallbackEffects(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("testdata/predicates/every_alias.ts")
	if readError != nil {
		test.Fatal(readError)
	}
	path := filepath.Join(test.TempDir(), "every.a")
	if writeError := os.WriteFile(path, source, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	lowering := &lowering{program: program, checker: checked, result: &ir.Program{}}
	refuseError := lowering.refuse(file)
	if refuseError == nil || !strings.Contains(refuseError.Error(), "type predicate") {
		test.Fatalf("want array proof refusal without callback effects, got %v", refuseError)
	}
}

// Behavior cannot observe the proof analysis's per-call use-direction facts.
func TestPredicateUseRegions(test *testing.T) {
	test.Parallel()
	prefix := `function some<T>(xs: readonly T[] | undefined): xs is readonly T[];
function some<T>(xs: readonly T[] | undefined): boolean { return xs !== undefined && xs.length > 0; }
`
	for _, probe := range []struct {
		name, body string
		want       predicateTruth
	}{
		{
			"unused",
			`function use(xs: readonly number[] | undefined): void { if (some(xs)) console.log("yes"); else console.log("no"); }`,
			0,
		},
		{
			"true",
			`function use(xs: readonly number[] | undefined): void { if (some(xs)) console.log(xs.length.toString()); }`,
			predicateTrue,
		},
		{
			"false",
			`function use(xs: readonly number[] | undefined): void { if (some(xs)) {} else console.log(xs === undefined ? "no" : "yes"); }`,
			predicateFalse,
		},
		{
			"both",
			`function use(xs: readonly number[] | undefined): void { if (some(xs)) console.log(xs.length.toString()); else console.log(xs === undefined ? "no" : "yes"); }`,
			predicateEither,
		},
		{
			"return continuation",
			`function use(xs: readonly number[] | undefined): void { if (!some(xs)) return; console.log(xs.length.toString()); }`,
			predicateTrue,
		},
		{
			"throw continuation",
			`function use(xs: readonly number[] | undefined): void { if (!some(xs)) throw new Error("no"); console.log(xs.length.toString()); }`,
			predicateTrue,
		},
		{
			"false continuation",
			`function use(xs: readonly number[] | undefined): void { if (some(xs)) return; console.log(xs === undefined ? "no" : "yes"); }`,
			predicateFalse,
		},
		{
			"const result alias",
			`function use(xs: readonly number[] | undefined): void { const present = some(xs); if (present) console.log(xs.length.toString()); }`,
			predicateTrue,
		},
		{
			"compound result alias",
			`function use(xs: readonly number[] | undefined, flag: boolean): void { const present = some(xs) && flag; if (present) console.log(xs.length.toString()); }`,
			predicateTrue,
		},
		{
			"short circuit",
			`function use(xs: readonly number[] | undefined): void { some(xs) && console.log(xs.length.toString()); }`,
			predicateTrue,
		},
		{
			"closure read",
			`function use(xs: readonly number[] | undefined): void { if (!some(xs)) return; const later = () => xs.length.toString(); console.log(later()); }`,
			predicateTrue,
		},
		{
			"indexed spelling",
			`function use(box: { readonly xs: readonly number[] | undefined }): void { if (!some(box.xs)) return; console.log(box["xs"].length.toString()); }`,
			predicateTrue,
		},
		{
			"indexed argument",
			`function use(box: { readonly xs: readonly number[] | undefined }): void { if (!some(box["xs"])) return; console.log(box.xs.length.toString()); }`,
			predicateTrue,
		},
		{
			"computed index",
			`function use(arrays: readonly (readonly number[] | undefined)[], index: number): void { if (!some(arrays[index])) return; console.log(arrays[index].length.toString()); }`,
			predicateTrue,
		},
		{
			"loop read",
			`function use(xs: readonly number[] | undefined): void { while (some(xs)) { console.log(xs.length.toString()); break; } }`,
			predicateTrue,
		},
		{
			"later read",
			`function use(xs: readonly number[] | undefined): void { if (!some(xs)) return; console.log("later"); console.log(xs.length.toString()); }`,
			predicateTrue,
		},
		{
			"replacement",
			`function use(xs: readonly number[] | undefined): void { if (!some(xs)) return; xs = [1]; console.log(xs.length.toString()); }`,
			0,
		},
		{
			"false comparison",
			`function use(xs: readonly number[] | undefined): void { if (some(xs) === false) console.log(xs === undefined ? "no" : "yes"); }`,
			predicateFalse,
		},
		{
			"loose false comparison",
			`function use(xs: readonly number[] | undefined): void { if (some(xs) == false) console.log(xs === undefined ? "no" : "yes"); }`,
			predicateFalse,
		},
		{
			"true inequality",
			`function use(xs: readonly number[] | undefined): void { if (some(xs) !== true) console.log(xs === undefined ? "no" : "yes"); }`,
			predicateFalse,
		},
		{
			"reversed false comparison",
			`function use(xs: readonly number[] | undefined): void { if (false === some(xs)) console.log(xs === undefined ? "no" : "yes"); }`,
			predicateFalse,
		},
		{
			"switch true default",
			`function use(xs: readonly number[] | undefined): void { switch (true) { case some(xs): return; default: console.log(xs === undefined ? "no" : "yes"); } }`,
			predicateFalse,
		},
		{
			"comma continuation",
			`function use(xs: readonly number[] | undefined): void { if ((console.log("tick"), some(xs))) return; console.log(xs === undefined ? "no" : "yes"); }`,
			predicateFalse,
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			path := filepath.Join(test.TempDir(), "uses.a")
			if writeError := os.WriteFile(path, []byte(prefix+probe.body), 0644); writeError != nil {
				test.Fatal(writeError)
			}
			program, loadError := load.Load([]string{path})
			if loadError != nil {
				test.Fatal(loadError)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			lowering := &lowering{program: program, checker: checked, result: &ir.Program{}}
			var call *ast.CallExpression
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindCallExpression &&
					node.Expression().Kind == ast.KindIdentifier &&
					node.Expression().Text() == "some" {
					call = node.AsCallExpression()
				}
				node.ForEachChild(visit)
				return false
			}
			file.AsNode().ForEachChild(visit)
			if call == nil {
				test.Fatal("missing probe call")
			}
			if got := lowering.predicateUseDirections(call); got != probe.want {
				test.Fatalf("directions %d, want %d", got, probe.want)
			}
		})
	}
}

// Behavior cannot observe which call owns each proof use-direction fact.
func TestPredicateUsesBelongToEachCall(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "separate.a")
	source := `function some<T>(xs: readonly T[] | undefined): xs is readonly T[];
function some<T>(xs: readonly T[] | undefined): boolean { return xs !== undefined && xs.length > 0; }
function use(xs: readonly number[] | undefined): void {
 if (some(xs)) console.log(xs.length.toString());
 if (some(xs)) {} else { const later = () => xs === undefined ? "no" : "yes"; console.log(later()); }
}`
	if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	lowering := &lowering{program: program, checker: checked}
	var calls []*ast.CallExpression
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression &&
			node.Expression().Kind == ast.KindIdentifier &&
			node.Expression().Text() == "some" {
			calls = append(calls, node.AsCallExpression())
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if len(calls) != 2 {
		test.Fatalf("calls %d", len(calls))
	}
	for index, want := range []predicateTruth{predicateTrue, predicateFalse} {
		if got := lowering.predicateUseDirections(calls[index]); got != want {
			test.Fatalf("call %d: directions %d, want %d", index+1, got, want)
		}
	}
}

// One lowering proves every predicate it's asked about from one fixed point per mode, and each answer is the one a
// fresh lowering gives for that predicate alone, whatever was asked before it: tags and typeof cells, proofs reached
// through a helper declared after them, and refusals (#pv935yw). Mutants: one map for both modes, or a summary missing
// from the memo, gives a proof where a fresh lowering refuses, or the other way round.
func TestPredicateProofMemoAnswersAsAFreshLowering(test *testing.T) {
	test.Parallel()
	source := `interface Node { readonly kind: number }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string }
interface StringLiteral extends Node { readonly kind: 11; readonly text: string }
type ModuleName = Identifier | StringLiteral;
function isModuleName(node: Node): node is ModuleName { return isIdentifier(node) || isStringLiteral(node); }
function isIdentifier(node: Node): node is Identifier { return node.kind === 80; }
function isStringLiteral(node: Node): node is StringLiteral { return node.kind === 11; }
function guard(x: unknown): x is string { return isText(x); }
function isText(x: unknown): x is string { return typeof x === "string"; }
function lies(x: unknown): x is number { return typeof x === "string"; }
function recursive(x: unknown): x is number { return recursive(x); }
`
	path := filepath.Join(test.TempDir(), "main.a")
	if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	var predicates []*ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindTypePredicate {
			predicates = append(predicates, node)
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	if len(predicates) != 7 {
		test.Fatalf("found %d predicates, want 7", len(predicates))
	}
	answer := func(proof predicateProof, sourceError error) string {
		if sourceError != nil {
			return "refused: " + sourceError.Error()
		}
		return fmt.Sprintf("proved, tagged=%v", proof.TaggedView)
	}
	fresh := map[*ast.Node]string{}
	for _, predicate := range predicates {
		fresh[predicate] = answer((&lowering{program: program, checker: checked}).provePredicate(predicate))
	}
	want := map[string]string{
		"isModuleName":    "proved, tagged=true",
		"isIdentifier":    "proved, tagged=true",
		"isStringLiteral": "proved, tagged=true",
		"guard":           "proved, tagged=false",
		"isText":          "proved, tagged=false",
		"lies":            "true return",
		"recursive":       "unconverged",
	}
	for _, predicate := range predicates {
		name := predicate.Parent.Name().Text()
		if !strings.Contains(fresh[predicate], want[name]) {
			test.Fatalf("%s alone: %s, want %s", name, fresh[predicate], want[name])
		}
	}
	shared := &lowering{program: program, checker: checked}
	reversed := slices.Clone(predicates)
	slices.Reverse(reversed)
	asked := append(slices.Clone(predicates), reversed...)
	for _, predicate := range asked {
		if got := answer(shared.provePredicate(predicate)); got != fresh[predicate] {
			test.Fatalf(
				"%s from the memo: %s, want what a fresh lowering gives: %s",
				predicate.Parent.Name().Text(),
				got,
				fresh[predicate],
			)
		}
	}
}
