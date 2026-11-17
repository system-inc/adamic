package lower

import (
	"errors"
	"strings"
	"testing"
)

// tsc's injectClassNamedEvaluationHelperBlockIfMissing (transformers/namedEvaluation.ts:173 at the stage 3
// pin), its three callers tsc's own (esDecorators.ts:1062, :676 and namedEvaluation.ts:250): a generic overload
// promising Extract<ClassLikeDeclaration, Pick<T, "kind">> over a non-generic implementation that returns its
// parameter, tests a member and reassigns the parameter by its kind. UPDATE is that reassignment.
const overloadKindExtractShape = `enum SyntaxKind {
    Unknown = 0,
    ClassExpression = 231,
    ClassDeclaration = 263,
}
interface Node {
    readonly kind: SyntaxKind;
}
interface ClassDeclaration extends Node {
    readonly kind: SyntaxKind.ClassDeclaration;
    readonly name: string;
    readonly members: number;
}
interface ClassExpression extends Node {
    readonly kind: SyntaxKind.ClassExpression;
    readonly name: string | undefined;
    readonly members: number;
}
type ClassLikeDeclaration = ClassDeclaration | ClassExpression;
const anonymous: ClassExpression = { kind: SyntaxKind.ClassExpression, name: undefined, members: 9 };
function injectClassNamedEvaluationHelperBlockIfMissing<T extends ClassLikeDeclaration>(node: T, assignedName: string): Extract<ClassLikeDeclaration, Pick<T, "kind">>;
function injectClassNamedEvaluationHelperBlockIfMissing(node: ClassLikeDeclaration, assignedName: string) {
    if (classHasExplicitlyAssignedName(node)) {
        return node;
    }
    let members = node.members + 1;
    if (node.name !== undefined) {
        members = members + assignedName.length;
    }
    node = UPDATE;
    return node;
}
function classHasExplicitlyAssignedName(node: ClassLikeDeclaration): boolean {
    return node.members > 9;
}
function updateClassDeclaration(node: ClassDeclaration, members: number): ClassDeclaration {
    return { kind: SyntaxKind.ClassDeclaration, name: node.name, members };
}
function updateClassExpression(node: ClassExpression, members: number): ClassExpression {
    return { kind: SyntaxKind.ClassExpression, name: node.name, members };
}
function either(pickDeclaration: boolean, declaration: ClassDeclaration, expression: ClassExpression): ClassLikeDeclaration {
    return pickDeclaration ? declaration : expression;
}
const declaration: ClassDeclaration = { kind: SyntaxKind.ClassDeclaration, name: "Declared", members: 2 };
const expression: ClassExpression = { kind: SyntaxKind.ClassExpression, name: undefined, members: 3 };
let declared: ClassDeclaration = declaration;
declared = injectClassNamedEvaluationHelperBlockIfMissing(declared, "C");
let like = either(false, declaration, expression);
like = injectClassNamedEvaluationHelperBlockIfMissing(like, "default");
const injected = injectClassNamedEvaluationHelperBlockIfMissing(expression, "E");
console.log(` + "`${declared.kind} ${declared.name} ${declared.members}`" + `);
console.log(` + "`${like.kind} ${like.members}`" + `);
console.log(` + "`${injected.kind} ${injected.name === undefined ? \"anonymous\" : injected.name} ${injected.members}`" + `);
`

const overloadKindExtractHonest = "node.kind === SyntaxKind.ClassDeclaration ? updateClassDeclaration(node, members) : updateClassExpression(node, members)"

// The honest body lowers, each of its three calls checked against its own instance of the Extract (a
// ClassDeclaration, the whole ClassLikeDeclaration, a ClassExpression), and agrees with Node on the
// JavaScript backend and natively (#0kkpw2e set 1).
func TestOverloadKindExtractIsChecked(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNodeAs(t, strings.Replace(overloadKindExtractShape, "UPDATE", overloadKindExtractHonest, 1), ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "263 Declared 4\n231 4\n231 anonymous 4\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want what Node prints", code, stdout, stderr)
	}
	if counts := program.OverloadResultChecks; counts.Checked != 3 || counts.Erased != 0 || counts.Proven != 0 {
		t.Fatalf("want each of the three read results checked, got %+v", counts)
	}
}

// A body that hands back the other class kind: Node prints the ClassExpression the ClassDeclaration caller
// was promised, and both backends stop at that call before the value lands in its slot.
func TestOverloadKindExtractLieStopsAtTheCall(t *testing.T) {
	t.Parallel()
	source := strings.Replace(overloadKindExtractShape, "UPDATE", "node.kind === SyntaxKind.ClassDeclaration ? anonymous : updateClassExpression(node, members)", 1)
	if node := runAgreementNode(t, writeAgreementSource(t, "main.ts", source)); node.code != 0 || string(node.stdout) != "231 undefined 9\n231 4\n231 anonymous 4\n" {
		t.Fatalf("Node: exit %d, stdout %q; want the lie printed", node.code, node.stdout)
	}
	overloadCheckStops(t, source, "", "overload 1 of injectClassNamedEvaluationHelperBlockIfMissing result at ", ": expected ClassDeclaration")
}

// In .a the overload stays refused at its declaration: a .a file never holds a hidden check.
func TestOverloadKindExtractStaysRefusedInA(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, strings.Replace(overloadKindExtractShape, "UPDATE", overloadKindExtractHonest, 1), ".a")
	overloadKindExtractRefused(t, err, "injectClassNamedEvaluationHelperBlockIfMissing")
}

// Shapes whose kind can't say which member a value is stay refused at the declaration (census_small.go's
// last refusal), as on main: two members sharing a kind's value (one enum's aliases, which tsc gives one
// literal type, with no kind test in the body so family 1's tag walk isn't what stops it; mutant: the
// distinct-value test dropped, and this one lowers), a member whose kind is a union of two literals, and
// classes whose kind is a getter.
func TestOverloadKindExtractUnsoundTagsStayRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, function, source string }{
		{"shared value", "rebuild", `enum SyntaxKind {
    Unknown = 0,
    Alpha = 1,
    Beta = 2,
    Gamma = 2,
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
const gamma: Gamma = { kind: SyntaxKind.Gamma, gamma: "gamma" };
function rebuild<T extends Shape>(node: T, swap: boolean): Extract<Shape, Pick<T, "kind">>;
function rebuild(node: Shape, swap: boolean) {
    return swap ? gamma : node;
}
const beta: Beta = { kind: SyntaxKind.Beta, beta: 2 };
const rebuilt = rebuild(beta, true);
console.log(String(rebuilt.kind));
`},
		{"kind union", "flip", `enum SyntaxKind {
    Unknown = 0,
    Plus = 1,
    Minus = 2,
    Name = 3,
}
interface Node {
    readonly kind: SyntaxKind;
}
interface Operator extends Node {
    readonly kind: SyntaxKind.Plus | SyntaxKind.Minus;
    readonly symbol: string;
}
interface Name extends Node {
    readonly kind: SyntaxKind.Name;
    readonly text: string;
}
type Shape = Operator | Name;
function flip<T extends Shape>(node: T): Extract<Shape, Pick<T, "kind">>;
function flip(node: Shape) {
    switch (node.kind) {
        case SyntaxKind.Plus:
            return makeOperator(SyntaxKind.Minus, "-");
        case SyntaxKind.Minus:
            return makeOperator(SyntaxKind.Plus, "+");
        case SyntaxKind.Name:
            return makeName(node.text);
    }
}
function makeOperator(kind: SyntaxKind.Plus | SyntaxKind.Minus, symbol: string): Operator {
    return { kind, symbol };
}
function makeName(text: string): Name {
    return { kind: SyntaxKind.Name, text };
}
const plus: Operator = { kind: SyntaxKind.Plus, symbol: "+" };
console.log(flip(plus).symbol);
`},
		{"getter kind", "renew", `enum SyntaxKind {
    Unknown = 0,
    Alpha = 1,
    Beta = 2,
}
let reads = 0;
class AlphaNode {
    readonly alpha: string;
    constructor(alpha: string) {
        this.alpha = alpha;
    }
    get kind(): SyntaxKind.Alpha {
        reads = reads + 1;
        return SyntaxKind.Alpha;
    }
}
class BetaNode {
    readonly beta: number;
    constructor(beta: number) {
        this.beta = beta;
    }
    get kind(): SyntaxKind.Beta {
        reads = reads + 1;
        return SyntaxKind.Beta;
    }
}
type Shape = AlphaNode | BetaNode;
function renew<T extends Shape>(node: T): Extract<Shape, Pick<T, "kind">>;
function renew(node: Shape) {
    switch (node.kind) {
        case SyntaxKind.Alpha:
            return new AlphaNode(node.alpha);
        case SyntaxKind.Beta:
            return new BetaNode(node.beta + 1);
    }
}
console.log(renew(new AlphaNode("a")).alpha);
console.log(String(reads));
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			overloadKindExtractRefused(t, err, probe.function)
		})
	}
}

// overloadKindExtractRefused fails unless err is overload 1 of function refused at its declaration for its
// result.
func overloadKindExtractRefused(t *testing.T, err error, function string) {
	t.Helper()
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "overload 1 of "+function+" result ") || !strings.Contains(refused.What, "cannot be served by implementation result") {
		t.Fatalf("want overload 1 of %s refused at its declaration for its result, got %v", function, err)
	}
}
