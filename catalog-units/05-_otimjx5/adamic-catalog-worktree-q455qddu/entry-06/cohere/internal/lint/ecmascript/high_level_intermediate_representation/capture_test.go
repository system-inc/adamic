// Tests for closure capture: the values a nested function closes over from an enclosing one.
//
// # What these must prove, beyond "captures is non-empty"
//
// A capture test that only asserts a non-empty `Captures` slice would pass on a lowering that
// captured every free name including true globals, which is a worse answer than capturing nothing:
// it would make `react/globals` silent on real global writes. So the negative cases carry equal
// weight here, and there are more of them than positive ones. A global, an import, a module-scope
// binding, and a parameter of the nested function itself must all stay OUT of the capture set.
package high_level_intermediate_representation

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// lowerTypedFunctions lowers every outermost function in code with a real checker.
func lowerTypedFunctions(t *testing.T, name string, code string) []*Function {
	t.Helper()
	var lowered []*Function
	probe := rule.Rule{
		Name:             "capture-test-harness",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					if ctx.TypeChecker == nil {
						t.Fatal("the harness produced no type checker, so this test would prove nothing")
					}
					forEachFunctionLike(node, func(function *ast.Node) {
						if fn := Lower(function, ctx.TypeChecker); fn != nil {
							lowered = append(lowered, fn)
						}
					})
				},
			}
		},
	}
	rule_testing.RunTyped(t, probe, name, code)
	if len(lowered) == 0 {
		t.Fatal("nothing lowered")
	}
	return lowered
}

// captureNames returns the names of the values a function expression captures, as the ENCLOSING
// function names them.
func captureNames(function *Function, expression *FunctionExpression) []string {
	names := make([]string, 0, len(expression.Captures))
	for _, capture := range expression.Captures {
		names = append(names, function.Identifiers[capture.Identifier].Name)
	}
	return names
}

// firstFunctionExpression returns the first nested-function instruction in a function.
func firstFunctionExpression(function *Function) *FunctionExpression {
	for _, instruction := range function.Instructions {
		if expression, ok := instruction.Value.(*FunctionExpression); ok {
			return expression
		}
	}
	return nil
}

// instructionNames returns the printed instruction values of a function, for asserting that a
// particular variant was or was not produced.
func instructionNames(function *Function) string {
	return Print(function)
}

// TestCaptureSymbolIdentityCrossesFunctions pins the property the whole capture mechanism rests on.
//
// `captureOf` finds captures by asking whether an ENCLOSING builder knows the symbol, which is only
// correct if the checker returns the same `*ast.Symbol` for a declaration and for a reference to it
// from inside a nested function. That is measured here rather than assumed, because if it ever
// stopped holding, every capture would silently revert to `LoadGlobal` and every test below would
// still be able to pass for the wrong reason.
func TestCaptureSymbolIdentityCrossesFunctions(t *testing.T) {
	t.Parallel()

	const code = `
export function Outer() {
  const captured = 1;
  const nested = () => captured;
  return nested;
}
`
	probe := rule.Rule{
		Name:             "symbol-identity-probe",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					var occurrences []*ast.Symbol
					var walk func(*ast.Node)
					walk = func(current *ast.Node) {
						if current == nil {
							return
						}
						if current.Kind == ast.KindIdentifier && current.Text() == "captured" {
							occurrences = append(occurrences, ctx.TypeChecker.GetSymbolAtLocation(current))
						}
						current.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
					}
					walk(node)

					if len(occurrences) != 2 {
						t.Fatalf("expected the declaration and one nested reference, found %d occurrences",
							len(occurrences))
					}
					if occurrences[0] == nil {
						t.Fatal("the checker resolved the declaration to no symbol")
					}
					if occurrences[0] != occurrences[1] {
						t.Error("the checker gave the declaration and the nested reference different symbols, " +
							"so capture detection cannot rely on symbol identity")
					}
				},
			}
		},
	}
	rule_testing.RunTyped(t, probe, "identity.tsx", code)
}

// TestCaptureAcceptanceCase is the case three separate rules independently reported.
//
// `react/static-components` measured it exactly: a component built by a call and then referenced
// from a nested arrow. Before captures existed the arrow's reference lowered to `LoadGlobal C`,
// indistinguishable from a true global, and the rule lost the shape.
func TestCaptureAcceptanceCase(t *testing.T) {
	t.Parallel()

	const code = `
export function Outer() {
  const C = makeIt();
  const render = () => <C/>;
  return render;
}
`
	outer := lowerTypedFunctions(t, "acceptance.tsx", code)[0]

	expression := firstFunctionExpression(outer)
	if expression == nil {
		t.Fatal("the nested arrow did not lower to a FunctionExpression")
	}
	names := captureNames(outer, expression)
	if len(names) != 1 || names[0] != "C" {
		t.Fatalf("expected the arrow to capture C, got %v", names)
	}

	nested := outer.Functions[expression.Function]
	if len(nested.Context) != 1 {
		t.Fatalf("expected one context value on the nested function, got %d", len(nested.Context))
	}

	body := instructionNames(nested)
	if !strings.Contains(body, "LoadContext") {
		t.Errorf("the nested body should read C through LoadContext, got:\n%s", body)
	}
	if strings.Contains(body, "LoadGlobal C") {
		t.Errorf("the nested body still reads C as a global, which is the defect this fixes:\n%s", body)
	}
}

// TestCaptureDoesNotClaimGlobals is the control that matters most to `react/globals`.
//
// The failure mode this guards is the opposite of the one above: a lowering that treated every
// unresolved name as a capture would make a real global write invisible. Each name here is free in
// the nested function and must NOT become a capture.
func TestCaptureDoesNotClaimGlobals(t *testing.T) {
	t.Parallel()

	const code = `
import defaultExport from "./module";
import { named } from "./module";
const moduleLocal = 1;

export function Outer() {
  const nested = () => {
    console.log(defaultExport, named, moduleLocal);
    return globalThis;
  };
  return nested;
}
`
	functions := lowerTypedFunctions(t, "globals.tsx", code)
	var outer *Function
	for _, function := range functions {
		if function.Name == "Outer" {
			outer = function
		}
	}
	if outer == nil {
		t.Fatal("Outer did not lower")
	}

	expression := firstFunctionExpression(outer)
	if expression == nil {
		t.Fatal("the nested arrow did not lower to a FunctionExpression")
	}
	if names := captureNames(outer, expression); len(names) != 0 {
		t.Errorf("globals, imports, and module-scope bindings must not be captures, got %v", names)
	}

	nested := outer.Functions[expression.Function]
	if len(nested.Context) != 0 {
		t.Errorf("expected no context values, got %d", len(nested.Context))
	}
	if body := instructionNames(nested); !strings.Contains(body, "LoadGlobal") {
		t.Errorf("the free names should still be globals, got:\n%s", body)
	}
}

// TestCaptureWriteIsNotAGlobalWrite is the over-report this fix removes.
//
// Before captures, `n = 1` inside a nested function emitted `StoreGlobal`, so a rule reading that
// instruction as evidence of a global write was wrong on every closure that assigns to an enclosing
// variable. It must now be a StoreContext and no StoreGlobal may remain.
func TestCaptureWriteIsNotAGlobalWrite(t *testing.T) {
	t.Parallel()

	const code = `
export function Outer() {
  let n = 0;
  const bump = () => { n = 1; };
  return [n, bump];
}
`
	outer := lowerTypedFunctions(t, "write.tsx", code)[0]
	expression := firstFunctionExpression(outer)
	if expression == nil {
		t.Fatal("the nested arrow did not lower to a FunctionExpression")
	}
	if names := captureNames(outer, expression); len(names) != 1 || names[0] != "n" {
		t.Fatalf("expected the arrow to capture n, got %v", names)
	}

	nested := outer.Functions[expression.Function]
	body := instructionNames(nested)
	if !strings.Contains(body, "StoreContext") {
		t.Errorf("the write to the captured n should be a StoreContext, got:\n%s", body)
	}
	if strings.Contains(body, "StoreGlobal") {
		t.Errorf("the write to a captured variable is still reported as a global write:\n%s", body)
	}
}

// TestCaptureShadowingIsNotCaptured checks that a nested binding of the same NAME is not a capture.
//
// This is the case a name-based implementation gets wrong and a symbol-based one gets right for
// free, which is the argument for resolving through the checker rather than by string.
func TestCaptureShadowingIsNotCaptured(t *testing.T) {
	t.Parallel()

	const code = `
export function Outer() {
  const value = 1;
  const nested = (value: number) => value + 1;
  return nested;
}
`
	outer := lowerTypedFunctions(t, "shadow.tsx", code)[0]
	expression := firstFunctionExpression(outer)
	if expression == nil {
		t.Fatal("the nested arrow did not lower to a FunctionExpression")
	}
	if names := captureNames(outer, expression); len(names) != 0 {
		t.Errorf("a parameter that shadows an outer name is not a capture, got %v", names)
	}
}

// TestCaptureThroughTwoBoundaries checks that the capture chain is complete at every level.
//
// An inner function reading a variable two functions up must produce a capture in the MIDDLE
// function too. Without that, the middle function's `Captures` would name an identifier that does
// not exist in its own table, which is a malformed graph rather than merely an imprecise one.
func TestCaptureThroughTwoBoundaries(t *testing.T) {
	t.Parallel()

	const code = `
export function Outer() {
  const deep = 1;
  const middle = () => {
    const inner = () => deep;
    return inner;
  };
  return middle;
}
`
	outer := lowerTypedFunctions(t, "chain.tsx", code)[0]

	outerExpression := firstFunctionExpression(outer)
	if outerExpression == nil {
		t.Fatal("the middle function did not lower")
	}
	if names := captureNames(outer, outerExpression); len(names) != 1 || names[0] != "deep" {
		t.Fatalf("the middle function should capture deep from Outer, got %v", names)
	}

	middle := outer.Functions[outerExpression.Function]
	middleExpression := firstFunctionExpression(middle)
	if middleExpression == nil {
		t.Fatal("the inner function did not lower")
	}
	if names := captureNames(middle, middleExpression); len(names) != 1 || names[0] != "deep" {
		t.Fatalf("the inner function should capture deep from the middle function, got %v", names)
	}

	// The pairing must be valid IN THE MIDDLE FUNCTION'S table: the capture the middle function
	// hands down has to be a value the middle function actually holds.
	capture := middleExpression.Captures[0]
	if int(capture.Identifier) >= len(middle.Identifiers) {
		t.Fatalf("the inner function's capture names identifier %d, which the middle function's "+
			"table of %d does not contain", capture.Identifier, len(middle.Identifiers))
	}
	found := false
	for _, context := range middle.Context {
		if context.Identifier == capture.Identifier {
			found = true
		}
	}
	if !found {
		t.Error("the value the middle function hands to the inner one is not in its own Context")
	}
}

// TestCaptureIndicesPairAcrossTheBoundary is the property a pass walking into a closure depends on.
//
// `Captures[i]` and `nested.Context[i]` must name the same source binding seen from the two sides.
// If the two orders ever diverged, a pass following a value into a closure would silently follow
// the WRONG value, which no structural check would catch.
func TestCaptureIndicesPairAcrossTheBoundary(t *testing.T) {
	t.Parallel()

	const code = `
export function Outer() {
  const first = 1;
  const second = 2;
  const third = 3;
  const nested = () => third + first + second;
  return nested;
}
`
	outer := lowerTypedFunctions(t, "pairing.tsx", code)[0]
	expression := firstFunctionExpression(outer)
	if expression == nil {
		t.Fatal("the nested arrow did not lower")
	}
	nested := outer.Functions[expression.Function]

	if len(expression.Captures) != len(nested.Context) {
		t.Fatalf("captures (%d) and context (%d) must have the same length",
			len(expression.Captures), len(nested.Context))
	}
	if len(expression.Captures) != 3 {
		t.Fatalf("expected three captures, got %d", len(expression.Captures))
	}

	for index := range expression.Captures {
		outerName := outer.Identifiers[expression.Captures[index].Identifier].Name
		innerName := nested.Identifiers[nested.Context[index].Identifier].Name
		if outerName != innerName {
			t.Errorf("index %d pairs %q on the outside with %q on the inside", index, outerName, innerName)
		}
	}

	// Use order, not declaration order: the nested body reads third first.
	names := captureNames(outer, expression)
	if names[0] != "third" {
		t.Errorf("captures should be in first-use order, got %v", names)
	}
}

// TestCaptureKeepsSSAValid runs construction over functions that capture and verifies the result.
//
// A capture introduces a value with NO defining instruction in the nested function, which is
// exactly the shape that breaks a use-must-be-dominated-by-its-definition checker if the value is
// not treated as live on entry. That it passes is asserted here; that the checker could still fail
// is asserted by TestCaptureVerifierStillDetectsAViolation below.
func TestCaptureKeepsSSAValid(t *testing.T) {
	t.Parallel()

	sources := map[string]string{
		"read": `export function Outer(flag: boolean) {
  let n = 1;
  if (flag) { n = 2; }
  const read = () => n;
  return read;
}`,
		"write in a loop": `export function Outer(items: number[]) {
  let total = 0;
  for (const item of items) {
    const add = () => { total = total + item; };
    add();
  }
  return total;
}`,
		"capture in a try": `export function Outer() {
  let value = 0;
  try {
    const set = () => { value = 1; };
    set();
  } catch (error) {
    const report = () => value + 1;
    report();
  }
  return value;
}`,
	}

	for name, code := range sources {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			outer := lowerTypedFunctions(t, "ssa.tsx", code)[0]
			Construct(outer)
			cohereEverywhere(t, outer)

			// A test that verified an empty graph would prove nothing.
			if len(outer.Functions) == 0 {
				t.Fatal("no nested function lowered, so nothing about captures was checked")
			}
			nested := outer.Functions[0]
			if len(nested.Context) == 0 {
				t.Error("the nested function captured nothing, so this case is not exercising captures")
			}
		})
	}
}

// TestCaptureVerifierStillDetectsAViolation is the negative control for the test above.
//
// `TestCaptureKeepsSSAValid` reports a clean verification, and a clean result is only evidence if
// the checker can fail on the same shape. This breaks a nested function that holds captures and
// requires the verifier to notice.
func TestCaptureVerifierStillDetectsAViolation(t *testing.T) {
	t.Parallel()

	const code = `
export function Outer(flag: boolean) {
  let n = 1;
  if (flag) { n = 2; }
  const read = () => { let local = n; if (flag) { local = 3; } return local; };
  return read;
}
`
	outer := lowerTypedFunctions(t, "control.tsx", code)[0]
	Construct(outer)
	if violations := VerifySSA(outer); len(violations) != 0 {
		t.Fatalf("the baseline must be clean before it is broken, got %v", violations)
	}

	nested := outer.Functions[0]
	if len(nested.Context) == 0 {
		t.Fatal("the nested function captured nothing, so this control is not testing captured code")
	}
	if violations := VerifySSA(nested); len(violations) != 0 {
		t.Fatalf("the nested baseline must be clean before it is broken, got %v", violations)
	}

	// Break it the way a mis-placed phi breaks it: move a phi off its join.
	var moved *Phi
	var from, to *BasicBlock
	for _, block := range nested.Blocks {
		if len(block.Phis) > 0 {
			moved, from = block.Phis[0], block
			break
		}
	}
	if moved == nil {
		t.Fatal("the nested function has no phi, so this control cannot break it")
	}
	for _, block := range nested.Blocks {
		if block.Id != from.Id && len(block.Predecessors) == 1 {
			to = block
			break
		}
	}
	if to == nil {
		t.Fatal("no block to move the phi to")
	}
	from.Phis = nil
	to.Phis = append(to.Phis, moved)

	violations := VerifySSA(nested)
	if len(violations) == 0 {
		t.Error("moving a phi off its join in a function that holds captures was not detected, " +
			"so a clean verification of captured code proves nothing")
	} else {
		t.Logf("detected: %v", violations)
	}
}
