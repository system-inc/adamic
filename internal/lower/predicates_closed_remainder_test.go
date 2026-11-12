package lower

import (
	"errors"
	"strings"
	"testing"
)

// closedRemainderPrelude is tsc's shape: an open SyntaxKind on a non-union Node, and node kinds that extend
// it with one literal kind each, so `isX(node: Node)` is a checked predicate over an open tag domain. The
// enum carries a value alias, as SyntaxKind's FirstToken and LastKeyword do.
const closedRemainderPrelude = `enum SyntaxKind { StringLiteral = 11, Identifier = 80, JsxNamespacedName = 295, FirstName = Identifier }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface Identifier extends Node { readonly kind: SyntaxKind.Identifier; readonly escapedText: string }
interface StringLiteral extends Node { readonly kind: SyntaxKind.StringLiteral; readonly text: string }
interface JsxNamespacedName extends Node { readonly kind: SyntaxKind.JsxNamespacedName; readonly namespace: Identifier; readonly name: Identifier }
function isIdentifier(node: Node): node is Identifier { return node.kind === SyntaxKind.Identifier; }
function isStringLiteral(node: Node): node is StringLiteral { return node.kind === SyntaxKind.StringLiteral; }
const id: Identifier = { kind: SyntaxKind.Identifier, pos: 0, escapedText: "a" };
const ns: JsxNamespacedName = { kind: SyntaxKind.JsxNamespacedName, pos: 1, namespace: { kind: SyntaxKind.Identifier, pos: 2, escapedText: "ns" }, name: { kind: SyntaxKind.Identifier, pos: 3, escapedText: "b" } };
const str: StringLiteral = { kind: SyntaxKind.StringLiteral, pos: 4, text: "s" };
`

// tsc's getEscapedTextOfJsxAttributeName (utilities.ts:11129): isIdentifier over an argument held as the
// closed union Identifier | JsxNamespacedName, whose false side reads the other member. The false side is
// checked against the argument's own remainder, so the call lowers where it was "checked predicate
// membership over an open tag domain" (#w3e1vr8). Mutant: checkedPredicateClosedRemainder answering nil,
// and this stops there again.
func TestClosedRemainderServesTscsShape(t *testing.T) {
	t.Parallel()
	source := closedRemainderPrelude + `function textOf(node: Identifier | JsxNamespacedName): string {
    return isIdentifier(node) ? node.escapedText : node.namespace.escapedText + ":" + node.name.escapedText;
}
console.log(textOf(id));
console.log(textOf(ns));
`
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "a\nns:b\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want a and ns:b, as Node prints", code, stdout, stderr)
	}
}

// A predicate that answers false for a value it should claim stops at its own false check, on sanitized
// native and the JavaScript backend, before the false side reads the value as the remainder. Mutant: the
// false check left out (no statement appended for the remainder), and the program reads an Identifier as
// a JsxNamespacedName.
func TestClosedRemainderLieStopsAtTheFalseCheck(t *testing.T) {
	t.Parallel()
	source := closedRemainderPrelude + `function isIdentifierLie(node: Node): node is Identifier { return node.kind === SyntaxKind.StringLiteral; }
function textOf(node: Identifier | JsxNamespacedName): string {
    return isIdentifierLie(node) ? node.escapedText : node.namespace.escapedText;
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
function textOf(node: Identifier | StringLiteral | JsxNamespacedName): string {
    if (!isIdentifierLie(node)) {
        if (isStringLiteral(node)) {
            return node.text;
        }
        return node.namespace.escapedText;
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
		// An open remainder: Expression's kind is the whole SyntaxKind.
		{"an open remainder", `interface Expression extends Node { readonly _expressionBrand: undefined }
function textOf(node: Identifier | Expression): string { return isIdentifier(node) ? node.escapedText : String(node.pos); }
console.log(textOf(id));
`},
		// Two members whose kinds are an alias pair, one value (Lower's condition). Mutant: tags compared by
		// enum member, and this lowers.
		{"a SyntaxKind alias pair", `interface FirstName extends Node { readonly kind: SyntaxKind.FirstName; readonly first: string }
function textOf(node: Identifier | FirstName | JsxNamespacedName): string { return isIdentifier(node) ? node.escapedText : (node.kind === SyntaxKind.JsxNamespacedName ? node.namespace.escapedText : String(node.pos)); }
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
