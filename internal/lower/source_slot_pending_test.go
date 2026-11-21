package lower

import (
	"errors"
	"strings"
	"testing"
)

// sourceSlotPendingPrelude is a checked view of NodeA, a cast of node, whose flowNode may hold undefined, and
// setIt, which writes it through the view at 7:2, and is called only where a row calls it.
const sourceSlotPendingPrelude = `interface FlowNode { readonly id: number }
interface Node { readonly kind: "a" | "b"; flowNode?: FlowNode | undefined }
interface NodeA extends Node { readonly kind: "a" }
const node: Node = { kind: "a", flowNode: { id: 1 } };
const view = node as NodeA;
function setIt(t: NodeA, v: FlowNode): void {
	t.flowNode = v;
}
`

// sourceSlotPendingRuleC is Rule C's NotYet at setIt's write, as the view walk made it.
const sourceSlotPendingRuleC = "main.ts:7:2: stage 0 can't lower writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds"

// sourceSlotPendingRow is a program and its stop matched from main.ts: on, or empty for one that lowers and
// agrees with Node.
type sourceSlotPendingRow struct{ name, source, want string }

// judgeSourceSlotPending runs each row as its own parallel subtest.
func judgeSourceSlotPending(t *testing.T, rows []sourceSlotPendingRow) {
	t.Helper()
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			if row.want == "" {
				lowersAndAgreesWithNodeAs(t, row.source, ".ts")
				return
			}
			_, err := lowerSourceAs(t, row.source, ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want NotYet %q, got %v", row.want, err)
			}
			at := strings.Index(err.Error(), "main.ts:")
			if at < 0 || err.Error()[at:] != row.want {
				t.Fatalf("want NotYet %q, got %q", row.want, err.Error())
			}
		})
	}
}

// Rule C's NotYet for a plain write to a checked field that may hold undefined or null stops only where the
// write's statement is lowered (Adamic's ruling on #rgrtt32, with Compiler's four conditions): the view walk
// judges every write in the program from types, and a write the program never lowers changes no slot. The
// stop that fires is the one the view walk made, its words and position unchanged. Mutant: the arm gives its
// NotYet back at the view walk again, and "never called" and "generic never called" go red, stopping at 7:2.
func TestSourceSlotPendingStopsOnlyWhereTheWriteIsLowered(t *testing.T) {
	t.Parallel()
	exported := strings.Replace(sourceSlotPendingPrelude, "function setIt", "export function setIt", 1)
	generic := strings.Replace(sourceSlotPendingPrelude, "function setIt(t: NodeA, v: FlowNode)", "function setIt<V extends FlowNode>(t: NodeA, v: V)", 1)
	judgeSourceSlotPending(t, []sourceSlotPendingRow{
		{"never called", sourceSlotPendingPrelude + "console.log(String(view.kind));\n", ""},
		{"called on the viewed allocation", sourceSlotPendingPrelude + "setIt(view, { id: 2 });\nconsole.log(String(node.flowNode === undefined));\n", sourceSlotPendingRuleC},
		// The view is asked twice, so the indexed walk reads its kept answer for the write the second time
		// (viewSiteJudged), which is nil now the arm holds the stop: the held stop still fires.
		{"viewed twice", sourceSlotPendingPrelude + "const again = node as NodeA;\nsetIt(again, { id: 2 });\nconsole.log(String(again.kind));\n", sourceSlotPendingRuleC},
		// An exported function's parameter reaches Unknown (Compiler's condition 3); called, its write stops.
		{"exported and called", exported + "setIt(view, { id: 2 });\nconsole.log(String(view.kind));\n", sourceSlotPendingRuleC},
		// Language's attacks: a write reached only through a monomorphized instance, and one a call through a
		// function value reaches, each lowered, each stopping.
		{"generic never called", generic + "console.log(String(view.kind));\n", ""},
		{"generic instance", generic + "setIt(view, { id: 2 });\nconsole.log(String(view.kind));\n", sourceSlotPendingRuleC},
		{"function value", sourceSlotPendingPrelude + "const set: (t: NodeA, v: FlowNode) => void = setIt;\nset(view, { id: 2 });\nconsole.log(String(view.kind));\n", sourceSlotPendingRuleC},
	})
}

// Only a plain `=` statement reaches Rule C's arm (hostMemberAssignment), so the other forms a write through the
// view takes are never held, and stop at the write as they did (Compiler's condition 4). The `??=` and
// destructuring rows are answered by view's own stop, "without its source-slot type certificate", at the view
// walk. The setter row writes a receiver whose class owns flowNode's accessors.
func TestSourceSlotPendingOtherWriteFormsStop(t *testing.T) {
	t.Parallel()
	const certificate = "stage 0 can't lower writing a checked object field without its source-slot type certificate"
	coalesce := strings.Replace(sourceSlotPendingPrelude, "t.flowNode = v;", "t.flowNode ??= v;", 1)
	destructure := strings.Replace(sourceSlotPendingPrelude, "function setIt(t: NodeA, v: FlowNode): void {\n\tt.flowNode = v;", "function setIt(t: NodeA, v: { flowNode: FlowNode }): void {\n\t({ flowNode: t.flowNode } = v);", 1)
	setter := strings.Replace(sourceSlotPendingPrelude, `const node: Node = { kind: "a", flowNode: { id: 1 } };`, `class Held { readonly kind = "a" as const; #flow: FlowNode | undefined = undefined; get flowNode(): FlowNode | undefined { return this.#flow; } set flowNode(v: FlowNode | undefined) { this.#flow = v; } }` + "\nconst node: Node = new Held();", 1)
	judgeSourceSlotPending(t, []sourceSlotPendingRow{
		{"??=", coalesce + "setIt(view, { id: 2 });\nconsole.log(String(view.kind));\n", "main.ts:7:2: " + certificate},
		{"destructuring", destructure + "setIt(view, { flowNode: { id: 2 } });\nconsole.log(String(view.kind));\n", "main.ts:7:15: " + certificate},
		{"through a setter", setter + "setIt(view, { id: 2 });\nconsole.log(String(view.kind));\n", "main.ts:8:2: SETTER"},
	})
}

// The shape 163 tsc units stopped at with #rgrtt32's field contracts (utilities.ts:4315, setValueDeclaration's
// `symbol.valueDeclaration = node`): a checked predicate's registration views Ident, whose symbol field's
// contract reaches Sym, and a function nothing calls writes Sym's optional valueDeclaration. It lowers and
// agrees. With #rgrtt32's fcc0f8da alone, which builds Ident's sibling contracts where an optional Map field
// had made the whole type unsupported, it stops at 7:2 in Rule C's words (Loom jobs in #5qx6qv4).
func TestSourceSlotPendingUncalledWriteThroughAPredicateView(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeAs(t, sourceSlotPendingPredicate, ".ts")
}

const sourceSlotPendingPredicate = `interface Sym { flags: number; valueDeclaration?: Decl | undefined }
interface Node { readonly kind: number; symbol?: Sym | undefined; locals?: Map<string, Sym> | undefined }
interface Decl extends Node { _declarationBrand: undefined }
interface Ident extends Node { readonly kind: 80; text: string }
function isIdent(n: Node): n is Ident { return n.kind === 80; }
function setValueDeclaration(symbol: Sym, node: Decl): void {
	symbol.valueDeclaration = node;
}
const raw = { kind: 80, text: "x" };
const n: Node = raw;
if (isIdent(n)) console.log(n.text);
`
