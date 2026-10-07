package lower

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

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
				if !l.result.CheckedFields["kind"] {
					t.Fatal("kind admission lost checked fields")
				}

			}
		})
	}
}

func TestConditionAssertionAdmission(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "assert.a")
	source := `function fail(message?: string): never { throw new Error(message ?? "failed"); }
function assert(expression: unknown, message?: string): asserts expression {
 if (!expression) { fail(message); }
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
	lowering := &lowering{program: program, checker: checked}
	if err := lowering.refuse(file); err != nil {
		t.Fatalf("condition assertion must pass the production admission seam: %v", err)
	}
}

func TestPredicateCallbackContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		refused      bool
	}{
		{"arrow", `function apply(callback: (value: number) => value is number): boolean { return callback(1); } apply((value: number): value is number => typeof value === "number");`, false},
		{"named", `function apply(callback: (value: number) => value is number): boolean { return callback(1); } function isNumber(value: number): value is number { return typeof value === "number"; } apply(isNumber);`, false},
		{"unproven", `function apply(callback: (value: number) => value is number): boolean { return callback(1); } apply((value: number): value is number => true);`, true},
		{"lying", `function apply(callback: (value: number | string) => value is number): boolean { return callback(1); } apply((value: number | string): value is number => typeof value === "number" && value > 0);`, true},
		{"reassigned", `function apply(callback: (value: number) => value is number): boolean { return callback(1); } let guard = (value: number): value is number => typeof value === "number"; guard = (value: number): value is number => true; apply(guard);`, true},
	} {
		t.Run(probe.name, func(t *testing.T) {
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
	source, err := os.ReadFile("testdata/predicates/every_alias.a")
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
	if err == nil || !strings.Contains(err.Error(), "return is not proven") {
		t.Fatalf("want array proof refusal without callback effects, got %v", err)
	}
}
