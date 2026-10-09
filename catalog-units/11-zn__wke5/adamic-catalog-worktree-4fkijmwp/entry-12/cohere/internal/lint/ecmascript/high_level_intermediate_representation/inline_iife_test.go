package high_level_intermediate_representation

// Tests for the CFG splice.
//
// Every assertion here is written against a graph PROPERTY rather than against a printed shape,
// because a printed shape passes for the wrong reason: a splice that routes a return to the wrong
// block still prints instructions in a plausible order. The properties asserted are the four the
// splice can violate independently -- the call is gone, no return of the copied body remains, every
// exit of the copied body reaches the continuation, and the call's result is defined on every path
// out of it.

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// inlinedFixture lowers source, splices, and returns the function plus how many calls were spliced.
func inlinedFixture(t *testing.T, source string, drop bool) (*Function, int) {
	t.Helper()
	var result *Function
	count := 0
	probe := rule.Rule{
		Name:             "inline-iife-fixture",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					if ctx.TypeChecker == nil {
						return
					}
					ForEachFunctionLike(node, func(functionNode *ast.Node) {
						if result != nil {
							return
						}
						lowered := Lower(functionNode, ctx.TypeChecker)
						if lowered == nil {
							return
						}
						if drop {
							DropManualMemoization(lowered)
						}
						Construct(lowered)
						count = InlineImmediatelyInvokedFunctionExpressions(lowered)
						result = lowered
					})
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"/react.d.ts":  reactiveDeclarations,
		"/fixture.tsx": source,
	}, "/fixture.tsx")
	return result, count
}

const singleReturnIifeSource = `
	function Component(properties: {items: Array<number>}) {
		const built = (() => {
			const out = [];
			out.push(properties.items);
			return out;
		})();
		return [built];
	}
`

// The branch is what produces two returns, which is what forces the labeled form. A single-return
// body exercises the other arm entirely and would leave the label path untested.
const multipleReturnIifeSource = `
	function Component(properties: {items: Array<number>; flag: boolean}) {
		const built = (() => {
			const out = [];
			if (properties.flag) {
				return out;
			}
			out.push(properties.items);
			return out;
		})();
		return [built];
	}
`

// TestInlineRemovesTheCallAndEveryReturnOfTheInlinedBody is the core guard.
//
// Three failures are independent and each produces a graph that builds:
//
//   - The `CallExpression` survives, so the closure is still called and the capture edge that
//     widens mutable ranges is still there. The splice then costs work and changes nothing, which
//     is the failure mode with no symptom at all.
//   - A `Return` in the copied body is left unrouted, so control leaves the CALLER at that point.
//     Everything after the call is silently unreachable.
//   - A rewritten return jumps somewhere other than the continuation, so the code after the call
//     never runs on that path.
func TestInlineRemovesTheCallAndEveryReturnOfTheInlinedBody(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
	}{
		{"single return", singleReturnIifeSource},
		{"multiple returns", multipleReturnIifeSource},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function, inlined := inlinedFixture(t, testCase.source, false)
			if function == nil {
				t.Fatal("the fixture did not lower")
			}
			if inlined != 1 {
				t.Fatalf("spliced %d calls, want 1; with none spliced every assertion below "+
					"passes vacuously", inlined)
			}

			// The call and the function expression that fed it are both gone. Either surviving
			// means the closure is still constructed and still called.
			for _, block := range function.Blocks {
				if block == nil {
					continue
				}
				for _, instructionId := range block.Instructions {
					instruction := function.Instructions[instructionId]
					if instruction == nil {
						continue
					}
					switch instruction.Value.(type) {
					case *CallExpression:
						t.Errorf("a CallExpression survives in block %d; the IIFE was not "+
							"actually removed and the capture edge is still there", block.Id)
					case *FunctionExpression:
						t.Errorf("a FunctionExpression survives in block %d; its value is dead "+
							"but range inference still sees a closure over the captures",
							block.Id)
					}
				}
			}

			// Exactly one `Return` remains and it is the CALLER's own. More than one means a
			// return of the inlined body was left unrouted and now ends the caller.
			returns := 0
			for _, block := range function.Blocks {
				if block == nil {
					continue
				}
				if _, isReturn := block.Terminal.(*Return); isReturn {
					returns++
				}
			}
			if returns != 1 {
				t.Errorf("the function has %d Return terminals, want 1; a return of the inlined "+
					"body was not routed to the continuation and now leaves the caller", returns)
			}

			// The code AFTER the call has to still be there, and this needed a mutation sweep to
			// find. Dropping the instructions the continuation inherits left every other assertion
			// in this file passing: the continuation is then empty, `ReversePostorder` collects no
			// instructions from it, and the graph that survives is internally consistent with the
			// caller's tail simply deleted. Consistency is not conservation.
			//
			// Named instructions are the check rather than a count, because the splice legitimately
			// adds and removes instructions and a count would have to be re-derived per fixture.
			// `built` is the binding the caller writes after the call, so it is the tail.
			tail := false
			for _, block := range function.Blocks {
				if block == nil {
					continue
				}
				for _, instructionId := range block.Instructions {
					instruction := function.Instructions[instructionId]
					if instruction == nil {
						continue
					}
					store, isStore := instruction.Value.(*StoreLocal)
					if !isStore {
						continue
					}
					if function.Identifiers[store.LValue.Identifier].Name == "built" {
						tail = true
					}
				}
			}
			if !tail {
				t.Error("the `const built` the caller writes after the call is gone from the " +
					"spliced function; the continuation did not inherit the instructions that " +
					"followed the call, so the caller's tail was dropped rather than moved")
			}
		})
	}
}

// TestInlineDefinesTheCallResultOnEveryPathToTheContinuation is the join half.
//
// The splice's whole purpose is that the call's lvalue still holds the value the body returned. A
// return routed to the continuation WITHOUT the store that carries its value produces a graph with
// no dangling reference and no unreachable block -- the value is simply never defined, and every
// later read of it reads whatever the identifier held before.
//
// Asserted as reachability rather than by counting stores, because the labeled form emits one store
// per return and the direct form emits one load, and a test that counted would encode the arm
// rather than the property.
//
// # The definition set includes lvalue ROLES, which took a measurement to get right
//
// Written first as "every `Instruction.LValue`, every phi, every parameter", this reported the
// caller's own `const built` as undefined -- BEFORE any splice had happened. `StoreLocal` and
// `DeclareLocal` name their binding inside the VALUE rather than in `Instruction.LValue`, which
// carries only the store's own temporary result, so a definition set built from lvalues alone
// misses every named binding in the function. The visitor already reports the role, so the fix is
// to read it rather than to loosen the assertion.
func TestInlineDefinesTheCallResultOnEveryPathToTheContinuation(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
	}{
		{"single return", singleReturnIifeSource},
		{"multiple returns", multipleReturnIifeSource},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function, inlined := inlinedFixture(t, testCase.source, false)
			if function == nil || inlined != 1 {
				t.Fatalf("the fixture spliced %d calls, want 1", inlined)
			}

			defined := map[static_single_assignment.IdentifierId]bool{}
			for _, block := range function.Blocks {
				if block == nil {
					continue
				}
				for _, instructionId := range block.Instructions {
					instruction := function.Instructions[instructionId]
					if instruction == nil {
						continue
					}
					defined[instruction.LValue.Identifier] = true
					EachPlace(instruction.Value, func(place Place, role PlaceRole) {
						if role == PlaceRoleDefine {
							defined[place.Identifier] = true
						}
					})
				}
				for _, phi := range block.Phis {
					defined[phi.Place.Identifier] = true
				}
			}
			for _, parameter := range function.Params {
				defined[parameter.Identifier] = true
			}
			for _, contextValue := range function.Context {
				defined[contextValue.Identifier] = true
			}

			reads := 0
			for _, block := range function.Blocks {
				if block == nil {
					continue
				}
				for _, instructionId := range block.Instructions {
					instruction := function.Instructions[instructionId]
					if instruction == nil {
						continue
					}
					EachPlace(instruction.Value, func(place Place, role PlaceRole) {
						if role == PlaceRoleDefine {
							return
						}
						reads++
						if !defined[place.Identifier] {
							t.Errorf("instruction %d reads identifier %d (%s), which nothing in "+
								"the spliced function defines; a return was routed to the "+
								"continuation without the store that carries its value",
								instructionId, place.Identifier,
								function.Identifiers[place.Identifier].Name)
						}
					})
				}
			}
			if reads == 0 {
				t.Error("no instruction reads any place, so this test asserted nothing")
			}

			// A DECLARATION is not a definition, and this is the part a mutation sweep had to
			// teach. Removing the store that carries a returned value left every assertion above
			// passing, because the labeled form emits `DeclareLocal let result` before the body and
			// the loop above counts that as defining `result`. What remains is a `let` that is
			// declared, read after the call, and never assigned on any path -- the exact failure
			// this test names in its own error message, invisible to it.
			//
			// So the result is asserted separately and on assignment alone: the value the caller
			// reads after the splice must be written by a store or a load somewhere, not merely
			// introduced.
			assigned := map[static_single_assignment.IdentifierId]bool{}
			declared := map[static_single_assignment.IdentifierId]bool{}
			for _, block := range function.Blocks {
				if block == nil {
					continue
				}
				for _, instructionId := range block.Instructions {
					instruction := function.Instructions[instructionId]
					if instruction == nil {
						continue
					}
					switch shape := instruction.Value.(type) {
					case *DeclareLocal:
						declared[shape.LValue.Identifier] = true
					case *StoreLocal:
						assigned[shape.LValue.Identifier] = true
					default:
						assigned[instruction.LValue.Identifier] = true
					}
				}
			}
			for identifier := range declared {
				if !assigned[identifier] {
					t.Errorf("identifier %d is declared by the splice and never assigned; the "+
						"store that carries a returned value into it was not emitted, so the "+
						"caller reads an uninitialised binding", identifier)
				}
			}
		})
	}
}

// TestInlineLeavesEveryBlockReferenceResolvable pins the graph's integrity after the surgery.
//
// The splice writes three terminals -- the cut block's, and one per rewritten return -- and each
// names a block by id. A wrong id here is not a crash: `Block` returns false and most passes skip,
// so the function analyses as if the edge were not there.
func TestInlineLeavesEveryBlockReferenceResolvable(t *testing.T) {
	t.Parallel()

	function, inlined := inlinedFixture(t, multipleReturnIifeSource, false)
	if function == nil || inlined != 1 {
		t.Fatalf("the fixture spliced %d calls, want 1", inlined)
	}
	references := 0
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		EachBlockReferencePointer(block.Terminal, func(reference *static_single_assignment.BlockId) {
			references++
			if _, found := function.Block(*reference); !found {
				t.Errorf("block %d names block %d, which the function does not hold",
					block.Id, *reference)
			}
		})
		if _, found := function.Block(block.Id); !found {
			t.Errorf("block %d is in Blocks but not in the id index", block.Id)
		}
	}
	if references == 0 {
		t.Error("no terminal names a block, so this test asserted nothing")
	}
}

// TestInlineDeclinesTheCasesItCannotExpress pins each refusal.
//
// Each of these is a condition upstream also declines, and each would produce a WRONG graph rather
// than a missing optimisation if it were accepted: an argument has no parameter to bind to, an
// escaping function value would be called somewhere the body no longer exists, and a named binding
// can be read again after the call.
func TestInlineDeclinesTheCasesItCannotExpress(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
	}{
		{
			// An argument would have to bind to a parameter, and this pass lowers no binding.
			name: "call with arguments",
			source: `
				function Component(properties: {value: number}) {
					const built = ((amount: number) => {
						return [amount];
					})(properties.value);
					return [built];
				}
			`,
		},
		{
			// The value escapes into another call, which may invoke it after the splice.
			name: "function value also passed elsewhere",
			source: `
				function Component(properties: {value: number}) {
					const make = () => {
						return [properties.value];
					};
					globalThis.hold = make;
					const built = make();
					return [built];
				}
			`,
		},
		{
			// `await` inside would make the caller's continuation a suspension point.
			name: "async body",
			source: `
				function Component(properties: {value: number}) {
					const built = (async () => {
						return [properties.value];
					})();
					return [built];
				}
			`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function, inlined := inlinedFixture(t, testCase.source, false)
			if function == nil {
				t.Skip("the fixture did not lower")
			}
			if inlined != 0 {
				t.Errorf("spliced %d calls, want 0; this shape has no correct splice and "+
					"accepting it produces a graph that builds and is wrong", inlined)
			}
		})
	}
}

// TestInlineDeclinesADroppedMemoCallback pins the guarded entry point's narrower selection.
//
// `DropManualMemoization` rewrites `useMemo(callback, deps)` into a zero-argument call of the
// callback, which is structurally an IIFE. Production uses the memo-inclusive wrapper and inlines
// it like upstream. This focused test deliberately uses the guarded wrapper so it can distinguish a
// programmer-written IIFE from a rewritten memo; see `memoizedResults`.
//
// The fixture holds BOTH a real IIFE and a `useMemo`, which is what makes the assertion sharp: a
// guard that declined everything would also pass a fixture with only the memo in it.
func TestInlineDeclinesADroppedMemoCallback(t *testing.T) {
	t.Parallel()

	const source = `
		function Component(properties: {items: Array<number>}) {
			const built = (() => {
				const out = [];
				out.push(properties.items);
				return out;
			})();
			const memo = useMemo(() => [properties.items], [properties.items]);
			return [built, memo];
		}
	`
	function, inlined := inlinedFixture(t, source, true)
	if function == nil {
		t.Fatal("the fixture did not lower")
	}
	if inlined != 1 {
		t.Fatalf("spliced %d calls, want exactly 1; the real IIFE must be spliced and the memo "+
			"callback must not, and a count of 0 or 2 means the guard is on the wrong one",
			inlined)
	}

	// The memo's call survives, and it is the one bracketed by the markers.
	calls, finishes := 0, 0
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			switch instruction.Value.(type) {
			case *CallExpression:
				calls++
			case *FinishMemoize:
				finishes++
			}
		}
	}
	if calls != 1 {
		t.Errorf("%d CallExpressions survive, want 1; the memo callback's call is the only one "+
			"that should remain", calls)
	}
	if finishes != 1 {
		t.Errorf("%d FinishMemoize markers survive, want 1; the markers the rule reads were "+
			"disturbed by the splice", finishes)
	}
}

// Optional dependency lowering can put the callback definition before an optional CFG and the
// rewritten zero-argument memo call after its join. Upstream's candidate table is function-wide,
// so the intervening block boundary does not stop the callback from being recognised as an IIFE.
func TestInlineFindsAMemoCallbackAcrossBlocks(t *testing.T) {
	t.Parallel()

	const source = `
		function Component(arg: {items?: Array<number>}) {
			const value = useMemo(() => {
				const x = [];
				x.push(arg?.items);
				return x;
			}, [arg?.items]);
			return value;
		}
	`
	function, guarded := inlinedFixture(t, source, true)
	if function == nil {
		t.Fatal("the fixture did not lower")
	}
	if guarded != 0 {
		t.Fatalf("the guarded pass spliced %d calls, want 0 for a dropped memo callback", guarded)
	}

	definitions := map[static_single_assignment.IdentifierId]static_single_assignment.BlockId{}
	calls := map[static_single_assignment.IdentifierId]static_single_assignment.BlockId{}
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			switch value := instruction.Value.(type) {
			case *FunctionExpression:
				definitions[instruction.LValue.Identifier] = block.Id
			case *CallExpression:
				if len(value.Args) == 0 {
					calls[value.Callee.Identifier] = block.Id
				}
			}
		}
	}
	separated := false
	for callee, callBlock := range calls {
		if definitionBlock, found := definitions[callee]; found && definitionBlock != callBlock {
			separated = true
		}
	}
	if !separated {
		t.Fatal("the callback definition and rewritten call did not land in different blocks; " +
			"the fixture no longer exercises function-wide candidate lifetime")
	}

	if inlined := InlineImmediatelyInvokedFunctionExpressionsIncludingMemoCallbacks(function); inlined != 1 {
		t.Fatalf("spliced %d callbacks, want 1 across the optional dependency CFG", inlined)
	}
}

// TestInlineIsIdempotent pins that a second run finds nothing.
//
// A pass that re-splices its own output would copy the body again on every call, and the pipeline
// calls this once per function per rule invocation.
func TestInlineIsIdempotent(t *testing.T) {
	t.Parallel()

	function, inlined := inlinedFixture(t, multipleReturnIifeSource, false)
	if function == nil || inlined != 1 {
		t.Fatalf("the fixture spliced %d calls, want 1", inlined)
	}
	blocks := len(function.Blocks)
	if again := InlineImmediatelyInvokedFunctionExpressions(function); again != 0 {
		t.Errorf("a second run spliced %d more calls, want 0", again)
	}
	if len(function.Blocks) != blocks {
		t.Errorf("a second run changed the block count from %d to %d", blocks, len(function.Blocks))
	}
}

// TestMemoInclusiveInliningCallSitesAreReviewed pins the production precondition on the
// memo-inclusive entry point.
//
// # Why this is a grep and not a behavioural assertion
//
// The splice is upstream's and is now the production path, including where reactive scopes are
// computed. It is still ordering-sensitive: a caller must run it after memo erasure, merge the
// consecutive blocks it creates, and rebuild graph-derived facts before range/scope inference.
//
// The scope and dependency oracles catch behavioural drift in their current callers. What they
// cannot catch is an unrelated caller introducing a second, incomplete pipeline, so this remains a
// census: each new call site must be reviewed for the ordering above and, if it computes scopes,
// measured with scope under-production held at zero.
//
// The known scope-producing callers below are justified by the final production-equivalent
// measurement: 26/48 scope fixtures exact, 0 fixtures / 0 scopes under, preserve-manual-memoization
// 28/28 goldens and 0/69 false positives.
var memoInclusiveCall = regexp.MustCompile(`inlineInvokedFunctions\([^,)]*,\s*true\s*\)`)

func TestMemoInclusiveInliningCallSitesAreReviewed(t *testing.T) {
	t.Parallel()

	// The call sites that are known and reviewed. `cache.go` computes no reactive scope. Production
	// and the structural probes, including state-setter pruning, use the complete prefix above.
	knownCallSites := map[string]bool{
		"../../ecmascript/high_level_intermediate_representation/cache.go": true,
		"cache.go":                          true,
		"dependency_oracle_test.go":         true,
		"effects_react_state_test.go":       true,
		"flatten_scopes_with_hooks_test.go": true,
		"hoistable_test.go":                 true,
		"preserve_manual_memoization.go":    true,
		"scope_oracle_test.go":              true,
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading the package directory: %v", err)
	}

	found := []string{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		// The definition itself and this test both name it; neither is a call site.
		if name == "inline_iife.go" || name == "inline_iife_test.go" {
			continue
		}
		body, readErr := os.ReadFile(name)
		if readErr != nil {
			t.Fatalf("reading %s: %v", name, readErr)
		}
		text := string(body)
		// The exported entry point by name, and the unexported one asked for the memo-inclusive
		// behaviour under any argument spelling. The second pattern is a regexp rather than a
		// literal because the first version of this test matched only
		// `inlineInvokedFunctions(function, true)` and a caller writing `(f, true)` walked past it
		// -- found by adding exactly that caller and watching the test stay green.
		if strings.Contains(text, "InlineImmediatelyInvokedFunctionExpressionsIncludingMemoCallbacks") ||
			memoInclusiveCall.MatchString(text) {
			if !knownCallSites[name] {
				found = append(found, name)
			}
		}
	}

	if len(found) != 0 {
		sort.Strings(found)
		t.Fatalf("new unreviewed caller(s) of memo-inclusive inlining: %v.\n"+
			"Read the production ordering at "+
			"InlineImmediatelyInvokedFunctionExpressionsIncludingMemoCallbacks. Confirm the caller "+
			"runs memo erasure first and merges changed blocks afterwards; if it computes scopes, "+
			"re-run the scope oracle and keep under-production at 0 fixtures / 0 scopes before "+
			"adding it to knownCallSites.", found)
	}
}

// A multi-exit body's stores must name the value each `return` produces, not `Function.Returns`.
//
// A `Return` terminal names `Function.Returns`, the one identifier every `return` in a function
// stores into. Nothing ever defines that identifier -- the stores target it -- so
// `CopyNestedBodyInto` has no remap entry for it and a copied terminal still names the nested
// function's version. Reading the terminal directly then produces a reassignment whose rvalue no
// instruction in the caller defines.
//
// The failure is quiet, which is why it is pinned here rather than left to the corpus. The store
// itself survives dead-code elimination, because a reassignment is unconditionally live, so the
// graph stays well formed and every structural assertion keeps passing. What disappears is whatever
// the return value was built from: nothing references it, and the sweep takes it.
//
// Measured on `error.useMemo-infer-less-specific-conditional-access` with the memo callback inlined:
// the store read identifier 30, defined nowhere, while the `ObjectExpression` the callback returns
// sat directly above it and was pruned along with the two `PropertyLoad`s feeding it. Upstream keeps
// all three at its own `DeadCodeElimination` stage.
func TestInlineStoresTheValueEachReturnProduces(t *testing.T) {
	t.Parallel()

	const source = `
	declare function mutate(v: unknown): void;
	function Component(properties: {flag: boolean; a: {b: number}}) {
		const built = (() => {
			const x = {};
			if (properties.flag) {
				mutate(x);
				return {value: properties.a.b};
			}
		})();
		return [built];
	}
`
	function, inlined := inlinedFixture(t, source, false)
	if function == nil || inlined != 1 {
		t.Fatalf("the fixture spliced %d calls, want 1", inlined)
	}

	// Every identifier an instruction defines. A store whose rvalue is outside this set names a
	// value the caller never produces.
	defined := map[static_single_assignment.IdentifierId]bool{}
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			if instruction := function.Instructions[instructionId]; instruction != nil {
				defined[instruction.LValue.Identifier] = true
			}
		}
	}

	stores, resolved := 0, 0
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			store, isStore := instruction.Value.(*StoreLocal)
			if !isStore || store.Kind != InstructionKindReassign {
				continue
			}
			stores++
			if defined[store.Value.Identifier] {
				resolved++
			}
		}
	}
	if stores != 2 {
		t.Errorf("found %d return stores, want 2 -- the fixture has two exits", stores)
	}
	// One arm returns an expression and one falls off the end. The first must name the value it
	// built; the second returns undefined and keeps the terminal's own place, which is
	// `Function.Returns` and is defined by nothing. Emitting an explicit `undefined` there was
	// measured and made the golden trade worse, so the asymmetry is deliberate -- see
	// `returnedPlace`.
	if resolved != 1 {
		t.Errorf("%d of %d return stores name a defined value, want 1", resolved, stores)
	}
}
