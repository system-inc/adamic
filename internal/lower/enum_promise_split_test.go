package lower

import (
	"errors"
	"strings"
	"testing"
)

const singletonFieldRefusal = "an unproven singleton numeric enum member field in an object view"

// tsc's destructuring.ts passes an array narrowed by Debug.assertEachNode, an
// intersection, to a function taking the narrower element union. No member of
// that union takes every element type whole, so its promises are checked
// against every member; the next boundary is the assertion overload's own proof.
func TestEnumPromiseAssertedArray(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `enum SyntaxKind { Identifier, BindingElement, OmittedExpression, ArrayLiteral }
interface Identifier { readonly kind: SyntaxKind.Identifier; readonly text: string }
interface BindingElement { readonly kind: SyntaxKind.BindingElement; readonly name: string }
interface OmittedExpression { readonly kind: SyntaxKind.OmittedExpression }
interface ArrayLiteral { readonly kind: SyntaxKind.ArrayLiteral; readonly size: number }
type ArrayBindingElement = BindingElement | OmittedExpression;
type BindingOrAssignmentElement = Identifier | BindingElement | OmittedExpression | ArrayLiteral;

function isArrayBindingElement(node: BindingOrAssignmentElement): node is ArrayBindingElement {
    return node.kind === SyntaxKind.BindingElement || node.kind === SyntaxKind.OmittedExpression;
}
function assertEachNode<T extends BindingOrAssignmentElement, U extends T>(nodes: readonly T[], test: (node: T) => node is U): asserts nodes is readonly U[];
function assertEachNode(nodes: readonly BindingOrAssignmentElement[], test: (node: BindingOrAssignmentElement) => boolean): void {
    if (!nodes.every(test)) throw new Error("Unexpected node.");
}
function createArrayBindingPattern(elements: readonly ArrayBindingElement[]): string {
    let names = "";
    for (const element of elements) {
        names += element.kind === SyntaxKind.BindingElement ? element.name : "_";
    }
    return names;
}
function makeArrayBindingPattern(elements: BindingOrAssignmentElement[]): string {
    assertEachNode(elements, isArrayBindingElement);
    return createArrayBindingPattern(elements);
}
const elements: BindingOrAssignmentElement[] = [
    { kind: SyntaxKind.BindingElement, name: "a" },
    { kind: SyntaxKind.OmittedExpression },
    { kind: SyntaxKind.BindingElement, name: "c" },
];
console.log(makeArrayBindingPattern(elements));
`)
	if err == nil || strings.Contains(err.Error(), singletonFieldRefusal) || !strings.Contains(err.Error(), "overload 1 of assertEachNode parameter test") {
		t.Fatalf("want the assertion overload refusal, not the singleton refusal: %v", err)
	}
}

// Split across a union, the value can be any member its tag selects, so a
// member whose field promises a singleton member still needs a proof.
func TestEnumPromiseSplitUnionKeepsEveryMember(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `enum K { A, B }
enum One { Only }
interface X { readonly kind: K.A; readonly s: One.Only }
interface Y { readonly kind: K.B; readonly s: One }
interface From { readonly kind: K.A | K.B; readonly s: One }
function view(value: From): X | Y { return value; }
const seen = view({ kind: K.A, s: One.Only });
console.log(String(seen.kind));
`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), singletonFieldRefusal) {
		t.Fatalf("want the singleton refusal for a split union, got %v", err)
	}
}
