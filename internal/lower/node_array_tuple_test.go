package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// nodeArrayTupleHeader is t01-alias.ts's construction (internal/oracle/testdata/node_array_tuple): a four-field
// MutableNodeArray literal its caller keeps as raw, handed out as a NodeArray, so swap writes beneath the view.
const nodeArrayTupleHeader = `interface Node { kind: number }
interface Ident extends Node { kind: 1; text: string }
interface Other extends Node { kind: 2 }
interface NodeArray<T extends Node> extends ReadonlyArray<T> { pos: number; end: number; hasTrailingComma: boolean; transformFlags: number }
interface MutableNodeArray<T extends Node> extends Array<T> { pos: number; end: number; hasTrailingComma: boolean; transformFlags: number }
function makeArray(elements: MutableNodeArray<Node>): NodeArray<Node> {
	elements.pos = 0;
	elements.end = 0;
	elements.hasTrailingComma = false;
	elements.transformFlags = 0;
	return elements;
}
function isOne(a: NodeArray<Node>): a is NodeArray<Node> & readonly [Ident] {
	return a.length === 1 && a[0]?.kind === 1;
}
function isAny(a: NodeArray<Node>): a is NodeArray<Node> & readonly [Node] {
	return a.length === 1;
}
const first: Node = { kind: 1, text: "name" } as Ident;
const raw = [first] as MutableNodeArray<Node>;
const list = makeArray(raw);
function swap(): void {
	raw[0] = { kind: 2 } as Other;
}
`

// What each position read of NodeArray<E> & readonly [T1..Tn] lowers to, under ruling 3 as Adamic amended it by
// Language's must (#bbssdcy): a checked view read of Ti, ir.ViewElement marked Position, where Ti is narrower
// than E and only a length test, or a proof with a call or a write after it, comes before the read, and where Ti
// is E with nothing proving the position; the plain element read, ir.ArrayIndex, where the predicate or a kind
// test of the position dominates with nothing between, and where Ti is E under a length test. It's the one pin
// the check has that builds no binary, so it holds the rule whatever the oracle's programs reach.
//
// Mutants (each kept unapplied beside the head's report): a dominating length test skipping the check where Ti is
// narrower, which turns "length test alone" red; the proof kept across a call, which turns "call between" and
// "kind test, call between" red.
func TestNodeArrayTuplePositionReads(t *testing.T) {
	t.Parallel()
	for _, form := range []struct {
		name, source string
		checked      []string // each checked read's ViewType, in the order lowering reaches them
	}{
		{"swap between", "if (isOne(list)) {\n\tswap();\n\tif (list.length === 1) {\n\t\tconsole.log(\"text \" + typeof list[0].text);\n\t}\n}\n", []string{"Ident"}},
		{"write between", "if (isOne(list)) {\n\traw[0] = { kind: 2 } as Other;\n\tif (list.length === 1) {\n\t\tconsole.log(\"text \" + typeof list[0].text);\n\t}\n}\n", []string{"Ident"}},
		{"call between", "if (isOne(list)) {\n\tconsole.log(\"before read\");\n\tif (list.length === 1) {\n\t\tconsole.log(\"text \" + typeof list[0].text);\n\t}\n}\n", []string{"Ident"}},
		{"predicate, nothing between", "if (isOne(list)) {\n\tif (list.length === 1) {\n\t\tconsole.log(\"text \" + typeof list[0].text);\n\t}\n}\n", nil},
		{"kind test, nothing between", "if (isOne(list)) {\n\tif (list[0].kind === 1) {\n\t\tconsole.log(\"text \" + typeof list[0].text);\n\t}\n}\n", nil},
		{"kind test, call between", "if (isOne(list)) {\n\tif (list[0].kind === 1) {\n\t\tswap();\n\t\tconsole.log(\"text \" + typeof list[0].text);\n\t}\n}\n", []string{"Ident"}},
		{"length test alone", "function textOf(narrowed: NodeArray<Node> & readonly [Ident]): string {\n\tif (narrowed.length === 1) {\n\t\treturn narrowed[0].text;\n\t}\n\treturn \"\";\n}\nif (isOne(list)) {\n\tconsole.log(textOf(list));\n}\n", []string{"Ident"}},
		{"element type, length test", "if (isAny(list)) {\n\tswap();\n\tif (list.length === 1) {\n\t\tconsole.log(\"kind \" + list[0].kind);\n\t}\n}\n", nil},
		{"element type, nothing proving", "function kindOf(narrowed: NodeArray<Node> & readonly [Node]): number {\n\treturn narrowed[0].kind;\n}\nif (isAny(list)) {\n\tconsole.log(String(kindOf(list)));\n}\n", []string{"Node"}},
	} {
		t.Run(form.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, nodeArrayTupleHeader+form.source, ".ts")
			if err != nil {
				t.Fatalf("lowering: %v", err)
			}
			checked := []string{}
			visit := func(node any) bool {
				if read, isRead := node.(ir.ViewElement); isRead && read.Position {
					if read.Contract == 0 {
						t.Errorf("a position read checked against no contract: %+v", read)
					}
					checked = append(checked, read.ViewType)
				}
				return true
			}
			walk(program.Main, visit)
			for _, function := range program.Functions {
				walk(function.Body, visit)
			}
			if len(checked) != len(form.checked) {
				t.Fatalf("checked position reads %q, want %q", checked, form.checked)
			}
			for index := range checked {
				if checked[index] != form.checked[index] {
					t.Errorf("checked position read %d is a view of %s, want %s", index, checked[index], form.checked[index])
				}
			}
		})
	}
}

// What the rule can't decide keeps a stop: a computed index, a write of the NodeArray part's field through the
// intersection, and a road that asks for the array's element type (for...of), each stopping in lowering.
func TestNodeArrayTupleUndecidedStops(t *testing.T) {
	t.Parallel()
	for _, form := range []struct{ name, source string }{
		{"computed index", "const index = 0;\nif (isOne(list)) {\n\tconsole.log(list[index].text);\n}\n"},
		{"field write", "if (isOne(list)) {\n\tlist.pos = 5;\n}\n"},
		{"for...of", "if (isOne(list)) {\n\tfor (const each of list) {\n\t\tconsole.log(String(each.kind));\n\t}\n}\n"},
	} {
		t.Run(form.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSourceAs(t, nodeArrayTupleHeader+form.source, ".ts"); err == nil {
				t.Fatal("lowered, want a stop")
			}
		})
	}
}
