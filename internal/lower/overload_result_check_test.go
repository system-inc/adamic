package lower

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// overloadCheckStops runs program natively under the sanitizers and on the JavaScript backend, and each
// must stop at exit 70 with stdout before the stop being want and stderr naming the overload's check.
func overloadCheckStops(t *testing.T, source, want string, checks ...string) {
	t.Helper()
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatalf("want it lowered with the call's result check, got %v", err)
	}
	directory := t.TempDir()
	binary := filepath.Join(directory, "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	generated := filepath.Join(directory, "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, command := range []*exec.Cmd{exec.Command(binary), exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated)} {
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		var out, errs bytes.Buffer
		command.Stdout, command.Stderr = &out, &errs
		code := 0
		if err := command.Run(); err != nil {
			var exit *exec.ExitError
			if !errors.As(err, &exit) {
				t.Fatal(err)
			}
			code = exit.ExitCode()
		}
		named := true
		for _, check := range checks {
			named = named && strings.Contains(errs.String(), check)
		}
		if code != 70 || out.String() != want || !named {
			t.Fatalf("%s: exit %d stdout %q stderr %q; want 70, %q and %q", command.Path, code, out.String(), errs.String(), want, checks)
		}
	}
}

// tsc's nodeCanBeDecorated (utilities.ts:3471), its first overload promising `true` for a class
// declaration over an implementation that answers boolean, with its body cut to the kind switch.
const overloadLiteralShape = `interface N { readonly kind: number }
interface ClassDeclaration extends N { readonly kind: 263 }
interface ClassExpression extends N { readonly kind: 231 }
function canDecorate(legacy: boolean, node: ClassDeclaration): true;
function canDecorate(legacy: boolean, node: ClassExpression): boolean;
function canDecorate(legacy: boolean, node: N): boolean;
function canDecorate(legacy: boolean, node: N): boolean {
    switch (node.kind) {
        case 263: return BODY;
        case 231: return !legacy;
    }
    return false;
}
const declaration: ClassDeclaration = { kind: 263 };
const expression: ClassExpression = { kind: 231 };
const yes: true = canDecorate(true, declaration);
console.log(String(yes));
console.log(String(canDecorate(true, expression)));
`

// The honest body serves the literal its first overload promises, and agrees with Node natively.
func TestOverloadLiteralResultIsChecked(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNodeAs(t, strings.Replace(overloadLiteralShape, "BODY", "true", 1), ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "true\nfalse\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want true and false, as Node prints", code, stdout, stderr)
	}
}

// A body that breaks the first overload's promise stops at the call that resolved to it, before the
// value reaches the `true` slot, on both backends. Mutant: censusOverload accepting the declaration
// without overloadResultCheck's wrap, and this program prints false into a true.
func TestOverloadLiteralResultLieStopsAtTheCall(t *testing.T) {
	t.Parallel()
	overloadCheckStops(t, strings.Replace(overloadLiteralShape, "BODY", "false", 1), "", "overload 1 of canDecorate result at ", ": expected true")
}

// The check follows the promise, not the path: a call resolving to the narrow overload is checked even
// where the second overload's caller would have taken any boolean.
func TestOverloadLiteralResultFollowsThePromise(t *testing.T) {
	t.Parallel()
	source := strings.Replace(overloadLiteralShape, "BODY", "true", 1) + "console.log(String(canDecorate(false, { kind: 263 } as ClassDeclaration)));\n"
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// A tagged union's member, as tsc's getThisContainer and visitFunctionBody promise one: the result is
// read by its tag, and a lie stops at the call.
const overloadTagShape = `interface Block { readonly kind: 242; readonly statements: number }
interface Arrow { readonly kind: 220; readonly body: number }
type Body = Block | Arrow;
function bodyOf(node: Block): Block;
function bodyOf(node: Body): Body;
function bodyOf(node: Body): Body {
    return PICK;
}
const block: Block = { kind: 242, statements: 3 };
const arrow: Arrow = { kind: 220, body: 4 };
const got = bodyOf(block);
console.log(String(got.statements));
`

func TestOverloadTagResultIsChecked(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNodeAs(t, strings.Replace(overloadTagShape, "PICK", "node", 1), ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "3\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 3, as Node prints", code, stdout, stderr)
	}
}

func TestOverloadTagResultLieStopsAtTheCall(t *testing.T) {
	t.Parallel()
	overloadCheckStops(t, strings.Replace(overloadTagShape, "PICK", "node.kind === 242 ? arrow : node", 1), "", "overload 1 of bodyOf result at ", ": expected Block")
}

// An open numeric tag, as tsc's Node holds kind: SyntaxKind: the tag can't prove the layout, so the
// result is read through the promised member's view from the call on, and the lie is stored in a local
// of the wider type before its read, so a view attached only where the call's value is used directly
// would let it through (Lower's condition).
const overloadOpenTagShape = `const enum Kind { Block = 242, Arrow = 220, Other = 1 }
interface Node { readonly kind: Kind }
interface Block extends Node { readonly kind: Kind.Block; readonly statements: number }
function bodyOf(node: Block): Block;
function bodyOf(node: Node): Node;
function bodyOf(node: Node): Node {
    return PICK;
}
const block: Block = { kind: Kind.Block, statements: 3 };
const liar: Node = { kind: Kind.Block };
const held: Node = bodyOf(block);
if (held.kind === Kind.Block) {
    console.log(String((held as Block).statements));
}
`

func TestOverloadOpenTagResultIsChecked(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeAs(t, strings.Replace(overloadOpenTagShape, "PICK", "node", 1), ".ts")
}

// The lie under an open tag: the body hands back a Node whose kind says Block but which has no
// statements. The tag matches, so the call's check passes, and the result goes into a union-typed local
// before it's narrowed back to Block and read. Block's view, registered at the call, stops that read on
// both backends.
func TestOverloadOpenTagResultLieStopsAtTheRead(t *testing.T) {
	t.Parallel()
	source := `const enum Kind { Block = 242, Arrow = 220 }
interface Node { readonly kind: Kind }
interface Block extends Node { readonly kind: Kind.Block; readonly statements: number }
interface Arrow extends Node { readonly kind: Kind.Arrow; readonly body: number }
const liar: Node = { kind: Kind.Block };
function bodyOf(node: Block): Block;
function bodyOf(node: Node): Node;
function bodyOf(node: Node): Node {
    return node.kind === Kind.Block ? liar : node;
}
const block: Block = { kind: Kind.Block, statements: 3 };
const wide: Block | Arrow = bodyOf(block);
if (wide.kind === Kind.Block) {
    console.log(String(wide.statements));
}
`
	overloadCheckStops(t, source, "", "statements")
}

// tsc's skipParentheses (utilities.ts:5041): Expression from Node, a branded supertype. The call checks
// that the result is an object, and Expression's view checks every field where it's read.
const overloadBrandedShape = `const enum Kind { Identifier = 80, Paren = 217, Block = 242 }
interface Node { readonly kind: Kind; readonly parent: Node | undefined }
interface Expression extends Node { readonly _expressionBrand: undefined }
interface Identifier extends Expression { readonly kind: Kind.Identifier; readonly text: string }
interface Paren extends Expression { readonly kind: Kind.Paren; readonly expression: Expression }
function isParen(node: Node): node is Paren { return node.kind === Kind.Paren; }
function skip(node: Expression): Expression;
function skip(node: Node): Node;
function skip(node: Node): Node {
    while (isParen(node)) {
        node = node.expression;
    }
    return node;
}
const id: Identifier = { kind: Kind.Identifier, parent: undefined, text: "x", _expressionBrand: undefined };
const paren: Paren = { kind: Kind.Paren, parent: undefined, expression: id, _expressionBrand: undefined };
const inner = skip(paren);
console.log(String(inner.kind));
`

func TestOverloadBrandedResultIsChecked(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeAs(t, overloadBrandedShape, ".ts")
}

// A nested overload set that captures is called through its closure value (callClosure), a road beside
// censusOverloadResult, so its result is checked there too. Mutant: that road returning its call
// unchecked, and expression's guard panics as a compiler bug.
func TestOverloadNestedResultLieStopsAtTheCall(t *testing.T) {
	t.Parallel()
	source := `function outer(flag: boolean): void {
    function pick(x: number): 1;
    function pick(x: string): number;
    function pick(x: number | string): number {
        return flag ? 1 : 2;
    }
    const one: 1 = pick(0);
    console.log(String(one));
}
outer(true);
outer(false);
`
	overloadCheckStops(t, source, "1\n", "overload 1 of pick result at ", ": expected 1")
}

// A nested overload set that captures nothing, called from a sibling nested function: the sibling
// road (nestedSibling's ir.CallClosure with its Direct target) checks the result as the others do.
func TestOverloadNestedSiblingResultLieStopsAtTheCall(t *testing.T) {
	t.Parallel()
	source := `function outer(): void {
    function pick(x: number): 1;
    function pick(x: string): number;
    function pick(x: number | string): number {
        return x === 0 ? 1 : 2;
    }
    function use(x: number): void {
        const one: 1 = pick(x);
        console.log(String(one));
    }
    use(0);
    use(1);
}
outer();
`
	overloadCheckStops(t, source, "1\n", "overload 1 of pick result at ", ": expected 1")
}

// An overload set read as a value has no call to wrap, so it stops where it's read, as main does
// (expression.go's "an overloaded function as a value"): through a const, a conditional of two sets,
// .call and .bind. If that stop ever lifts, the value's calls need the same check.
func TestOverloadNarrowAsAValueStops(t *testing.T) {
	t.Parallel()
	const set = "function one(x: number): 1;\nfunction one(x: string): number;\nfunction one(x: number | string): number { return 1; }\nfunction two(x: number): 1;\nfunction two(x: string): number;\nfunction two(x: number | string): number { return 1; }\n"
	for _, probe := range []struct{ name, use string }{
		{"const", "const f = one;\nconsole.log(String(f(0)));\n"},
		{"union of two sets", "function choose(flag: boolean): void {\n    const f = flag ? one : two;\n    console.log(String(f(0)));\n}\nchoose(true);\n"},
		{"call", "console.log(String(one.call(undefined, \"a\")));\n"},
		{"bind", "const bound = one.bind(undefined);\nconsole.log(String(bound(0)));\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, set+probe.use, ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want a stop where the overload set is read as a value, got %v", err)
			}
			t.Logf("%s stops: %v", probe.name, err)
		})
	}
}

// In .a the narrow overload stays refused at its declaration: a .a file never holds a hidden check.
func TestOverloadNarrowResultStaysRefusedInA(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, strings.Replace(overloadLiteralShape, "BODY", "true", 1), ".a")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "overload 1 of canDecorate result true cannot be served by implementation result boolean") {
		t.Fatalf("want overload 1 refused at its declaration, got %v", err)
	}
}

// A promise the call can't check stays refused at the declaration (rulings.md:126): an array's elements,
// as tsc's filter promises U[], and an untagged interface subtype, as createBuilderProgram promises one.
func TestOverloadUncheckableResultStaysRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, what string }{
		{"array elements", "interface A { readonly kind: 1 }\ninterface B { readonly kind: 2 }\nfunction keep(xs: readonly A[]): A[];\nfunction keep(xs: readonly (A | B)[]): (A | B)[];\nfunction keep(xs: readonly (A | B)[]): (A | B)[] { return [...xs]; }\nconsole.log(String(keep([{ kind: 1 }]).length));\n", "overload 1 of keep result A[] cannot be served"},
		{"untagged subtype", "interface P { run(): number }\ninterface Q extends P { more(): number }\nfunction make(full: true): Q;\nfunction make(full: boolean): P;\nfunction make(full: boolean): P { return { run: () => 1 }; }\nconsole.log(String(make(true).run()));\n", "overload 1 of make result Q cannot be served"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, probe.what) {
				t.Fatalf("want %q, got %v", probe.what, err)
			}
		})
	}
}
