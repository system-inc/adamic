package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// overloadKindExtractIntersectionSource is tsc's transformNamedEvaluation (transformers/namedEvaluation.ts:469
// at the stage 3 pin) cut down, #0kkpw2e's x2 probe: an overload promising Extract<NamedEvaluation, Pick<T,
// "kind" | keyof T & "operatorToken" | keyof T & "name">> whose NamedEvaluation members are intersections,
// over an implementation returning the plain node types. Its overload is at line 64.
func overloadKindExtractIntersectionSource(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("testdata", "overload_kind_extract", "transform_named_evaluation.ts"))
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

// In .ts the overload stops as its own gap, not as the covariance refusal, in Adamic's words for it
// (#0kkpw2e set 2, ruling (c)), at the first call that resolves to it, line 110 (#4czdn4q).
func TestOverloadKindExtractIntersectionIsNotYet(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, overloadKindExtractIntersectionSource(t), ".ts")
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != "an overload result Extract over intersection members the implementation's result doesn't serve" || !strings.HasSuffix(notYet.Where, "main.ts:110:12") {
		t.Fatalf("want the call at main.ts:110:12 to the overload not yet lowered as an Extract over intersection members, got %v", err)
	}
}

// In .a the same program stays refused at its declaration for its result, as on main: a .a file never holds
// a hidden gap. Mutant: the branch taken in .a too, and this row reds.
func TestOverloadKindExtractIntersectionStaysRefusedInA(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, overloadKindExtractIntersectionSource(t), ".a")
	overloadKindExtractRefused(t, err, "transformNamedEvaluation")
	if !strings.HasSuffix(err.Error(), "cannot be served by implementation result BinaryExpression | PropertyAssignment | VariableDeclaration; make the implementation result covariant with every overload result") {
		t.Fatalf("want the covariance refusal over the implementation's three plain node types, got %v", err)
	}
}

// A control: a narrower union with an intersection member that isn't an Extract, which no call can check
// (overloadResultCheckable reads no intersection), stays the covariance refusal in .ts.
func TestOverloadKindExtractIntersectionControlStaysRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, `enum SyntaxKind {
    Unknown = 0,
    Alpha = 1,
    Beta = 2,
    Gamma = 3,
}
interface Node {
    readonly kind: SyntaxKind;
}
interface Alpha extends Node {
    readonly kind: SyntaxKind.Alpha;
    readonly alpha: string;
}
interface Beta extends Node {
    readonly kind: SyntaxKind.Beta;
    readonly beta: number;
}
interface Gamma extends Node {
    readonly kind: SyntaxKind.Gamma;
    readonly gamma: string;
}
type Shape = Alpha | Beta | Gamma;
function rebuild(node: Shape): (Alpha & { readonly extra: string; }) | Beta;
function rebuild(node: Shape) {
    return node;
}
const beta: Beta = { kind: SyntaxKind.Beta, beta: 2 };
console.log(String(rebuild(beta).kind));
`, ".ts")
	overloadKindExtractRefused(t, err, "rebuild")
}
