package lower

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// A claim on a function type is checked at every call through it in .ts (#1w5bve2, agreed Oct 30):
// a parameter's, an interface method's, and an overload's whose implementation checks its own. The
// lies are asserts claims filled by a function returning void, which tsc relates to an asserts
// signature, so nothing but the call's check stands between the claim and the read after it. Each
// stops at that call with the located message, on sanitized native and the JavaScript backend, where
// Node prints undefined. Honest programs agree with Node, a generic claim checked against the target
// its call resolved to among them. Mutant: the call through a function type left unchecked, which
// reddens every lie.
func TestFunctionTypeClaimIsCheckedAtTheCall(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"param_honest", "iface_honest", "trycast"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			source, err := os.ReadFile("testdata/predicates_function_type/" + name + ".ts")
			if err != nil {
				test.Fatal(err)
			}
			lowersAndAgreesWithNodeAs(test, string(source), ".ts")
		})
	}
	for _, lie := range []struct{ name, callee, where string }{
		{"param_asserts_lie", "assertIdentifier", ":3:108"},
		{"iface_asserts_lie", "a.assertIdentifier", ":5:48"},
		{"overload_checked_impl", "assert", ":4:101"},
	} {
		test.Run(lie.name, func(test *testing.T) {
			test.Parallel()
			source, err := os.ReadFile("testdata/predicates_function_type/" + lie.name + ".ts")
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowerTypeScriptAssertionSource(test, string(source))
			if err != nil {
				test.Fatal(err)
			}
			directory := test.TempDir()
			binary := filepath.Join(directory, "native")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				test.Fatal(err)
			}
			runner, err := filepath.Abs("../../oracle/node.mjs")
			if err != nil {
				test.Fatal(err)
			}
			generated := filepath.Join(directory, "generated.mjs")
			if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
				test.Fatal(err)
			}
			prefix := "adamic: panic: predicate " + lie.callee + " at "
			suffix := lie.where + " asserts branch: source Node, target Identifier\n"
			for _, command := range []*exec.Cmd{
				exec.Command(binary),
				exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated),
			} {
				command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
				var out, errs bytes.Buffer
				command.Stdout, command.Stderr = &out, &errs
				code := 0
				if err := command.Run(); err != nil {
					var exit *exec.ExitError
					if !errors.As(err, &exit) {
						test.Fatal(err)
					}
					code = exit.ExitCode()
				}
				if code != 70 ||
					out.Len() != 0 ||
					!strings.HasPrefix(errs.String(), prefix) ||
					!strings.HasSuffix(errs.String(), suffix) {
					test.Fatalf(
						"%s: exit %d stdout %q stderr %q; want 70 and %q...%q",
						command.Path,
						code,
						out.String(),
						errs.String(),
						prefix,
						suffix,
					)
				}
			}
		})
	}
}

// What the check at the call can't reach keeps its refusal. A library's claim is called inside the
// library, so a checked parameter forwarded to filter or find keeps predicateArguments' proof (Lower's
// must). An overload whose implementation takes a plain callback calls it unchecked, so the overload's
// argument keeps its proof too: overload_optional_impl writes the callback or undefined out, which
// predicateParameter trusts as a plain callback, and overload_union_impl names it through an alias,
// which plainCallbackType doesn't read, so only checkedFunctionTypeClaim's overload rule keeps its
// argument's proof. A target with no runtime test is NotYet at the call, never a vacuous check. .a
// needs proof and refuses the claim. Mutants: the library test dropped, which admits the forwards; the
// overload's implementation rule dropped, which lets overload_union_impl's argument through to the
// census proof's refusal of its result.
func TestFunctionTypeClaimKeepsWhatTheCallCantCheck(test *testing.T) {
	test.Parallel()
	read := func(name string) string {
		source, err := os.ReadFile("testdata/predicates_function_type/" + name + ".ts")
		if err != nil {
			test.Fatal(err)
		}
		return string(source)
	}
	for _, refusal := range []struct{ name, what string }{
		{"forward_filter", `an unproven predicate argument for parameter predicate (argument "test")`},
		{"forward_find", `an unproven predicate argument for parameter predicate (argument "test")`},
		{
			"overload_union_impl",
			`an unproven predicate argument for parameter assert (argument "(node: Node): void => {}")`,
		},
		{
			"overload_optional_impl",
			`an unproven predicate argument for parameter assert (argument "(node: Node): void => {}")`,
		},
	} {
		_, err := lowerTypeScriptAssertionSource(test, read(refusal.name))
		var refused *Refused
		if !errors.As(err, &refused) || refused.What != refusal.what {
			test.Errorf("%s: want refused as %q in .ts, got %v", refusal.name, refusal.what, err)
		}
	}
	// A method's check reads the method off the receiver after the arguments, and JavaScript reads it
	// before them, so arguments that run code stop it: method_order swaps the method in an argument, and
	// Node calls the old one (Lower's must 1); getter_order swaps it in a class's getter, read through a
	// structural field. A union of callables gets one signature from tsc, with one member's declaration,
	// so a narrowing call through it stops too (Lower's must 2). Mutants: the argument test dropped, where
	// native prints "new swapped" and Node "old swapped"; a field read counted as quiet, where native
	// prints "new" and Node "old"; the union test dropped, where union.ts lowers.
	for _, pending := range []struct{ name, what string }{
		{"untagged", "predicate overload target Named"},
		{"method_order", "a checked method claim whose arguments can run code before the method is read"},
		{"getter_order", "a checked method claim whose arguments can run code before the method is read"},
		{"union", "a call narrowing through a union or intersection of callables"},
	} {
		_, err := lowerTypeScriptAssertionSource(test, read(pending.name))
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), pending.what) {
			test.Errorf("%s: want NotYet with %q, got %v", pending.name, pending.what, err)
		}
	}
	_, err := lowerSource(test, read("iface_honest"))
	var refused *Refused
	if !errors.As(err, &refused) ||
		refused.What != "a type predicate whose return is not proven (there is no body proving this parameter)" {
		test.Errorf("iface_honest in .a: want the claim refused, got %v", err)
	}
}

// parseTreeSource is getParseTreeNode's shape (tsc's utilitiesPublic.ts:826): a generic overload whose
// optional predicate the implementation takes as a plain boolean callback, followed by calls. tsc's
// overload returns T | undefined, which an implementation returning Node | undefined is refused for at
// the overload, so this one returns Node | undefined and leaves the argument's outcome to read. The first
// %s fills nodeTest's optional mark, so the optional parameter and its required twin differ only there,
// and the calls start on the same line in both.
const parseTreeSource = `interface Node { readonly kind: number; readonly flags: number; readonly original: Node | undefined; }
interface ModuleDeclaration extends Node { readonly kind: 268; readonly name: string; }
function isModuleDeclaration(node: Node): node is ModuleDeclaration { return node.kind === 268; }
function isParseTreeNode(node: Node): boolean { return (node.flags & 16) === 0; }
function getParseTreeNode<T extends Node>(node: Node | undefined, nodeTest%s: (node: Node) => node is T): Node | undefined;
function getParseTreeNode(node: Node | undefined, nodeTest?: (node: Node) => boolean): Node | undefined {
  while (node !== undefined) {
    if (isParseTreeNode(node)) { return nodeTest === undefined || nodeTest(node) ? node : undefined; }
    node = node.original;
  }
  return undefined;
}
const parsed: ModuleDeclaration = { kind: 268, flags: 0, original: undefined, name: "parsed" };
const synthesized: ModuleDeclaration = { kind: 268, flags: 16, original: parsed, name: "synthesized" };
const other: Node = { kind: 80, flags: 0, original: undefined };
const wrapped: ModuleDeclaration = { kind: 268, flags: 16, original: other, name: "wrapped" };
%s
`

// An optional predicate parameter takes the same proven argument a required one does (#1w5bve2,
// transformers/ts.ts:1975's getParseTreeNode(nodeIn, isModuleDeclaration)): its type is the function or
// undefined, an absent argument is never called, and a present one is the function, whose claim the
// argument proves. Called with a one-line kind test and with none, it agrees with Node. Mutant: the
// parameter's call signatures read off the type with undefined left in, which refuses this as "an
// unproven predicate argument for parameter nodeTest".
func TestOptionalPredicateParameterTakesAProvenArgument(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNodeAs(
		test,
		fmt.Sprintf(parseTreeSource, "?", `const found = getParseTreeNode(synthesized, isModuleDeclaration);
console.log(found === undefined ? "none" : String(found.kind));
const unmatched = getParseTreeNode(wrapped, isModuleDeclaration);
console.log(unmatched === undefined ? "none" : String(unmatched.kind));
const untested = getParseTreeNode(synthesized);
console.log(untested === undefined ? "none" : String(untested.kind));
console.log(String(getParseTreeNode(undefined, isModuleDeclaration) === undefined));`),
		".ts",
	)
}

// The optional mark changes nothing else: each argument is admitted or refused for an optional
// parameter exactly as for its required twin, in .ts, where an arrow claiming more than its body
// proves is a checked closure, and in .a, where it's refused. A let that's reassigned can hold any
// function of its type, so it's refused in both. Mutant: the fix reverted, where every optional row
// is refused as an unproven argument and the honest and lying .ts rows stop matching their twins.
func TestOptionalPredicateParameterMatchesItsRequiredTwin(test *testing.T) {
	test.Parallel()
	outcome := func(err error) string {
		var refused *Refused
		if errors.As(err, &refused) {
			return "refused: " + refused.What
		}
		if err != nil {
			return err.Error()
		}
		return "admitted"
	}
	for _, probe := range []struct{ name, calls, typeScript, adamic string }{
		{"proven", `const found = getParseTreeNode(synthesized, isModuleDeclaration);
console.log(found === undefined ? "none" : String(found.kind));`, "admitted", "admitted"},
		{"lying arrow", `const found = getParseTreeNode(synthesized, (node: Node): node is ModuleDeclaration => true);
console.log(found === undefined ? "none" : String(found.kind));`, "admitted", `refused: an unproven predicate argument for parameter nodeTest (argument "(node: Node): node is ModuleDeclaration => true")`},
		{"reassigned let", `let test = (node: Node): node is ModuleDeclaration => node.kind === 268;
test = (node: Node): node is ModuleDeclaration => true;
const found = getParseTreeNode(synthesized, test);
console.log(found === undefined ? "none" : String(found.kind));`, `refused: an unproven predicate argument for parameter nodeTest (argument "test")`, `refused: an unproven predicate argument for parameter nodeTest (argument "test")`},
	} {
		for _, mode := range []struct{ extension, want string }{{".ts", probe.typeScript}, {".a", probe.adamic}} {
			test.Run(probe.name+mode.extension, func(test *testing.T) {
				test.Parallel()
				_, optionalErr := lowerSourceAs(test, fmt.Sprintf(parseTreeSource, "?", probe.calls), mode.extension)
				_, requiredErr := lowerSourceAs(test, fmt.Sprintf(parseTreeSource, "", probe.calls), mode.extension)
				optional, required := outcome(optionalErr), outcome(requiredErr)
				if optional != required {
					test.Fatalf("optional %q, required twin %q; want them to match", optional, required)
				}
				if optional != mode.want {
					test.Fatalf("got %q for both; want %q", optional, mode.want)
				}
			})
		}
	}
}

// assertNodeSource is debug.ts's assertNode (tsc's debug.ts:294): a generic overload asserting its node
// through a predicate parameter, a plain overload for a test that may be absent, and an implementation
// taking the callback or undefined, `((node: Node) => boolean) | undefined`, not a plain function type.
// The %s is what the overload's test takes: Node, as assertNotNode's does, or T, as assertNode's does.
const assertNodeSource = `interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
function isIdentifier(node: Node): node is Identifier { return node.kind === 80; }
function assertNode<T extends Node, U extends T>(node: T | undefined, test: (node: %s) => node is U, message?: string): asserts node is U;
function assertNode(node: Node | undefined, test: ((node: Node) => boolean) | undefined, message?: string): void;
function assertNode(node: Node | undefined, test: ((node: Node) => boolean) | undefined, message?: string): void {
  if (node === undefined || (test !== undefined && !test(node))) { console.log("failed: " + (message === undefined ? "node" : message)); }
}
const identifier: Identifier = { kind: 80, escapedText: "named" };
const node: Node = identifier;
assertNode(node, isIdentifier);
console.log(node.escapedText);
assertNode(node, undefined, "untested");
console.log(String(node.kind));
`

// visitArraySource is visitorPublic.ts's visitArrayWorker: an overload whose test claims node is U, over
// an implementation taking the same test or undefined, in parentheses. The first %s is the
// implementation's test type and the second its guard, so its plain twin, core.ts filter's shape, differs
// only there, and the calls start on the same line in both.
const visitArraySource = `interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
function isIdentifier(node: Node): node is Identifier { return node.kind === 80; }
function visitArrayWorker<U extends Node>(nodes: readonly Node[], test: (node: Node) => node is U, start: number, count: number): readonly Node[];
function visitArrayWorker(nodes: readonly Node[], test: %s, start: number, count: number): readonly Node[] {
  const visited: Node[] = [];
  for (let index = start; index < start + count; index++) { const node = nodes[index]; if (node !== undefined && (%s)) { visited.push(node); } }
  return visited;
}
const identifier: Identifier = { kind: 80, escapedText: "named" };
const nodes: readonly Node[] = [{ kind: 11 }, identifier, { kind: 80 }, { kind: 9 }];
%s
`

// An overload's predicate parameter over an implementation taking `((node: Node) => boolean) | undefined`
// is trusted as one over a plain callback is (#1w5bve2, debug.ts's assertNode, assertNotNode,
// assertOptionalNode and assertEachNode, and visitorPublic.ts's visitArrayWorker): an absent argument is
// never called, a present one is exactly the function, and every argument through the overload keeps
// predicateArguments' proof. Called with a one-line kind test, and through the plain overload with none,
// both agree with Node. tsc's assertNode, assertEachNode and assertOptionalNode take a test on T, not
// Node, and with the claim trusted they meet the census overload proof next, which refuses a callback on
// T served by one the implementation calls with any Node; that refusal is pinned here so its move is
// seen. Mutant: the implementation's type read as written, without looking through the parentheses and
// undefined, which refuses both agreeing shapes as "there is no body proving this parameter".
func TestOverloadPredicateOverOptionalCallback(test *testing.T) {
	test.Parallel()
	test.Run("assertNode", func(test *testing.T) {
		test.Parallel()
		lowersAndAgreesWithNodeAs(test, fmt.Sprintf(assertNodeSource, "Node"), ".ts")
	})
	test.Run("assertNode on T", func(test *testing.T) {
		test.Parallel()
		_, err := lowerSourceAs(test, fmt.Sprintf(assertNodeSource, "T"), ".ts")
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(
			err.Error(),
			"overload 1 of assertNode parameter test cannot be served by implementation parameter test",
		) {
			test.Fatalf("want the census overload proof's refusal, got %v", err)
		}
	})
	test.Run("visitArrayWorker", func(test *testing.T) {
		test.Parallel()
		lowersAndAgreesWithNodeAs(test, fmt.Sprintf(
			visitArraySource,
			"((node: Node) => boolean) | undefined",
			"test === undefined || test(node)",
			`console.log(String(visitArrayWorker(nodes, isIdentifier, 0, 4).length));
console.log(String(visitArrayWorker(nodes, isIdentifier, 2, 2).length));`,
		), ".ts")
	})
}

// The optional callback changes nothing else: each argument through the overload is admitted or refused
// exactly as for its plain twin, whose implementation takes `(node: Node) => boolean` as core.ts's filter
// does, in .ts, where an arrow claiming more than its body proves is a checked closure, and in .a, where
// it's refused. A let that's reassigned can hold any function of its type, so it's refused in both.
// Mutant: the fix reverted, where the proven rows and the lying arrow in .ts are refused at the claim,
// "there is no body proving this parameter", and stop matching their twins.
func TestOverloadPredicateOverOptionalCallbackMatchesItsPlainTwin(test *testing.T) {
	test.Parallel()
	outcome := func(err error) string {
		var refused *Refused
		if errors.As(err, &refused) {
			return "refused: " + refused.What
		}
		if err != nil {
			return err.Error()
		}
		return "admitted"
	}
	for _, probe := range []struct{ name, calls, typeScript, adamic string }{
		{"proven", `console.log(String(visitArrayWorker(nodes, isIdentifier, 0, 4).length));`, "admitted", "admitted"},
		{
			"lying arrow",
			`console.log(String(visitArrayWorker(nodes, (node: Node): node is Identifier => true, 0, 4).length));`,
			"admitted",
			`refused: an unproven predicate argument for parameter test (argument "(node: Node): node is Identifier => true")`,
		},
		{"reassigned let", `let test = (node: Node): node is Identifier => node.kind === 80;
test = (node: Node): node is Identifier => true;
console.log(String(visitArrayWorker(nodes, test, 0, 4).length));`, `refused: an unproven predicate argument for parameter test (argument "test")`, `refused: an unproven predicate argument for parameter test (argument "test")`},
	} {
		for _, mode := range []struct{ extension, want string }{{".ts", probe.typeScript}, {".a", probe.adamic}} {
			test.Run(probe.name+mode.extension, func(test *testing.T) {
				test.Parallel()
				_, optionalErr := lowerSourceAs(test, fmt.Sprintf(
					visitArraySource,
					"((node: Node) => boolean) | undefined",
					"test === undefined || test(node)",
					probe.calls,
				), mode.extension)
				_, plainErr := lowerSourceAs(
					test,
					fmt.Sprintf(visitArraySource, "(node: Node) => boolean", "test(node)", probe.calls),
					mode.extension,
				)
				optional, plain := outcome(optionalErr), outcome(plainErr)
				if optional != plain {
					test.Fatalf("optional %q, plain twin %q; want them to match", optional, plain)
				}
				if optional != mode.want {
					test.Fatalf("got %q for both; want %q", optional, mode.want)
				}
			})
		}
	}
}

// Only undefined is looked through: an implementation taking the callback or null, with undefined or
// without, isn't a plain callback, and the overload's claim stays refused as before (#1w5bve2, Lower's
// should on block 1). An argument of null reaches the body, which must test for it, so a widening that
// takes null out too goes red here instead of opening the parameter silently. Mutant: plainCallbackType
// looking through null as through undefined, which admits every row.
func TestOverloadPredicateOverNullableCallbackStaysRefused(test *testing.T) {
	test.Parallel()
	for _, implementation := range []struct{ name, test, guard string }{
		{"null", "((node: Node) => boolean) | null", "test !== null && test(node)"},
		{
			"null or undefined",
			"((node: Node) => boolean) | null | undefined",
			"test !== null && test !== undefined && test(node)",
		},
	} {
		for _, extension := range []string{".ts", ".a"} {
			test.Run(implementation.name+extension, func(test *testing.T) {
				test.Parallel()
				_, err := lowerSourceAs(test, fmt.Sprintf(
					visitArraySource,
					implementation.test,
					implementation.guard,
					`console.log(String(visitArrayWorker(nodes, isIdentifier, 0, 4).length));`,
				), extension)
				var refused *Refused
				if !errors.As(err, &refused) ||
					refused.What != "a type predicate whose return is not proven (there is no body proving this parameter)" {
					test.Fatalf("want the overload's claim refused, got %v", err)
				}
			})
		}
	}
}
