package lower

import (
	"errors"
	"strings"
	"testing"
)

// tsc's getParseTreeNode (utilitiesPublic.ts:817 at the stage 3 pin): its two overloads and implementation
// copied whole, JSDoc included, with only `export` dropped, over a minimal Node with an original chain and
// tsc's isParseTreeNode. Overload 2 promises T | undefined from a body that returns its parameter before
// reassigning it, then walks original and returns a node only its optional nodeTest vouches for (#0kkpw2e
// set 3, Adamic's ruling).
const overloadParseTreeNodeTypes = `enum SyntaxKind {
    Unknown = 0,
    Identifier = 80,
    Block = 242,
}
enum NodeFlags {
    None = 0,
    Synthesized = 16,
}
interface Node {
    readonly kind: SyntaxKind;
    readonly flags: NodeFlags;
    readonly original: Node | undefined;
}
interface Identifier extends Node {
    readonly kind: SyntaxKind.Identifier;
    readonly text: string;
}
interface Block extends Node {
    readonly kind: SyntaxKind.Block;
    readonly size: number;
}
interface Named extends Node {
    readonly text: string;
}
function isParseTreeNode(node: Node): boolean {
    return (node.flags & NodeFlags.Synthesized) === 0;
}
`

const overloadParseTreeNodeOverloads = `/**
 * Gets the original parse tree node for a node.
 *
 * @param node The original node.
 * @returns The original parse tree node if found; otherwise, undefined.
 */
function getParseTreeNode(node: Node | undefined): Node | undefined;

/**
 * Gets the original parse tree node for a node.
 *
 * @param node The original node.
 * @param nodeTest A callback used to ensure the correct type of parse tree node is returned.
 * @returns The original parse tree node if found; otherwise, undefined.
 */
function getParseTreeNode<T extends Node>(node: T | undefined, nodeTest?: (node: Node) => node is T): T | undefined;
`

const overloadParseTreeNodeImplementation = `function getParseTreeNode(node: Node | undefined, nodeTest?: (node: Node) => boolean): Node | undefined {
    if (node === undefined || isParseTreeNode(node)) {
        return node;
    }

    node = node.original;
    while (node) {
        if (isParseTreeNode(node)) {
            return !nodeTest || nodeTest(node) ? node : undefined;
        }
        node = node.original;
    }
}
`

// parsed and word are parse tree nodes; made and block are synthesized, each made from parsed.
const overloadParseTreeNodeNodes = `function isIdentifier(node: Node): node is Identifier {
    return node.kind === SyntaxKind.Identifier;
}
function isBlock(node: Node): node is Block {
    return node.kind === SyntaxKind.Block;
}
const parsed: Identifier = { kind: SyntaxKind.Identifier, flags: NodeFlags.None, original: undefined, text: "parsed" };
const word: Identifier = { kind: SyntaxKind.Identifier, flags: NodeFlags.None, original: undefined, text: "word" };
const made: Identifier = { kind: SyntaxKind.Identifier, flags: NodeFlags.Synthesized, original: parsed, text: "made" };
const block: Block = { kind: SyntaxKind.Block, flags: NodeFlags.Synthesized, original: parsed, size: 1 };
`

// overloadParseTreeNodeProgram is the shape with implementation as getParseTreeNode's body, then calls.
func overloadParseTreeNodeProgram(implementation, calls string) string {
	return overloadParseTreeNodeTypes + overloadParseTreeNodeOverloads + implementation + overloadParseTreeNodeNodes + calls
}

// The refusal a bare claim with no checkable instance gets at its call (Adamic's wording).
const (
	overloadParseTreeNodeClaimWhat = "a call to an overload whose result is honest only for a proven node test, made without one"
	overloadParseTreeNodeClaimFix  = "pass the node test that proves the result's type"
)

// Calls passing a named predicate resolve to overload 2 and are erased (rule 1): census accepts it by the
// node-test proof where main refused it for its result, and predicateArguments holds each predicate to its
// claim. A parse tree node comes back by proof (a), a synthesized one by (b), and a test that says no gives
// undefined. A one-argument call resolves to overload 1, served by relation. All four lower and agree with
// Node natively and on the JavaScript backend, and none is checked (a guess: no call reaches a narrow road).
func TestOverloadParseTreeNodeTestedCallsLower(t *testing.T) {
	t.Parallel()
	source := overloadParseTreeNodeProgram(overloadParseTreeNodeImplementation, `const found = getParseTreeNode(made, isIdentifier);
console.log(found === undefined ? "none" : found.text);
const itself = getParseTreeNode(word, isIdentifier);
console.log(itself === undefined ? "none" : itself.text);
const missed = getParseTreeNode(block, isBlock);
console.log(missed === undefined ? "none" : ` + "`${missed.size}`" + `);
const plain = getParseTreeNode(block);
console.log(plain === undefined ? "none" : ` + "`${plain.kind}`" + `);
`)
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "parsed\nword\nnone\n80\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want what Node prints", code, stdout, stderr)
	}
	if counts := program.OverloadResultChecks; counts.Checked != 0 || counts.Erased != 0 || counts.Proven != 0 {
		t.Fatalf("want no call checked, got %+v", counts)
	}
}

// A bare claim whose T is the constraint, getParseTreeNode<Node>(made), narrows nothing: erased (rule 2),
// unchecked. The counts are a guess, as above.
func TestOverloadParseTreeNodeAtItsConstraintLowers(t *testing.T) {
	t.Parallel()
	source := overloadParseTreeNodeProgram(overloadParseTreeNodeImplementation, `const wide = getParseTreeNode<Node>(made);
console.log(wide === undefined ? "none" : ` + "`${wide.kind}`" + `);
`)
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "80\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want what Node prints", code, stdout, stderr)
	}
	if counts := program.OverloadResultChecks; counts.Checked != 0 {
		t.Fatalf("want no call checked, got %+v", counts)
	}
}

// The bare claim getParseTreeNode<Identifier>(made) is an unchecked claim (rule 3), and Identifier |
// undefined is checkable at the call by its kind literal, so in .ts its result is checked on
// overloadResultCheck's road: the honest one lowers, agrees with Node, and counts one check (the count is
// a guess).
func TestOverloadParseTreeNodeBareClaimIsChecked(t *testing.T) {
	t.Parallel()
	source := overloadParseTreeNodeProgram(overloadParseTreeNodeImplementation, `const claimed = getParseTreeNode<Identifier>(made);
console.log(claimed === undefined ? "none" : claimed.text);
`)
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "parsed\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want what Node prints", code, stdout, stderr)
	}
	if counts := program.OverloadResultChecks; counts.Checked != 1 {
		t.Fatalf("want the bare claim's result checked, got %+v", counts)
	}
}

// The lie the ruling names: a synthesized Identifier whose parse tree original is a Block, claimed
// Identifier with no test. Node prints the Block's missing text; both backends stop at the call. The
// stderr words are a guess, read from overloadResultCheck's message.
func TestOverloadParseTreeNodeBareClaimLieStopsAtTheCall(t *testing.T) {
	t.Parallel()
	source := overloadParseTreeNodeProgram(overloadParseTreeNodeImplementation, `const statement: Block = { kind: SyntaxKind.Block, flags: NodeFlags.None, original: undefined, size: 7 };
const wrapped: Identifier = { kind: SyntaxKind.Identifier, flags: NodeFlags.Synthesized, original: statement, text: "wrapped" };
const claimed = getParseTreeNode<Identifier>(wrapped);
console.log(claimed === undefined ? "none" : claimed.text);
`)
	if node := runAgreementNode(t, writeAgreementSource(t, "main.ts", source)); node.code != 0 || string(node.stdout) != "undefined\n" {
		t.Fatalf("Node: exit %d, stdout %q; want the lie printed", node.code, node.stdout)
	}
	overloadCheckStops(t, source, "", "overload 2 of getParseTreeNode result at ", ": expected Identifier")
}

// A bare claim with no check to take is Refused at the call, naming the fix: Named | undefined has no kind
// literal to check in .ts, and a .a file holds no hidden check, so the checkable Identifier claim is
// Refused there too. Both texts are Adamic's wording; that the .a row reaches the call before any other
// stop is a guess. Mutant m1 (the Refused dropped): the .ts row stops elsewhere or lowers.
func TestOverloadParseTreeNodeBareClaimIsRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, extension, calls string }{
		{"uncheckable instance", ".ts", `const named = getParseTreeNode<Named>(made);
console.log(named === undefined ? "none" : named.text);
`},
		{"checkable instance in .a", ".a", `const claimed = getParseTreeNode<Identifier>(made);
console.log(claimed === undefined ? "none" : claimed.text);
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, overloadParseTreeNodeProgram(overloadParseTreeNodeImplementation, probe.calls), probe.extension)
			var refused *Refused
			if !errors.As(err, &refused) || refused.What != overloadParseTreeNodeClaimWhat || refused.Fix != overloadParseTreeNodeClaimFix {
				t.Fatalf("want the bare claim Refused at its call, got %v", err)
			}
		})
	}
}

// A test passed as undefined is main's no-type-predicate stop at the call, measured on the scout probe
// (J-tj6knrx, x3_parse_tree_node), before the node-test proof is asked.
func TestOverloadParseTreeNodeUndefinedTestStops(t *testing.T) {
	t.Parallel()
	source := overloadParseTreeNodeProgram(overloadParseTreeNodeImplementation, `const untested = getParseTreeNode(block, undefined);
console.log(untested === undefined ? "none" : ` + "`${untested.size}`" + `);
`)
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != `an unproven predicate argument for parameter nodeTest (argument "undefined")` {
		t.Fatalf("want the undefined test refused at the call, got %v", err)
	}
}

// Bodies the proof can't read stay refused at the declaration for overload 2's result, as on main
// (census_small.go's last refusal): one returns node.original untested, and one returns its parameter after
// reassigning it, so what it hands back is no longer the argument (mutant m2, the write test dropped, and
// this row is accepted). Each is a real lie: block, tested by isBlock, comes back as parsed.
func TestOverloadParseTreeNodeUnprovenBodiesStayRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, implementation string }{
		{"returns original", `function getParseTreeNode(node: Node | undefined, nodeTest?: (node: Node) => boolean): Node | undefined {
    if (node === undefined) {
        return undefined;
    }
    return node.original;
}
`},
		{"returns the parameter reassigned", `function getParseTreeNode(node: Node | undefined, nodeTest?: (node: Node) => boolean): Node | undefined {
    if (node !== undefined && !isParseTreeNode(node)) {
        node = node.original;
    }
    return node;
}
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := overloadParseTreeNodeProgram(probe.implementation, `const found = getParseTreeNode(block, isBlock);
console.log(found === undefined ? "none" : ` + "`${found.size}`" + `);
`)
			_, err := lowerSourceAs(t, source, ".ts")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, "overload 2 of getParseTreeNode result ") || !strings.Contains(refused.What, "cannot be served by implementation result") {
				t.Fatalf("want overload 2 of getParseTreeNode refused at its declaration for its result, got %v", err)
			}
		})
	}
}

// A non-direct entry is certificate A's gate's: NotYet at the call in its words, naming the form, before
// the callee is lowered (rule 4). tsc refuses (0, f)(x) itself, so the comma's left side counts.
func TestOverloadParseTreeNodeCommaCalleeIsNotYet(t *testing.T) {
	t.Parallel()
	source := overloadParseTreeNodeProgram(overloadParseTreeNodeImplementation, `let touched = 0;
const viaComma = (touched++, getParseTreeNode)(made, isIdentifier);
console.log(viaComma === undefined ? "none" : viaComma.text);
`)
	_, err := lowerSourceAs(t, source, ".ts")
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != overloadCertificateGateWords+"a comma expression callee" {
		t.Fatalf("want NotYet %q at the call, got %v", overloadCertificateGateWords+"a comma expression callee", err)
	}
}
