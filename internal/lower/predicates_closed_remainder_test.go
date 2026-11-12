package lower

import (
	"errors"
	"strings"
	"testing"
)

// closedRemainderPrelude is tsc's shape: an open SyntaxKind on a non-union Node, and node kinds that extend
// it with one literal kind each, so `isX(node: Node)` is a checked predicate over an open tag domain. The
// enum carries a value alias, as SyntaxKind's FirstToken and LastKeyword do.
const closedRemainderPrelude = `enum SyntaxKind { NumericLiteral = 9, StringLiteral = 11, Identifier = 80, JsxNamespacedName = 295, FirstName = Identifier }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface Identifier extends Node { readonly kind: SyntaxKind.Identifier; readonly escapedText: string }
interface StringLiteral extends Node { readonly kind: SyntaxKind.StringLiteral; readonly text: string }
interface NumericLiteral extends Node { readonly kind: SyntaxKind.NumericLiteral; readonly text: string }
interface JsxNamespacedName extends Node { readonly kind: SyntaxKind.JsxNamespacedName; readonly namespace: Identifier; readonly name: Identifier }
function isIdentifier(node: Node): node is Identifier { return node.kind === SyntaxKind.Identifier; }
function isStringLiteral(node: Node): node is StringLiteral { return node.kind === SyntaxKind.StringLiteral; }
const id: Identifier = { kind: SyntaxKind.Identifier, pos: 0, escapedText: "a" };
const ns: JsxNamespacedName = { kind: SyntaxKind.JsxNamespacedName, pos: 1, namespace: { kind: SyntaxKind.Identifier, pos: 2, escapedText: "ns" }, name: { kind: SyntaxKind.Identifier, pos: 3, escapedText: "b" } };
const str: StringLiteral = { kind: SyntaxKind.StringLiteral, pos: 4, text: "s" };
const num: NumericLiteral = { kind: SyntaxKind.NumericLiteral, pos: 5, text: "7" };
`

// tsc's getTextOfIdentifierOrLiteral (utilities.ts:5530): isIdentifier over an argument held as the closed
// union Identifier | StringLiteral | NumericLiteral, whose false side reads the other members. The false side
// is checked against the argument's own remainder, so the call lowers where it was "checked predicate
// membership over an open tag domain" (#w3e1vr8). Mutant: checkedPredicateClosedRemainder answering nil, and
// this stops there again.
func TestClosedRemainderServesTscsShape(t *testing.T) {
	t.Parallel()
	source := closedRemainderPrelude + `function textOf(node: Identifier | StringLiteral | NumericLiteral): string {
    return isIdentifier(node) ? node.escapedText : node.text;
}
console.log(textOf(id));
console.log(textOf(str));
console.log(textOf(num));
`
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "a\ns\n7\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want a, s and 7, as Node prints", code, stdout, stderr)
	}
}

// A remainder member with node-typed fields (JsxNamespacedName's namespace and name) gives its view a contract
// inside those fields, and #80cs7rz's rule then holds every Node.kind read whose reach is unknown, the
// predicates' own bodies included, until #zpzm9ta makes such reads checked. So tsc's
// getEscapedTextOfJsxAttributeName (utilities.ts:11129) moves from the open tag domain to that wall, fail
// closed, the price the true side pays for such a target too.
func TestClosedRemainderWithNodeFieldsWaitsOnCheckedReads(t *testing.T) {
	t.Parallel()
	source := closedRemainderPrelude + `function textOf(node: Identifier | JsxNamespacedName): string {
    return isIdentifier(node) ? node.escapedText : node.namespace.escapedText + ":" + node.name.escapedText;
}
console.log(textOf(id));
console.log(textOf(ns));
`
	_, err := lowerSourceAs(t, source, ".ts")
	var pending *NotYet
	if !errors.As(err, &pending) || !strings.Contains(err.Error(), "a read of field kind through type Node, which may hold a checked view's value") {
		t.Fatalf("want NotYet at the widened Node.kind read, got %v", err)
	}
}

// A predicate that answers false for a value it should claim stops at its own false check, on sanitized
// native and the JavaScript backend, before the false side reads the value as the remainder. Mutant: the
// false check left out (no statement appended for the remainder), and the program reads an Identifier as
// a StringLiteral.
func TestClosedRemainderLieStopsAtTheFalseCheck(t *testing.T) {
	t.Parallel()
	source := closedRemainderPrelude + `function isIdentifierLie(node: Node): node is Identifier { return node.kind === SyntaxKind.StringLiteral; }
function textOf(node: Identifier | StringLiteral): string {
    return isIdentifierLie(node) ? node.escapedText : node.text;
}
console.log(textOf(id));
`
	overloadCheckStops(t, source, "", "predicate isIdentifierLie", "false branch")
}

// Compiler's seam: the second call's argument is the first call's false-narrowed type, and the first call's
// own false check holds, since this call reads its argument on the false path (#39wtqdj). A lie in the first
// predicate stops there, not at the second.
func TestClosedRemainderFirstLieStopsFirst(t *testing.T) {
	t.Parallel()
	source := closedRemainderPrelude + `function isIdentifierLie(node: Node): node is Identifier { return node.kind === SyntaxKind.StringLiteral; }
function textOf(node: Identifier | StringLiteral | NumericLiteral): string {
    if (!isIdentifierLie(node)) {
        if (isStringLiteral(node)) {
            return "string " + node.text;
        }
        return "number " + node.text;
    }
    return node.escapedText;
}
console.log(textOf(id));
`
	overloadCheckStops(t, source, "", "predicate isIdentifierLie", "false branch")
}

// Language's phantom-brand target: tsc keeps Identifier on the false side, since Identifier isn't assignable
// to TaggedNode, so the check is against the checker's own remainder, and an honest untagged Identifier passes
// it. Judged by "the members whose kind the target names", it would fail an honest program.
func TestClosedRemainderUsesTheCheckersNarrowing(t *testing.T) {
	t.Parallel()
	source := closedRemainderPrelude + `interface TaggedNode extends Identifier { readonly __tagged: undefined }
function isTagged(node: Node): node is TaggedNode { return false; }
function textOf(node: Identifier | StringLiteral): string {
    if (isTagged(node)) {
        return "tagged";
    }
    return node.kind === SyntaxKind.Identifier ? node.escapedText : node.text;
}
console.log(textOf(id));
console.log(textOf(str));
`
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "a\ns\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want a and s, as Node prints", code, stdout, stderr)
	}
}

// What the rule can't tell apart by tag stays "over an open tag domain".
func TestClosedRemainderKeepsRefusing(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		// Two members whose kinds are an alias pair, one value (Lower's condition). Mutant: the distinctness
		// test dropped, and this lowers past it. The checker gives both members one literal type, so comparing
		// by member and by value agree here.
		{"a SyntaxKind alias pair", `interface FirstName extends Node { readonly kind: SyntaxKind.FirstName; readonly first: string }
function textOf(node: Identifier | FirstName | StringLiteral): string { return isIdentifier(node) ? node.escapedText : (node.kind === SyntaxKind.StringLiteral ? node.text : String(node.pos)); }
console.log(textOf(id));
`},
		// A member whose kind is a getter (Adamic's pin): the test and a later read could disagree.
		{"a getter kind", `class Getter { get kind(): SyntaxKind.StringLiteral { return SyntaxKind.StringLiteral; } readonly pos = 0; readonly text = "g"; }
function textOf(node: Identifier | Getter): string { return isIdentifier(node) ? node.escapedText : node.text; }
console.log(textOf(id));
`},
		// A remainder tag a member the false side excludes also holds: the check would pass for the excluded
		// member.
		{"a remainder tag the excluded member holds", `interface Twin extends Node { readonly kind: SyntaxKind.Identifier; readonly twin: number }
function isTwin(node: Node): node is Twin { return false; }
function textOf(node: Identifier | Twin): string { return isTwin(node) ? String(node.twin) : node.escapedText; }
console.log(textOf(id));
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, closedRemainderPrelude+probe.source, ".ts")
			var pending *NotYet
			if !errors.As(err, &pending) || !strings.Contains(err.Error(), "checked predicate membership over an open tag domain") {
				t.Fatalf("want NotYet over an open tag domain, got %v", err)
			}
		})
	}
}

// Language's property path: tsc keeps the false narrowing of holder.name across a call that may write it.
// Whether the remainder is read by tag or at a fixed layout decides whether this is a stop or a hole, so the
// pin asks only that it never reads memory wrongly: NotYet at compile time, or a loud stop natively.
func TestClosedRemainderPropertyPathNeverMisreads(t *testing.T) {
	t.Parallel()
	source := closedRemainderPrelude + `interface Holder { name: Identifier | JsxNamespacedName }
function touch(holder: Holder): void { holder.name = id; }
function textOf(holder: Holder): string {
    if (!isIdentifier(holder.name)) {
        touch(holder);
        return holder.name.namespace.escapedText;
    }
    return "identifier";
}
console.log(textOf({ name: ns }));
`
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		var pending *NotYet
		var refused *Refused
		if !errors.As(err, &pending) && !errors.As(err, &refused) {
			t.Fatal(err)
		}
		t.Logf("stops at compile time: %v", err)
		return
	}
	code, stdout, stderr := brandedNative(t, program)
	if code != 70 || strings.Contains(stderr, "AddressSanitizer") {
		t.Fatalf("native: exit %d, stdout %q, stderr %.600q; want a loud stop at exit 70, never a misread", code, stdout, stderr)
	}
	t.Logf("stops natively: %.300q", stderr)
}

// An open remainder (Expression's kind is the whole SyntaxKind) never reaches this rule: the refinement is
// refused first as an open numeric enum used as a literal tag, so it's held fail closed there.
func TestClosedRemainderOpenRemainderRefusedFirst(t *testing.T) {
	t.Parallel()
	source := closedRemainderPrelude + `interface Expression extends Node { readonly _expressionBrand: undefined }
function textOf(node: Identifier | Expression): string { return isIdentifier(node) ? node.escapedText : String(node.pos); }
console.log(textOf(id));
`
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "an object refinement using an open numeric enum as a literal tag") {
		t.Fatalf("want the open enum refinement refused, got %v", err)
	}
}
