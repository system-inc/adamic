package lower

import (
	"errors"
	"strings"
	"testing"
)

// sourceSlotPendingPrelude is a checked view of NodeA, a cast of node, whose flowNode may hold undefined, and
// setIt, which writes it through the view at 7:2.
const sourceSlotPendingPrelude = `interface FlowNode { readonly id: number }
interface Node { readonly kind: "a" | "b"; flowNode?: FlowNode | undefined }
interface NodeA extends Node { readonly kind: "a" }
const node: Node = { kind: "a", flowNode: { id: 1 } };
const view = node as NodeA;
function setIt(t: NodeA, v: FlowNode): void {
	t.flowNode = v;
}
`

// sourceSlotPendingRuleC is Rule C's NotYet at setIt's write, as the view walk made it, and
// sourceSlotPendingReach the optional-field write rule's, which judges a lowered write by what reaches it.
const (
	sourceSlotPendingRuleC = "main.ts:7:2: stage 0 can't lower writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds yet"
	sourceSlotPendingReach = "main.ts:7:2: stage 0 can't lower a write to optional field flowNode through a checked view, where what reaches its receiver isn't known yet"
)

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

// Rule C's NotYet for a plain write to a checked field that may hold undefined or null is held at the view walk
// and given back only where the write's statement was lowered (Adamic's ruling on #rgrtt32, with Compiler's four
// conditions): the view walk judges every write in the program from types, and a write the program never
// lowers changes no slot. When it fires it's the NotYet the view walk made, its words and position unchanged.
// A function declaration's body is lowered whether or not anything calls it, so its write is judged: an
// uncalled one, an exported one and one called through a function value stop at the optional-field write rule,
// which judges the lowered write by what reaches its receiver and runs before the held stop. A generic
// function's body is lowered only for an instance, so one nothing instantiates writes nothing, and lowers.
// Mutant: the arm gives its NotYet back at the view walk again, and "generic never instantiated" goes red,
// stopping at 7:2 in Rule C's words (Loom job J-v6rhqyq, at #rgrtt32's fcc0f8da without this change).
func TestSourceSlotPendingStopsOnlyWhereTheWriteIsLowered(t *testing.T) {
	t.Parallel()
	exported := strings.Replace(sourceSlotPendingPrelude, "function setIt", "export function setIt", 1)
	generic := strings.Replace(sourceSlotPendingPrelude, "function setIt(t: NodeA, v: FlowNode)", "function setIt<V extends FlowNode>(t: NodeA, v: V)", 1)
	judgeSourceSlotPending(t, []sourceSlotPendingRow{
		{"never called", sourceSlotPendingPrelude + "console.log(String(view.kind));\n", sourceSlotPendingReach},
		{"called on the viewed allocation", sourceSlotPendingPrelude + "setIt(view, { id: 2 });\nconsole.log(String(node.flowNode === undefined));\n", sourceSlotPendingRuleC},
		// The view is asked twice, so the indexed walk reads its kept answer for the write the second time
		// (viewSiteJudged), which is nil now the arm holds the stop: the held stop still fires.
		{"viewed twice", sourceSlotPendingPrelude + "const again = node as NodeA;\nsetIt(again, { id: 2 });\nconsole.log(String(again.kind));\n", sourceSlotPendingRuleC},
		// An exported function's parameter reaches Unknown (Compiler's condition 3), so the reach rule stops it.
		{"exported and called", exported + "setIt(view, { id: 2 });\nconsole.log(String(view.kind));\n", sourceSlotPendingReach},
		// Language's attacks: a write reached only through a monomorphized instance, and one a call through a
		// function value reaches, each lowered, each stopping.
		{"generic never instantiated", generic + "console.log(String(view.kind));\n", ""},
		{"generic instance", generic + "setIt(view, { id: 2 });\nconsole.log(String(view.kind));\n", sourceSlotPendingRuleC},
		{"function value", sourceSlotPendingPrelude + "const set: (t: NodeA, v: FlowNode) => void = setIt;\nset(view, { id: 2 });\nconsole.log(String(view.kind));\n", sourceSlotPendingReach},
	})
}

// Only a plain `=` statement reaches Rule C's arm (hostMemberAssignment), so the other forms a write through the
// view takes are never held, and stop as they did (Compiler's condition 4). The `??=` and destructuring rows are
// answered at the write by view's own stop, "without its source-slot type certificate", at the view walk. The
// setter row's receiver is a class owning flowNode's accessors, and it stops before the write is judged, at the
// getter the view's contract would read through (4:85, "a getter in a checked field contract").
func TestSourceSlotPendingOtherWriteFormsStop(t *testing.T) {
	t.Parallel()
	const certificate = "stage 0 can't lower writing a checked object field without its source-slot type certificate yet"
	coalesce := strings.Replace(sourceSlotPendingPrelude, "t.flowNode = v;", "t.flowNode ??= v;", 1)
	destructure := strings.Replace(sourceSlotPendingPrelude, "function setIt(t: NodeA, v: FlowNode): void {\n\tt.flowNode = v;", "function setIt(t: NodeA, v: { flowNode: FlowNode }): void {\n\t({ flowNode: t.flowNode } = v);", 1)
	setter := strings.Replace(sourceSlotPendingPrelude, `const node: Node = { kind: "a", flowNode: { id: 1 } };`, `class Held { readonly kind = "a" as const; #flow: FlowNode | undefined = undefined; get flowNode(): FlowNode | undefined { return this.#flow; } set flowNode(v: FlowNode | undefined) { this.#flow = v; } }`+"\nconst node: Node = new Held();", 1)
	judgeSourceSlotPending(t, []sourceSlotPendingRow{
		{"??=", coalesce + "setIt(view, { id: 2 });\nconsole.log(String(view.kind));\n", "main.ts:7:2: " + certificate},
		{"destructuring", destructure + "setIt(view, { flowNode: { id: 2 } });\nconsole.log(String(view.kind));\n", "main.ts:7:15: " + certificate},
		{"through a setter", setter + "setIt(view, { id: 2 });\nconsole.log(String(view.kind));\n", "main.ts:4:85: stage 0 can't lower a getter in a checked field contract yet"},
	})
}

// The shape 163 tsc units stopped at with #rgrtt32's field contracts (utilities.ts:4315, setValueDeclaration's
// `symbol.valueDeclaration = node`): a checked predicate's registration views Ident, whose symbol field's
// contract reaches Sym, and a function nothing calls writes Sym's optional valueDeclaration. The census lowers a
// unit's own reach, which holds no setValueDeclaration, so it lowered no store there (Correctness's trace); a
// whole program lowers every function declaration's body, so the nearest whole-program shape is a generic
// setValueDeclaration nothing instantiates, which lowers and agrees. A declared one is lowered and judged by
// its reach. Mutant: the arm's NotYet given back at the view walk, and the generic row stops at 7:2 in Rule C's
// words, as it does at #rgrtt32's fcc0f8da without this change (Loom jobs in #5qx6qv4).
func TestSourceSlotPendingUncalledWriteThroughAPredicateView(t *testing.T) {
	t.Parallel()
	generic := strings.Replace(sourceSlotPendingPredicate, "function setValueDeclaration(symbol: Sym, node: Decl)", "function setValueDeclaration<D extends Decl>(symbol: Sym, node: D)", 1)
	judgeSourceSlotPending(t, []sourceSlotPendingRow{
		{"generic never instantiated", generic, ""},
		{"declared and lowered", sourceSlotPendingPredicate, "main.ts:7:2: stage 0 can't lower a write to optional field valueDeclaration through a checked view, where what reaches its receiver isn't known yet"},
	})
}

const sourceSlotPendingPredicate = `interface Sym { flags: number; valueDeclaration?: Decl | undefined }
interface Node { readonly kind: number; symbol?: Sym | undefined }
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

// A function constructor's body is the one road a `this.name = value` statement takes outside a class or a
// function's own statements (Compiler's gate on #5qx6qv4): its instance is born as a checked view of Nd, so
// its write of flowNode, a field that may hold undefined, is Rule C's, and the synthesized constructor lowers
// its body through lowerFunction and statement (function_constructor.go), which marks each statement before
// lowering it. So the held stop fires at the constructor's write, as the view walk gave it before. Mutant: the
// mark in statements.go taken out, and the row no longer stops there.
func TestSourceSlotPendingFunctionConstructorWriteStops(t *testing.T) {
	t.Parallel()
	judgeSourceSlotPending(t, []sourceSlotPendingRow{
		{"function constructor", sourceSlotPendingConstructor, "main.ts:6:5: stage 0 can't lower writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds yet"},
	})
}

const sourceSlotPendingConstructor = `interface FlowNode { readonly id: number }
interface Nd { kind: number; flowNode?: FlowNode | undefined }
interface Allocator { getNodeConstructor(): new (kind: number) => Nd }
function NodeObject(this: Nd, kind: number): void {
    this.kind = kind;
    this.flowNode = undefined;
}
const objectAllocator: Allocator = {
    getNodeConstructor: () => NodeObject as any,
};
let NodeConstructor: (new (kind: number) => Nd) | undefined;
function createBaseNode(kind: number): Nd {
    return new (NodeConstructor || (NodeConstructor = objectAllocator.getNodeConstructor()))(kind);
}
const first = createBaseNode(7);
console.log(String(first.kind));
`
