package react

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/static_single_assignment"
)

// Refs flags reading or writing a ref's `current` property during render.
//
//	valid:   function C() { const r = useRef(null); useEffect(() => { r.current; }, []); return <div ref={r} />; }
//	valid:   function C() { const r = useRef(null); return <div onClick={() => r.current} />; }
//	valid:   function C() { const r = useRef(null); if (r.current == null) { r.current = 1; } }
//	invalid: function C() { const r = useRef(null); const v = r.current; return v; }
//	invalid: function C() { const r = useRef(null); r.current = 1; return null; }
//	invalid: function C() { const r = useRef(null); const x = foo(r); return x; }
//	invalid: function C(props) { const v = props.ref.current; return <div>{v}</div>; }
//
// A ref is deliberately outside React's data flow: writing one does not schedule a render, and
// during render the ref may not yet point at the thing it will point at once the tree commits.
// Reading `current` while rendering therefore produces a value that is either stale or not yet
// attached, and because the read creates no subscription, nothing re-renders when it changes. The
// symptom is an interface that silently shows the previous frame's data and never corrects itself,
// which reads as a caching bug rather than as what it is.
//
// # Where this comes from
//
// React's own `ValidateNoRefAccessInRender.ts`, transcribed by oxc at
// `oxc_react_compiler/src/react_compiler_validation/validate_no_ref_access_in_render.rs`, which is
// the version this follows at 1,060 lines. oxc's *linter* rule at `rules/react/refs.rs` is a
// four-line dispatcher into `run_react_compiler_rule(ctx, ErrorCategory::Refs)`, so reading only
// the rule file makes a large port look trivial. The judgment lives entirely in the compiler crate.
//
// # This rule reads no effect facts, which is why it could be written now
//
// It was widely believed that this rule waited on the mutation-effect table that
// `immutability` needs. It does not.
// `grep -cE 'impure|no_alias|callee_effect|known_incompatible|positional_params|return_value_kind'`
// over all 1,060 lines of the validator returns 0, measured with a live control in the same command
// (`RefAccessType` returns 100 on the same file, so the zero is a real absence rather than a bad
// pattern). What the rule actually consults is its own six-element lattice, two nominal type
// predicates, and the phi/load/store propagation this representation already provides.
//
// # The lattice, and the equality that is not identity
//
// Six elements: none, nullable, guard, ref, refValue, structure. The important part is not the
// elements but the equality relation over them, and it is the single subtlety most worth carrying
// forward. Upstream hand-writes `PartialEq` rather than deriving it, and its doc comment says why:
// a join mints a FRESH ref id, so an equality that compared ids would report "changed" on every
// iteration and **the fixpoint would never converge**. `refsTypeEqual` below reproduces that
// deliberately: `ref` ignores its id entirely, and `refValue` compares only its access span.
//
// Equality for convergence is not equality for identity. Two values that are genuinely distinct
// refs compare EQUAL here, and that is correct, because what the fixpoint needs to know is whether
// the lattice element moved, not whether the underlying ref is the same object. Deriving this
// comparison is the natural thing to do in Go and it produces a rule that either loops ten times on
// every function or emits a spurious non-convergence finding. `TestRefsConvergenceEqualityIgnores
// RefIdentity` pins it, and task `#g29j9ab` tracks lifting this traversal onto the shelf when a
// third rule needs it; if that happens, this hazard is the thing to carry with it.
//
// # Why this is a bounded sweep rather than a dataflow solve
//
// Upstream is `for iteration in 0..10` over the whole function, with one `changed` flag on a
// per-identifier map, and it emits a non-convergence diagnostic if still moving at ten. That is
// chaotic iteration keyed by identifier, not a per-block entry/exit lattice, so neither
// `control_flow_graph.Solve` nor a dominator tree is the right instrument even in principle.
//
// It is also worth recording that `control_flow_graph.Solve[V,E]` is mechanically unreachable from this
// representation regardless of fit: `control_flow_graph.Block`'s `index`, `final` and `thrown` fields are
// unexported and set only by `control_flow_graph.Build`, so a caller holding a `high_level_intermediate_representation.Function` cannot
// construct one. `ssa.go` and `postdominator.go` both already record this in their headers. Three
// rules shipped on this representation have now read that machinery and deliberately used none of
// it, so the bar for reaching for it is high and this rule does not clear it.
//
// The bound is reproduced exactly, ten rounds, including the non-convergence finding. On the real
// tree the sweep settled in two rounds on every function it saw; see the test file.
//
// # Two signals identify a ref, and BOTH are required
//
// This is the part a reader is most likely to get wrong, because either signal alone looks
// sufficient and each covers a different part of the corpus.
//
//   - **The checker.** `useRef` returns `RefObject<T>`, and a ref is identified by that type's
//     SYMBOL. `RefObject` is an interface, so its type alias is correctly nil. This is the exact
//     mirror image of `set_state_in_render.go`'s setter predicate, which keys on the ALIAS because
//     `Dispatch` is a type alias whose symbol renders as a non-printable internal marker. A
//     predicate written for one of them silently reports nothing for the other, and an agent
//     writing this rule keyed on the alias first and reported a genuine `useRef` as neither.
//     Probed on our own checker rather than assumed: `const myRef = useRef(null)` gives
//     `alias=<nil> symbol=RefObject`, while `setX` gives `alias=Dispatch symbol=<binary noise>`.
//
//   - **The name.** Upstream's `enableTreatRefLikeIdentifiersAsRefs` DEFAULTS TO TRUE
//     (`environment_config.rs:157`) and matches `/^(?:[a-zA-Z$_][\w$]*)Ref$|^ref$/` against the
//     name of the binding whose `current` is being accessed. Only 1 of the 42 corpus fixtures sets
//     the flag explicitly, so it is silently active across the whole corpus. This is what reaches
//     `props.ref`, a ref arriving as an untyped prop, where there is no `useRef` call for the
//     checker to type.
//
// Neither signal alone covers the corpus and shipping only the first would decline six fixtures
// silently. See the scope note below.
//
// # What the name signal matches, measured against the executable rather than read
//
// The regex reads as though it matches any `somethingRef.current` anywhere. It does not, and the
// difference is the object's BINDING name rather than the text of the member expression. Driven on
// `eslint-plugin-react-hooks` 7.1.1 through the ESLint Linter API on each input:
//
//	const fooRef = {};      fooRef.current        REPORTS   binding named fooRef
//	function C({fooRef})    fooRef.current        REPORTS   binding named fooRef
//	function C(props)       props.fooRef.current  CLEAN     binding is props, not fooRef
//	function C(props)       props.ref.current     REPORTS   the intermediate load is named ref
//	const q = props.fooRef; q.current             CLEAN     binding named q
//	const ref = props.any;  ref.current           REPORTS   binding named ref
//	const o = props.o;      o.ref.current         CLEAN     binding named o
//
// The apparent inconsistency between rows 3 and 4 is not a special case for `props`. Upstream types
// a property load as `Type::Property{object_name, property_name}` where `object_name` is the name
// of the object's own binding, and `is_ref_like_name` is asked at each unification. For
// `props.ref.current` the INNER load produces a value whose binding name is `ref`, and the outer
// `.current` then matches against that. For `props.fooRef.current` the inner load is also named
// `fooRef`, so by the regex it should match too, and upstream is nonetheless silent. Reading
// alone would have produced the wrong rule here in both directions; every row above is a
// measurement, and rows 3 and 5 are reproduced as silence deliberately.
//
// This rule keys the name test on `high_level_intermediate_representation.Identifier.Name`, which is the binding name for a named
// value and empty for a temporary, giving the same partition without re-deriving it.
//
// # `Ref` alone does not match, and that is upstream's text rather than an oversight
//
// The regex requires at least one character before `Ref`, so `fooRef` and `aRef` match while a bare
// `Ref` does not. Upstream states this in a comment at `infer_types.rs:235` and the executable
// agrees (`props.Ref.current` is clean, measured). Reproduced rather than tidied, because a rule
// that helpfully matched `Ref` would disagree with React on real code while looking more correct.
//
// # The component gate is NOT used here, and that is a deliberate difference from its neighbours
//
// `static_components.go`, `globals.go` and `set_state_in_render.go` all gate on the shelf's
// `IsComponentOrHookLike`, and three of them reuse the predicate whole. This rule does not, and the
// reason is that upstream's own gate for this validator is different. Measured on the executable: a
// lowercase `function helper(props) { return props.ref.current; }` is CLEAN, so a gate is real; but
// the gate that decides it is compilation admission, which for this validator already excludes a
// non-component. What this rule needs beyond that is the `!is_function_expression` condition at
// `infer_mutation_aliasing_ranges.rs:470`, which is "this is the outer function under compilation,
// not a nested function expression" and is a different question from "is this a component".
//
// So the compilation gate is reused (a function must look like a component or a hook and must
// actually create JSX or call a hook), and nested functions are analysed as INNER functions through
// the structure element rather than as subjects in their own right. A nested function that touches
// a ref does not report at its own site; it makes its enclosing binding carry a `readRefEffect`, and
// the finding lands where that function is CALLED. That is what makes `capture-ref-for-mutation`
// report at `handleKey('left')()` rather than inside `handleKey`.
//
// # One span where upstream has two, and one message shape
//
// Upstream's `ref_update` diagnostic carries a primary span and a secondary label pointing at the
// ref's origin. Our `Diagnostic` carries a single `Range`, so the primary span is kept where
// upstream puts it and the origin is dropped rather than moved into the message, because unlike
// `static_components` the origin here is usually the same identifier the span already names.
var Refs = rule.Rule{
	Name:             "react-hooks/refs",
	NeedsTypeChecker: true,
	// Reads other files only through shape readers (rule.ExportNameIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			// Taken at the file level for the reason static_components.go gives: lowering already
			// descends into nested functions through the function arena, so listening per function
			// kind would lower every inner function twice, once as a child and once standalone
			// with no enclosing context.
			ast.KindSourceFile: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				// A file that cannot hold a component or a hook is not lowered at all; see
				// high_level_intermediate_representation.MayHoldComponentOrHook for why that is exact.
				if !high_level_intermediate_representation.MayHoldComponentOrHook(ctx) {
					return
				}
				refsForEachCompiledFunction(node, func(functionNode *ast.Node) {
					lowered := high_level_intermediate_representation.ForFunctionWithoutManualMemoization(ctx, functionNode)
					if lowered == nil {
						return
					}
					refsAnalyzeCompilationUnit(ctx, lowered)
				})
			},
		}
	},
}

// refsForEachCompiledFunction visits the outermost function-like node on each branch.
//
// Identical in shape to static_components.go's walker and deliberately not shared: two porters
// collided on unprefixed helper names in this package today and both then removed their copies at
// once, turning a redeclaration error into an undefined-symbol error pointing at nobody. Every
// helper here carries the rule's name. Lifting the shared shape onto the shelf is task `#g29j9ab`.
func refsForEachCompiledFunction(root *ast.Node, visit func(*ast.Node)) {
	if root == nil {
		return
	}
	root.ForEachChild(func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			visit(node)
			return false
		}
		refsForEachCompiledFunction(node, visit)
		return false
	})
}

// refsAnalyzeCompilationUnit runs the sweep over the functions upstream would have compiled.
//
// The gate is upstream's `getReactFunctionType`: a function is admitted only when its name looks
// like a component or a hook AND its body actually creates JSX or calls a hook. A function that is
// not a unit is walked past into the functions it contains, because one of those may be a unit even
// when its wrapper is not.
func refsAnalyzeCompilationUnit(ctx rule.Context, function *high_level_intermediate_representation.Function) {
	if function == nil {
		return
	}
	if !refsIsCompilationUnit(function) {
		for _, nested := range function.Functions {
			refsAnalyzeCompilationUnit(ctx, nested)
		}
		return
	}
	// A unit found nested in a function that is not one is analyzed as React Compiler compiles it,
	// on its own; see AsCompilationUnit.
	function = high_level_intermediate_representation.AsCompilationUnit(ctx, function,
		high_level_intermediate_representation.ForFunctionWithoutManualMemoization)
	findings := refsSweepFunction(ctx, function)
	for _, finding := range findings {
		refsReport(ctx, function, finding)
	}
}

// refsIsCompilationUnit reports whether upstream would compile this function.
//
// Both halves of `getReactFunctionType` are required and the name alone is not enough. The name
// half is taken from `high_level_intermediate_representation.Function.Kind` rather than recomputed, so this rule and the
// representation cannot drift about what a component name is. The body half asks whether the
// function creates JSX or calls a hook.
//
// Unlike static_components.go, the hook half IS asked here and is reachable. That rule's finding
// requires a JSX tag, so every reportable input contains JSX and the hook branch cannot change its
// answer. This rule reports on a bare `useRef` read with no JSX anywhere, so a hook-calling
// component with no JSX is a real subject: `function useHook({value}) { const ref = useRef(null);
// ref.current = value; return ref; }` is the shipped fixture
// `error.invalid-write-but-dont-read-ref-in-render` and it contains no JSX at all.
func refsIsCompilationUnit(function *high_level_intermediate_representation.Function) bool {
	if function == nil || function.Kind == high_level_intermediate_representation.FunctionKindOther {
		return false
	}
	return refsCreatesJsxOrCallsHook(function)
}

// refsCreatesJsxOrCallsHook asks upstream's body test, without descending into nested functions.
//
// The non-descent is the point rather than an optimization: a component whose body is only a
// callback returning JSX is not itself a component to upstream.
func refsCreatesJsxOrCallsHook(function *high_level_intermediate_representation.Function) bool {
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := refsInstructionAt(function, instructionId)
			if instruction == nil {
				continue
			}
			switch value := instruction.Value.(type) {
			case *high_level_intermediate_representation.JsxExpression, *high_level_intermediate_representation.JsxFragment:
				return true
			case *high_level_intermediate_representation.LoadGlobal:
				if refsIsHookName(value.Name) {
					return true
				}

			// `React.useRef(...)` is the same call written through the namespace, and it lowers to
			// a `MethodCall` whose property is a primitive string rather than to a global of its
			// own. Admitting only `LoadGlobal` meant a file that imports React as a default and
			// spells every hook `React.useHook` was never compiled at all, so nothing in it was
			// checked by this rule.
			//
			// Found on a real file rather than by a fixture: `useKingdomLive.tsx` reads a ref
			// during render, ESLint reported it, and this rule was silent because every hook call
			// in that file is spelled through the namespace. `unsupported_syntax.go`'s
			// `isCompilerHookCallee` already accepts both spellings, which is what makes this a gap
			// rather than a decision.
			case *high_level_intermediate_representation.MethodCall:
				if refsIsHookName(refsPrimitiveStringAt(function, value.Property)) {
					return true
				}
			}
		}
	}
	return false
}

// refsPrimitiveStringAt returns the string a place holds, when it holds a primitive string.
//
// A namespaced call lowers its property to a `Primitive` rather than keeping it as syntax, so
// `React.useRef` arrives as a `MethodCall` whose `Property` place is defined by `Primitive{useRef}`.
// Empty for anything else, which reads as "not a hook name" at the one call site.
func refsPrimitiveStringAt(function *high_level_intermediate_representation.Function, place high_level_intermediate_representation.Place) string {
	for _, instruction := range function.Instructions {
		if instruction == nil || instruction.LValue.Identifier != place.Identifier {
			continue
		}
		if primitive, isPrimitive := instruction.Value.(*high_level_intermediate_representation.Primitive); isPrimitive {
			if text, isText := primitive.Value.(string); isText {
				return text
			}
		}
		return ""
	}
	return ""
}

// refsIsHookName reports whether a name is a hook by React's own syntactic test.
//
// `use` followed by an uppercase letter, matching `isHookName` upstream. A bare `use` is a hook
// too under React's rules, but it cannot appear as a `LoadGlobal` callee in any corpus fixture and
// is not admitted here, which is stated rather than left as an accident.
func refsIsHookName(name string) bool {
	if !strings.HasPrefix(name, "use") || len(name) < 4 {
		return false
	}
	initial := name[3]
	return initial >= 'A' && initial <= 'Z'
}

// refsInstructionAt returns one instruction by id, or nil when the id is out of range.
func refsInstructionAt(function *high_level_intermediate_representation.Function, id high_level_intermediate_representation.InstructionId) *high_level_intermediate_representation.Instruction {
	if function == nil || int(id) >= len(function.Instructions) {
		return nil
	}
	return function.Instructions[id]
}

// refsIdentifierNode returns the syntax a value came from, or nil for a pure temporary.
//
// This is the seam `high_level_intermediate_representation.Identifier.Node` documents itself as existing for: the node goes to the
// checker rather than to a local inference pass. Reporting through it rather than through a Place's
// range matters because a Place's range is the range of the node that produced the INSTRUCTION.
func refsIdentifierNode(function *high_level_intermediate_representation.Function, id static_single_assignment.IdentifierId) *ast.Node {
	if function == nil || int(id) >= len(function.Identifiers) {
		return nil
	}
	identifier := function.Identifiers[id]
	if identifier == nil {
		return nil
	}
	return identifier.Node
}

// refsIdentifierName returns a value's source binding name, empty for a temporary.
func refsIdentifierName(function *high_level_intermediate_representation.Function, id static_single_assignment.IdentifierId) string {
	if function == nil || int(id) >= len(function.Identifiers) {
		return ""
	}
	identifier := function.Identifiers[id]
	if identifier == nil {
		return ""
	}
	return identifier.Name
}

// refsIsRefLikeName is upstream's `is_ref_like_name`, minus the property half.
//
// The TypeScript regex is `/^(?:[a-zA-Z$_][\w$]*)Ref$|^ref$/`. Note that `Ref` alone does NOT
// match: the alternation requires at least one character before `Ref`. Upstream states this at
// `infer_types.rs:235` and the executable agrees, measured on `props.Ref.current`, which is clean.
func refsIsRefLikeName(name string) bool {
	if name == "ref" {
		return true
	}
	if len(name) <= 3 || !strings.HasSuffix(name, "Ref") {
		return false
	}
	initial := name[0]
	isAlphabetic := (initial >= 'a' && initial <= 'z') || (initial >= 'A' && initial <= 'Z')
	return isAlphabetic || initial == '$' || initial == '_'
}

// refsIsUseRefType asks the checker whether a value is a `RefObject`.
//
// Keyed on the type's SYMBOL, not its alias. `RefObject` is an interface, so the checker records it
// on the type's symbol and the alias is correctly nil. `set_state_in_render.go`'s setter predicate
// is the exact mirror and reads the alias instead, because `Dispatch` is a type alias whose symbol
// renders as a non-printable internal marker. A predicate written for one silently answers false
// for the other; this one was probed on our own checker before being written, not inferred from the
// shape of its neighbour.
//
// The symbol is compared positively rather than checked for non-nil. The shim is a hand-mirrored
// struct read through `unsafe.Pointer` and has returned a silently wrong type before, so asserting
// that what came back is what was asked for is the only check that can see that class of failure.
func refsIsUseRefType(ctx rule.Context, function *high_level_intermediate_representation.Function, id static_single_assignment.IdentifierId) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	node := refsIdentifierNode(function, id)
	if node == nil || refsCannotHaveRefType(node) {
		return false
	}
	return refsNodeHasRefType(ctx, node)
}

// refsCannotHaveRefType reports whether the language fixes a node's type as something no ref type
// can be, so the checker need not be asked.
//
// The first ask about a node is where the checker computes its type, and for a JSX element that
// means resolving the component and checking every attribute against its props. Profiled on ahra,
// www and connected, 2.1 million asks came from this rule, and only identifiers, calls, property
// accesses and variable declarations ever answered yes. The kinds below answer no by construction,
// not by measurement: a literal, a template, a JSX node, an arithmetic, comparison, bitwise or unary
// result, an object or array literal, and an anonymous function, whose type symbol is the
// anonymous `__function`. `TestRefsCannotHaveRefTypeAgreesWithTheChecker` asks the checker about
// every node this declines in a source written to tempt it.
//
// Deliberately absent: `&&`, `||`, `??`, the comma and the assignments `=`, `&&=`, `||=` and `??=`,
// whose result is an operand and can be a ref; a conditional, an `await`, a `new`, an `as`; and a
// named function, because its type's symbol carries the function's name, so `function RefCallback`
// would answer yes to a test that reads the symbol's name.
func refsCannotHaveRefType(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression,
		ast.KindRegularExpressionLiteral, ast.KindNullKeyword, ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment, ast.KindJsxText,
		ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression, ast.KindArrowFunction,
		ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression, ast.KindTypeOfExpression,
		ast.KindVoidExpression, ast.KindDeleteExpression:
		return true
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Name() == nil
	case ast.KindBinaryExpression:
		return refsOperatorYieldsAPrimitive(node.AsBinaryExpression().OperatorToken.Kind)
	}
	return false
}

// refsOperatorYieldsAPrimitive reports whether a binary operator's result is a number, a bigint, a
// string or a boolean whatever its operands are.
func refsOperatorYieldsAPrimitive(operator ast.Kind) bool {
	switch operator {
	case ast.KindPlusToken, ast.KindMinusToken, ast.KindAsteriskToken, ast.KindSlashToken,
		ast.KindPercentToken, ast.KindAsteriskAsteriskToken,
		ast.KindAmpersandToken, ast.KindBarToken, ast.KindCaretToken,
		ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken,
		ast.KindGreaterThanGreaterThanGreaterThanToken,
		ast.KindLessThanToken, ast.KindGreaterThanToken, ast.KindLessThanEqualsToken,
		ast.KindGreaterThanEqualsToken, ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken,
		ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken,
		ast.KindInstanceOfKeyword, ast.KindInKeyword:
		return true
	}
	return false
}

// refsNodeHasRefType is refsIsUseRefType's checker half, over a node.
func refsNodeHasRefType(ctx rule.Context, node *ast.Node) bool {
	valueType := ctx.TypeChecker.GetTypeAtLocation(node)
	if valueType == nil {
		return false
	}
	// The ALIAS is checked as well as the symbol, and the two carry different names.
	//
	// `RefObject` is an interface, so it lands on the type's symbol and its alias is nil.
	// `RefCallback` is a TYPE ALIAS, so it lands on the alias and its symbol is an anonymous
	// function type. A predicate reading only one field answers false for the other, which is the
	// same asymmetry `set_state_in_render.go` documents from the opposite side, and
	// `internal/utilities/typecheck/specifier.go:39` already does this alias-then-symbol fallback for
	// exactly this reason. Probed on our own checker rather than inferred.
	if alias := shimchecker.Type_alias(valueType); alias != nil {
		if aliasSymbol := alias.Symbol(); aliasSymbol != nil && aliasSymbol.Name == "RefCallback" {
			return true
		}
	}
	symbol := shimchecker.Type_symbol(valueType)
	if symbol == nil {
		return false
	}
	// `RefCallback` is included for the mergeRefs pattern, and it is upstream's exemption rather
	// than a convenience. A function whose result is a ref legitimately RECEIVES refs: that is what
	// merging two refs into one callback means. Upstream expresses this as "the lvalue is a ref"
	// at the call site (`is_ref_lvalue`, the mergeRefs comment at `:697`), and the type our own
	// `mergeReferences` returns is `React.RefCallback<T>`.
	//
	// Measured: `const set = mergeReferences([inner, ref])` is CLEAN upstream and was 8 findings on
	// Kirk's tree, one in every form-field component that forwards a ref. The corpus cannot see this
	// because no fixture merges refs.
	return symbol.Name == "RefObject" || symbol.Name == "MutableRefObject" || symbol.Name == "RefCallback"
}

// ---------------------------------------------------------------------------
// The lattice
// ---------------------------------------------------------------------------

// refsAccessKind is which of the six lattice elements a value holds.
type refsAccessKind uint8

const (
	// refsNone is the bottom element: nothing known, and the join identity.
	refsNone refsAccessKind = iota
	// refsNullable is a null or undefined literal, which participates in guard recognition.
	refsNullable
	// refsGuard is the result of comparing a ref value against null, as in `if (r.current == null)`.
	refsGuard
	// refsRef is the ref object itself, the thing `useRef` returns.
	refsRef
	// refsRefValue is what came OUT of a ref, the `current` property. Reading one during render is
	// the finding this rule is named for.
	refsRefValue
	// refsStructure is an object, array or function that carries a ref somewhere inside it.
	refsStructure
)

// refsAccessType is one lattice element.
//
// Flattened into a single struct with a kind tag rather than an interface hierarchy, because unlike
// `high_level_intermediate_representation.InstructionValue` this set is small, closed to this file, and every consumer switches on the
// kind anyway. The fields that are meaningful depend on Kind and the equality below is what keeps
// that honest.
type refsAccessType struct {
	Kind refsAccessKind

	// RefId identifies which ref this came from. Meaningful for refsGuard, refsRef and refsRefValue.
	//
	// A join MINTS A FRESH ONE, which is the whole reason `refsTypeEqual` must ignore it for
	// refsRef and refsRefValue. See the rule comment.
	RefId int
	// HasRefId separates "no ref id" from ref id zero, which a bare int cannot express.
	HasRefId bool

	// Span is where the ref value was accessed, carried so a finding can point at the access rather
	// than at the operand that happened to reach the check.
	Span       static_single_assignment.IdentifierId
	HasSpan    bool
	RefSpan    static_single_assignment.IdentifierId
	HasRefSpan bool

	// Value is what a structure carries inside it, nil when it carries nothing.
	Value *refsAccessType

	// Function is set on a structure that is a function, recording whether calling it reads a ref.
	Function *refsFunctionType
}

// refsFunctionType records what calling a function value would do to a ref.
type refsFunctionType struct {
	// ReadRefEffect is true when analysing the function's body produced a ref finding, so calling it
	// during render commits that finding at the CALL site rather than inside the function.
	ReadRefEffect bool
	// RefAccessSpan is where inside the function the ref was touched.
	RefAccessSpan    static_single_assignment.IdentifierId
	HasRefAccessSpan bool
	// ReturnType is what the function yields.
	ReturnType *refsAccessType
}

// refsTypeEqual is equality FOR CONVERGENCE, and it is deliberately not equality for identity.
//
// This reproduces upstream's hand-written `PartialEq`, and the reason it is hand-written rather
// than derived is the single most load-bearing detail in this rule. `refsJoin` mints a fresh ref id
// whenever it merges two different refs. A derived comparison would see that fresh id, report the
// environment as changed, and do so again on the next round: THE FIXPOINT WOULD NEVER CONVERGE, and
// every function carrying a joined ref would burn all ten rounds and then emit a spurious
// non-convergence finding.
//
// So: refsRef ignores its ref id entirely, and refsRefValue compares only its access span. Two
// genuinely distinct refs therefore compare EQUAL here. That is correct, because the question the
// sweep asks is "did this lattice element move", not "is this the same ref object".
//
// refsGuard is the exception and DOES compare its id, because a guard is only sound for the
// specific ref it guards; treating two guards as interchangeable would let `if (a.current == null)`
// authorize a write to `b.current`.
func refsTypeEqual(a, b *refsAccessType) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case refsNone, refsNullable:
		return true
	case refsGuard:
		return a.RefId == b.RefId && a.HasRefId == b.HasRefId
	case refsRef:
		// Ref identity deliberately ignored. See the doc comment.
		return true
	case refsRefValue:
		// Only the access span, deliberately. Ref origin and ref id ignored.
		return a.Span == b.Span && a.HasSpan == b.HasSpan
	case refsStructure:
		return refsTypeEqual(a.Value, b.Value) && refsFunctionEqual(a.Function, b.Function)
	}
	return false
}

// refsFunctionEqual compares two function types for convergence.
func refsFunctionEqual(a, b *refsFunctionType) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.ReadRefEffect == b.ReadRefEffect && refsTypeEqual(a.ReturnType, b.ReturnType)
}

// refsJoin merges two lattice elements, reproducing upstream's `join_ref_access_types`.
//
// The ordering of the cases is upstream's and is load-bearing: refsNone is the identity, a guard
// meeting a nullable collapses to refsNone (the guard is no longer sound once a non-guarded path
// reaches the same value), and two different guards also collapse. Everything that survives to the
// bottom is a ref-carrying element, where a mismatch mints a FRESH ref id. That mint is exactly the
// thing `refsTypeEqual` must not look at.
func refsJoin(a, b *refsAccessType, nextRefId func() int) *refsAccessType {
	if a == nil {
		a = &refsAccessType{Kind: refsNone}
	}
	if b == nil {
		b = &refsAccessType{Kind: refsNone}
	}
	if a.Kind == refsNone {
		return b
	}
	if b.Kind == refsNone {
		return a
	}
	if a.Kind == refsGuard && b.Kind == refsGuard {
		if a.RefId == b.RefId && a.HasRefId == b.HasRefId {
			return a
		}
		return &refsAccessType{Kind: refsNone}
	}
	// A guard meeting a nullable is no longer a guard: one path proved the ref null and the other
	// did not, so nothing downstream may rely on it.
	if (a.Kind == refsGuard && b.Kind == refsNullable) || (a.Kind == refsNullable && b.Kind == refsGuard) {
		return &refsAccessType{Kind: refsNone}
	}
	if a.Kind == refsGuard {
		return b
	}
	if b.Kind == refsGuard {
		return a
	}
	if a.Kind == refsNullable {
		return b
	}
	if b.Kind == refsNullable {
		return a
	}
	return refsJoinRefCarrying(a, b, nextRefId)
}

// refsJoinRefCarrying merges two elements that both carry a ref, upstream's
// `join_ref_access_ref_types`.
//
// A ref value joined with anything is a ref value with its provenance erased, because the result
// may or may not have come out of a ref and the conservative answer is that it did. Two refs that
// are not the same ref produce a fresh id, which is the mint the convergence equality ignores.
func refsJoinRefCarrying(a, b *refsAccessType, nextRefId func() int) *refsAccessType {
	if a.Kind == refsRefValue && b.Kind == refsRefValue {
		if a.HasRefId && b.HasRefId && a.RefId == b.RefId {
			return a
		}
		return &refsAccessType{Kind: refsRefValue}
	}
	if a.Kind == refsRefValue || b.Kind == refsRefValue {
		return &refsAccessType{Kind: refsRefValue}
	}
	if a.Kind == refsRef && b.Kind == refsRef {
		if a.HasRefId && b.HasRefId && a.RefId == b.RefId {
			return a
		}
		return &refsAccessType{Kind: refsRef, RefId: nextRefId(), HasRefId: true}
	}
	if a.Kind == refsRef || b.Kind == refsRef {
		return &refsAccessType{Kind: refsRef, RefId: nextRefId(), HasRefId: true}
	}
	// Both structures.
	joined := &refsAccessType{Kind: refsStructure}
	switch {
	case a.Function == nil:
		joined.Function = b.Function
	case b.Function == nil:
		joined.Function = a.Function
	default:
		returnType := refsJoin(a.Function.ReturnType, b.Function.ReturnType, nextRefId)
		merged := &refsFunctionType{
			ReadRefEffect: a.Function.ReadRefEffect || b.Function.ReadRefEffect,
			ReturnType:    returnType,
		}
		if a.Function.HasRefAccessSpan {
			merged.RefAccessSpan, merged.HasRefAccessSpan = a.Function.RefAccessSpan, true
		} else if b.Function.HasRefAccessSpan {
			merged.RefAccessSpan, merged.HasRefAccessSpan = b.Function.RefAccessSpan, true
		}
		joined.Function = merged
	}
	switch {
	case a.Value == nil:
		joined.Value = b.Value
	case b.Value == nil:
		joined.Value = a.Value
	default:
		joined.Value = refsJoinRefCarrying(a.Value, b.Value, nextRefId)
	}
	return joined
}

// refsJoinMany folds a slice, starting from the identity.
func refsJoinMany(types []*refsAccessType, nextRefId func() int) *refsAccessType {
	result := &refsAccessType{Kind: refsNone}
	for _, one := range types {
		result = refsJoin(result, one, nextRefId)
	}
	return result
}

// refsDestructure unwraps a structure down to what it carries, upstream's `destructure`.
//
// A structure holding a ref answers the validation checks as though it were that ref, which is what
// makes `const o = {r}; foo(o)` report where `foo(r)` reports.
func refsDestructure(t *refsAccessType) *refsAccessType {
	for t != nil && t.Kind == refsStructure && t.Value != nil {
		t = t.Value
	}
	if t == nil {
		return &refsAccessType{Kind: refsNone}
	}
	return t
}

// ---------------------------------------------------------------------------
// The environment and the sweep
// ---------------------------------------------------------------------------

// refsSafeBlock is a region where one write to a specific ref is permitted.
//
// Registered on the fallthrough of an `if` whose test is a guard, and CONSUMED by the first write
// it authorizes, which is what makes a second write under one guard still report.
type refsSafeBlock struct {
	Block static_single_assignment.BlockId
	RefId int
}

// refsFindingKind is which of the four diagnostics a finding is.
type refsFindingKind uint8

const (
	// refsFindingValueAccess is reading `ref.current` during render.
	refsFindingValueAccess refsFindingKind = iota
	// refsFindingPassedToFunction is handing a ref to a function that may read it.
	refsFindingPassedToFunction
	// refsFindingUpdate is writing `ref.current` during render.
	refsFindingUpdate
	// refsFindingFunctionAccessesRef is calling a function whose body reads a ref.
	refsFindingFunctionAccessesRef
	// refsFindingDidNotConverge is the invariant: the sweep was still moving after ten rounds.
	refsFindingDidNotConverge
)

// refsFinding is one diagnostic, held as data until the sweep settles.
//
// Deferred rather than reported inline for a reason that is upstream's and is easy to miss: the
// sweep runs the SAME instructions up to ten times, so reporting at the moment a check fails would
// emit each finding once per round. Upstream clears its error list at the top of every iteration
// and returns early once any error exists; this collects per-round and keeps the last round's set.
type refsFinding struct {
	Kind refsFindingKind
	// Value is the identifier to point the finding at.
	Value    static_single_assignment.IdentifierId
	HasValue bool
	// Node overrides where the finding points when the value alone would point at the wrong syntax.
	//
	// A ref access is a member expression, and upstream underlines the WHOLE of it (`ref.current`,
	// eleven characters in the goldens). The identifier a lattice value carries is `ref`, which is
	// the object rather than the access, so pointing through it underlines three characters and no
	// message-id fixture can see the difference. A read whose value is a pure temporary has no
	// identifier node at all and would otherwise fall back to the whole function.
	Node *ast.Node
}

// refsEnvironment is the per-identifier lattice map plus the temporaries side map.
//
// `temporaries` is upstream's `Env::temporaries`: an aliasing side table letting a load or a store
// resolve to the value it ultimately names, so a chain of assignments behaves like the value at its
// head without the lattice map carrying an entry per link.
type refsEnvironment struct {
	changed     bool
	data        map[static_single_assignment.IdentifierId]*refsAccessType
	temporaries map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId
	// names carries a source binding name onto the temporary that loaded it.
	//
	// Populated ONLY from LoadLocal, which is upstream's `set_name` and is called from exactly one
	// instruction arm (`infer_types.rs:482`). That single restriction is what makes
	// `props.fooRef.current` silent while `const fooRef = x; fooRef.current` reports: a property
	// load's result never receives a name, so the name test finds nothing to match on the outer
	// access. Propagating names through property loads as well would look more thorough and would
	// disagree with React on real code, measured on the executable in both directions.
	names map[static_single_assignment.IdentifierId]string
	// propertyNames is the property a load read, keyed by the value it produced. Consulted only by
	// the exact `props.ref` shape; see setPropertyName's call site.
	propertyNames map[static_single_assignment.IdentifierId]string
	// declarations and byDeclaration recover a value whose single-assignment numbering did not
	// unify. See `get`.
	// accessNodes is the member-expression node a ref VALUE came from, so a finding underlines the
	// whole access rather than only the object identifier.
	accessNodes   map[static_single_assignment.IdentifierId]*ast.Node
	declarations  map[static_single_assignment.IdentifierId]static_single_assignment.DeclarationId
	byDeclaration map[static_single_assignment.DeclarationId]*refsAccessType
	refIdSeed     int
}

func newRefsEnvironment() *refsEnvironment {
	return &refsEnvironment{
		data:          map[static_single_assignment.IdentifierId]*refsAccessType{},
		temporaries:   map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId{},
		names:         map[static_single_assignment.IdentifierId]string{},
		propertyNames: map[static_single_assignment.IdentifierId]string{},
		accessNodes:   map[static_single_assignment.IdentifierId]*ast.Node{},
		declarations:  map[static_single_assignment.IdentifierId]static_single_assignment.DeclarationId{},
		byDeclaration: map[static_single_assignment.DeclarationId]*refsAccessType{},
	}
}

// nextRefId mints a fresh ref identity. Every call site of this is a place the convergence equality
// must not look, which is why it is funnelled through one method.
func (env *refsEnvironment) nextRefId() int {
	env.refIdSeed++
	return env.refIdSeed
}

// operandId resolves an identifier through the temporaries side map.
func (env *refsEnvironment) operandId(key static_single_assignment.IdentifierId) static_single_assignment.IdentifierId {
	if resolved, found := env.temporaries[key]; found {
		return resolved
	}
	return key
}

func (env *refsEnvironment) define(key static_single_assignment.IdentifierId, value static_single_assignment.IdentifierId) {
	resolved := env.operandId(value)
	env.temporaries[key] = resolved
	// The access node travels with the value along an alias chain, so `const v = ref.current;`
	// still underlines `ref.current` when the finding is raised at the read of `v`.
	if node, ok := env.accessNodes[resolved]; ok {
		env.accessNodes[key] = node
	} else if node, ok := env.accessNodes[value]; ok {
		env.accessNodes[key] = node
	}
}

// carryAccessNode copies a recorded access node from one value to another.
func (env *refsEnvironment) carryAccessNode(to static_single_assignment.IdentifierId, from static_single_assignment.IdentifierId) {
	if node, ok := env.accessNodes[env.operandId(from)]; ok {
		env.accessNodes[to] = node
		return
	}
	if node, ok := env.accessNodes[from]; ok {
		env.accessNodes[to] = node
	}
}

// setName carries a source name onto a temporary, upstream's `set_name`.
func (env *refsEnvironment) setName(function *high_level_intermediate_representation.Function, target static_single_assignment.IdentifierId, source static_single_assignment.IdentifierId) {
	if name := refsIdentifierName(function, source); name != "" {
		env.names[target] = name
	}
}

// setName2 records a name directly onto a value, used for a global's own name.
func (env *refsEnvironment) setName2(target static_single_assignment.IdentifierId, name string) {
	if name != "" {
		env.names[target] = name
	}
}

// setPropertyName records the property a load read, as the result's name in this position.
func (env *refsEnvironment) setPropertyName(target static_single_assignment.IdentifierId, property string) {
	if property != "" {
		env.propertyNames[target] = property
	}
}

// nameOf is the value's own binding name, or the one a LoadLocal carried onto it.
func (env *refsEnvironment) nameOf(function *high_level_intermediate_representation.Function, id static_single_assignment.IdentifierId) string {
	if name := refsIdentifierName(function, id); name != "" {
		return name
	}
	return env.names[id]
}

func (env *refsEnvironment) get(key static_single_assignment.IdentifierId) *refsAccessType {
	if found, ok := env.data[env.operandId(key)]; ok {
		return found
	}
	// Fall back to whatever the same source BINDING holds elsewhere in this function.
	//
	// This exists because single-assignment renaming is not total in our lowering: measured on the
	// corpus fixture `error.invalid-aliased-ref-in-callback-invoked-during-render-`, a closure that
	// builds an alias chain produces two identifiers for one `const current`, ids 8 and 20, both
	// carrying declaration 3, and the LOAD reads id 8 while the STORE wrote id 20. Without the
	// alias in the same closure the two coincide and everything works, which is why this is
	// invisible on all but one fixture.
	//
	// So the value is looked up by declaration when the identifier itself has no entry. That is
	// sound here because a declaration groups exactly the values one source binding takes
	// (`high_level_intermediate_representation.Identifier.Declaration`), and it is conservative in the direction this rule wants: it
	// can only find a ref that some other numbering of the same binding already proved.
	//
	// This is a workaround for a representation defect rather than a transcription of upstream,
	// which has no such gap. It belongs in lowering; see the report.
	for _, candidate := range [2]static_single_assignment.IdentifierId{key, env.operandId(key)} {
		if declaration, ok := env.declarations[candidate]; ok {
			if found, ok := env.byDeclaration[declaration]; ok {
				return found
			}
		}
	}
	return nil
}

// set widens rather than overwrites, and records whether the element MOVED.
//
// The widening join is what makes the sweep monotone, and the `changed` flag it maintains is the
// fixpoint's only termination signal. The comparison uses `refsTypeEqual`, which ignores minted ref
// identity: reading that comparison as ordinary equality is the defect that would prevent
// convergence.
func (env *refsEnvironment) set(key static_single_assignment.IdentifierId, value *refsAccessType) {
	operandId := env.operandId(key)
	current, hadCurrent := env.data[operandId]
	widened := refsJoin(value, current, env.nextRefId)
	if !hadCurrent && widened.Kind == refsNone {
		env.data[operandId] = widened
		return
	}
	if !hadCurrent || !refsTypeEqual(current, widened) {
		env.changed = true
	}
	env.data[operandId] = widened
	// Record against the ORIGINAL key's declaration as well as the resolved one: an alias chain
	// resolves a named binding onto an unnamed temporary, so the declaration lives on the key that
	// was asked about rather than on the value it resolved to.
	if widened.Kind != refsNone {
		if declaration, ok := env.declarations[key]; ok {
			env.byDeclaration[declaration] = widened
		}
		if declaration, ok := env.declarations[operandId]; ok {
			env.byDeclaration[declaration] = widened
		}
	}
}

// noteDeclaration records which source binding a value belongs to, so `get` can recover a value
// whose single-assignment numbering did not unify with the one that was written.
func (env *refsEnvironment) noteDeclaration(function *high_level_intermediate_representation.Function, id static_single_assignment.IdentifierId) {
	if function == nil || int(id) >= len(function.Identifiers) {
		return
	}
	if identifier := function.Identifiers[id]; identifier != nil && identifier.Name != "" {
		env.declarations[id] = identifier.Declaration
	}
}

// refsSweepFunction runs the bounded ten-round sweep over one function and returns its findings.
//
// The structure is upstream's: seed the parameters, collect the temporaries side map once, then
// iterate the whole function in block order until the environment stops moving or ten rounds pass.
// `Blocks` is in reverse postorder, which is what lets a single forward pass settle an acyclic
// region in one round; a loop needs the second round, which is what the bound is for.
func refsSweepFunction(ctx rule.Context, function *high_level_intermediate_representation.Function) []refsFinding {
	env := newRefsEnvironment()
	refsCollectTemporaries(ctx, function, env)

	var findings []refsFinding
	for iteration := 0; iteration < 10; iteration++ {
		if iteration > 0 && !env.changed {
			break
		}
		env.changed = false
		findings = refsRunOneRound(ctx, function, env)
		// Upstream returns as soon as any error exists rather than iterating to a fixpoint over a
		// function it has already decided about. Reproduced: it also bounds the work.
		if len(findings) > 0 {
			return findings
		}
	}

	if env.changed {
		return []refsFinding{{Kind: refsFindingDidNotConverge}}
	}
	return findings
}

// refsCollectTemporaries builds the aliasing side map, upstream's `collect_temporaries_sidemap`.
//
// One detail is load-bearing and looks like an omission: a property load of `current` off a value
// that IS a ref is deliberately NOT aliased. That is what keeps `ref` and `ref.current` distinct,
// which is the entire distinction between "you passed a ref somewhere" and "you read a ref's
// value". Aliasing them would collapse two of the four diagnostics into one.
func refsCollectTemporaries(ctx rule.Context, function *high_level_intermediate_representation.Function, env *refsEnvironment) {
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := refsInstructionAt(function, instructionId)
			if instruction == nil {
				continue
			}
			env.noteDeclaration(function, instruction.LValue.Identifier)
			high_level_intermediate_representation.EachInstructionPlace(instruction, func(place high_level_intermediate_representation.Place, role high_level_intermediate_representation.PlaceRole) {
				env.noteDeclaration(function, place.Identifier)
			})
			switch value := instruction.Value.(type) {
			case *high_level_intermediate_representation.LoadLocal:
				env.setName(function, instruction.LValue.Identifier, value.Place.Identifier)
				env.define(instruction.LValue.Identifier, value.Place.Identifier)
			// There is deliberately NO LoadContext arm here, and it is worth saying why, because
			// adding one is the obvious move and it was in this file until a mutant proved it
			// inert.
			//
			// A captured binding is read through LoadContext rather than LoadLocal, so aliasing it
			// here looks necessary for `const aliasedRef = ref; aliasedRef.current` inside a
			// closure. It is not: the declaration fallback in `get` already recovers that case,
			// because both numberings of the alias carry the same DeclarationId. Measured three
			// ways: deleting this arm alone changes nothing across the whole suite, deleting the
			// fallback alone changes nothing, and deleting BOTH turns
			// `error.invalid-aliased-ref-in-callback-invoked-during-render-` silent. Two mechanisms
			// covering one case is one mechanism plus a thing that will drift, so the redundant one
			// is gone. Upstream has no LoadContext arm here either.
			case *high_level_intermediate_representation.StoreLocal:
				env.define(instruction.LValue.Identifier, value.Value.Identifier)
				env.define(value.LValue.Identifier, value.Value.Identifier)
			case *high_level_intermediate_representation.LoadGlobal:
				// A hook is called through a temporary that a LoadGlobal produced, so the callee's
				// own Name is empty and a hook test keyed on it never fires. Recording the global's
				// name here is what makes `useEffect(...)` recognizable as a hook call at all.
				// Without it every hook call is scored as an ordinary function call and a ref passed
				// to one reports spuriously, which was five of this corpus's residuals.
				env.setName2(instruction.LValue.Identifier, value.Name)
			case *high_level_intermediate_representation.PropertyLoad:
				// The loaded PROPERTY names the result, which is what carries `props.ref` into the
				// name test. Upstream reaches the same place by unifying a nested `Type::Property`,
				// whose `object_name` is resolved when the outer `.current` unifies against it
				// (`infer_types.rs:1353`); recording the property name here is that resolution
				// performed eagerly, and it is what makes `props.ref.current` report while
				// `props.fooRef.current` stays silent, because only the former's inner name is a
				// ref-like name that upstream's regex accepts in this position.
				env.setPropertyName(instruction.LValue.Identifier, value.Property)
				if value.Property == "current" && refsIsRefBinding(ctx, function, env, value.Object.Identifier) {
					continue
				}
				// A load whose RESULT is itself a ref is not aliased to its object either, for the
				// same reason the `current` case above is not: the alias would make the object and
				// the field indistinguishable, and the object here is a props bag rather than a ref.
				//
				// Without this, reading one ref-typed field marks the whole object as a ref, and
				// every OTHER field read off it then reports as a ref access. Measured on Kirk's
				// tree: an interface carrying two refs beside a `currentWidth: number` and a
				// `columnId: string` produced six findings in one file, all on the non-ref fields.
				// Upstream never has this problem because it types each property independently
				// rather than aliasing through the object.
				if refsIsUseRefType(ctx, function, instruction.LValue.Identifier) {
					continue
				}
				env.define(instruction.LValue.Identifier, value.Object.Identifier)
			}
		}
	}
}

// refsIsRefBinding is the two-signal ref test: the checker's type OR the binding's name.
//
// Both are required and each covers a part of the corpus the other cannot. See the rule comment for
// the measurements.
//
// # The name must be read through the aliasing side map, and this is the trap
//
// The name test asks for the binding name of the value whose `current` is being accessed, and the
// obvious way to ask that is to read `Identifier.Name` off the property load's object. That is
// WRONG and it is silently wrong: lowering emits `LoadLocal ref -> t13` and then
// `PropertyLoad t13.current`, so the object is the TEMPORARY, whose Name is empty, while the name
// lives on the source identifier the temporary loaded from. Asked directly, every `props.ref` case
// declines and looks exactly like a rule that simply does not implement them.
//
// Measured by dumping the representation for all three shapes rather than inferred: for
// `props.ref.current` the object is `t11` with an empty name, and `t11` resolves through the
// temporaries map to the value produced by the `ref` property load. `refsEnvironment.operandId` is
// upstream's `Env::operand_id` and is the resolution that recovers it.
func refsIsRefBinding(ctx rule.Context, function *high_level_intermediate_representation.Function, env *refsEnvironment, id static_single_assignment.IdentifierId) bool {
	if refsIsUseRefType(ctx, function, id) {
		return true
	}
	if refsIsRefLikeName(refsIdentifierName(function, id)) {
		return true
	}
	if env == nil {
		return false
	}
	if refsIsRefLikeName(env.nameOf(function, id)) {
		return true
	}
	// A property load names its result by the property read, but ONLY the exact name `ref` is
	// honoured here rather than the full regex. Measured on the executable: `props.ref.current`
	// reports while `props.fooRef.current` is clean, so the wider regex is wrong in this position
	// even though it is right for a binding. Reproduced as measured rather than as read.
	if env.propertyNames[id] == "ref" {
		return true
	}
	resolved := env.operandId(id)
	return resolved != id && refsIsUseRefType(ctx, function, resolved)
}

// refsRunOneRound is one pass over every block, phi and instruction.
//
// Returns the findings this round produced. Findings are recomputed each round rather than
// accumulated, because the same instruction is visited on every round and an accumulating list
// would report each finding up to ten times.
func refsRunOneRound(ctx rule.Context, function *high_level_intermediate_representation.Function, env *refsEnvironment) []refsFinding {
	var findings []refsFinding

	// Parameters are seeded from their declared type. A parameter that is a ref by type or by name
	// enters as one, which is how `function Component({ref})` reports with no `useRef` in sight.
	for _, param := range function.Params {
		env.set(param.Identifier, refsSeedType(ctx, function, param.Identifier, env))
	}

	// A value interpolated as a JSX child is checked more leniently than one passed to a function:
	// `<div>{value}</div>` where value came out of a ref reports, but a function returning a ref is
	// allowed to be rendered. Gathered once per round because a use can precede its definition
	// across a back edge.
	interpolatedAsJsx := map[static_single_assignment.IdentifierId]bool{}
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := refsInstructionAt(function, instructionId)
			if instruction == nil {
				continue
			}
			switch value := instruction.Value.(type) {
			case *high_level_intermediate_representation.JsxExpression:
				for _, child := range value.Children {
					interpolatedAsJsx[child.Identifier] = true
				}
			case *high_level_intermediate_representation.JsxFragment:
				for _, child := range value.Children {
					interpolatedAsJsx[child.Identifier] = true
				}
			}
		}
	}

	// safeBlocks records the fallthrough of an `if` whose test is a guard, so a write to the guarded
	// ref inside that region is permitted. This is what makes the `if (r.current == null) { r.current
	// = 1; }` initialization idiom clean while a bare write reports.
	var safeBlocks []refsSafeBlock

	for _, block := range function.Blocks {
		// Entering a block retires any safe region registered for it: the guard's protection ends
		// where the branch rejoins.
		filtered := safeBlocks[:0]
		for _, safe := range safeBlocks {
			if safe.Block != block.Id {
				filtered = append(filtered, safe)
			}
		}
		safeBlocks = filtered

		for _, phi := range block.Phis {
			operandTypes := make([]*refsAccessType, 0, len(phi.Operands))
			for _, entry := range phi.Operands {
				operandTypes = append(operandTypes, env.get(entry.Place.Identifier))
			}
			env.set(phi.Place.Identifier, refsJoinMany(operandTypes, env.nextRefId))
		}

		for _, instructionId := range block.Instructions {
			instruction := refsInstructionAt(function, instructionId)
			if instruction == nil {
				continue
			}
			findings = refsTransfer(ctx, function, env, instruction, interpolatedAsJsx, findings, &safeBlocks)

			// A guard value may only be consumed by an `if` test. Using one anywhere else means the
			// code read `ref.current` for its value rather than to check it against null.
			//
			// OPERANDS only. `EachInstructionPlace` also yields the instruction's own LValue, and
			// including it makes the guard report itself the moment it is created: the binary
			// expression that produces the guard has that guard as its lvalue, so the check fires on
			// the very instruction implementing `ref.current == null` and the sanctioned idiom
			// reports. Upstream walks `each_instruction_value_operand`, which is the value's
			// operands and not the lvalue. Measured: with the lvalue included,
			// `if (ref.current == null) { ref.current = 1; }` reports once and is clean upstream.
			high_level_intermediate_representation.EachPlace(instruction.Value, func(place high_level_intermediate_representation.Place, role high_level_intermediate_representation.PlaceRole) {
				if current := env.get(place.Identifier); current != nil && current.Kind == refsGuard {
					findings = append(findings, refsFinding{Kind: refsFindingValueAccess, Value: place.Identifier, HasValue: true})
				}
			})

			// Whatever the transfer decided, a value the checker calls a ref is a ref. This is
			// upstream's post-pass and it is what catches a ref arriving through a shape the
			// instruction switch does not model.
			refsApplyDeclaredType(ctx, function, env, instruction.LValue.Identifier)
		}

		if branch, isIf := block.Terminal.(*high_level_intermediate_representation.If); isIf {
			if test := env.get(branch.Test.Identifier); test != nil && test.Kind == refsGuard && test.HasRefId {
				alreadyRegistered := false
				for _, safe := range safeBlocks {
					if safe.RefId == test.RefId {
						alreadyRegistered = true
					}
				}
				if !alreadyRegistered {
					safeBlocks = append(safeBlocks, refsSafeBlock{Block: branch.Fallthrough, RefId: test.RefId})
				}
			}
		}

		_, isReturn := block.Terminal.(*high_level_intermediate_representation.Return)
		high_level_intermediate_representation.EachTerminalPlace(block.Terminal, func(place high_level_intermediate_representation.Place, role high_level_intermediate_representation.PlaceRole) {
			if isReturn {
				// Returning a ref object is allowed; returning a ref VALUE is not. That asymmetry is
				// what makes `return ref;` clean in a hook while `return ref.current;` reports.
				findings = refsCheckDirectValueAccess(env, place.Identifier, findings)
				return
			}
			findings = refsCheckValueAccess(env, place.Identifier, findings)
		})
	}

	return findings
}

// refsSeedType is a value's lattice element as decided by the checker and the name alone.
//
// Consulted whenever the instruction switch has nothing better to say. A value the checker types as
// `RefObject` enters as a ref; everything else enters as none.
func refsSeedType(ctx rule.Context, function *high_level_intermediate_representation.Function, id static_single_assignment.IdentifierId, env *refsEnvironment) *refsAccessType {
	if refsIsRefBinding(ctx, function, env, id) {
		return &refsAccessType{Kind: refsRef, RefId: env.nextRefId(), HasRefId: true}
	}
	return &refsAccessType{Kind: refsNone}
}

// refsApplyDeclaredType widens a value to a ref when the checker says it is one.
//
// Upstream runs this after every instruction, so a ref reaching a value through a shape the switch
// does not model is still recognized. Skipped when the value is already a ref, so the join does not
// mint an id on every round and stall the fixpoint.
func refsApplyDeclaredType(ctx rule.Context, function *high_level_intermediate_representation.Function, env *refsEnvironment, id static_single_assignment.IdentifierId) {
	if !refsIsRefBinding(ctx, function, env, id) {
		return
	}
	current := env.get(id)
	if current != nil && current.Kind == refsRef {
		return
	}
	// A value already known to be a FUNCTION is never widened into a ref by the name test. The
	// name heuristic is a guess and this is not: a closure named `setRef` is a function whose body
	// was just analysed, and letting `setRef` match the `...Ref$` regex would discard that analysis
	// and score the call as passing a ref rather than as calling a ref-reading function.
	//
	// Measured on the executable, because the collision is not hypothetical: a closure named
	// `setRef` and one named `doIt` both report, so upstream does not let the name win here either.
	// Three of this corpus's fixtures are exactly this shape and all three went silent before this
	// guard existed.
	if current != nil && current.Kind == refsStructure && current.Function != nil {
		return
	}
	env.set(id, &refsAccessType{Kind: refsRef, RefId: env.nextRefId(), HasRefId: true})
}

// refsCheckDirectValueAccess reports only a bare ref value, not a function that reads one.
//
// The looser of the two read checks. Used where a function carrying a ref is legitimately allowed
// to be there, such as a return value or a JSX child.
func refsCheckDirectValueAccess(env *refsEnvironment, id static_single_assignment.IdentifierId, findings []refsFinding) []refsFinding {
	current := env.get(id)
	if current == nil {
		return findings
	}
	if refsDestructure(current).Kind == refsRefValue {
		return append(findings, refsFinding{Kind: refsFindingValueAccess, Value: id, HasValue: true, Node: env.accessNodes[id]})
	}
	return findings
}

// refsCheckValueAccess reports a bare ref value OR a function whose body reads one.
func refsCheckValueAccess(env *refsEnvironment, id static_single_assignment.IdentifierId, findings []refsFinding) []refsFinding {
	current := env.get(id)
	if current == nil {
		return findings
	}
	resolved := refsDestructure(current)
	if resolved.Kind == refsRefValue {
		return append(findings, refsFinding{Kind: refsFindingValueAccess, Value: id, HasValue: true, Node: env.accessNodes[id]})
	}
	if resolved.Kind == refsStructure && resolved.Function != nil && resolved.Function.ReadRefEffect {
		return append(findings, refsFinding{Kind: refsFindingValueAccess, Value: id, HasValue: true, Node: env.accessNodes[id]})
	}
	return findings
}

// refsCheckPassedToFunction reports handing a ref, or a ref-reading function, to a call.
func refsCheckPassedToFunction(env *refsEnvironment, id static_single_assignment.IdentifierId, findings []refsFinding) []refsFinding {
	current := env.get(id)
	if current == nil {
		return findings
	}
	resolved := refsDestructure(current)
	switch {
	case resolved.Kind == refsRef || resolved.Kind == refsRefValue:
		return append(findings, refsFinding{Kind: refsFindingPassedToFunction, Value: id, HasValue: true})
	case resolved.Kind == refsStructure && resolved.Function != nil && resolved.Function.ReadRefEffect:
		return append(findings, refsFinding{Kind: refsFindingPassedToFunction, Value: id, HasValue: true})
	}
	return findings
}

// refsCheckUpdate reports writing through a ref during render.
func refsCheckUpdate(env *refsEnvironment, id static_single_assignment.IdentifierId, node *ast.Node, findings []refsFinding) []refsFinding {
	current := env.get(id)
	if current == nil {
		return findings
	}
	resolved := refsDestructure(current)
	if resolved.Kind == refsRef || resolved.Kind == refsRefValue {
		return append(findings, refsFinding{Kind: refsFindingUpdate, Value: id, HasValue: true, Node: node})
	}
	return findings
}

// refsTransfer applies one instruction to the environment and collects any findings it produces.
//
// The arms are upstream's, in upstream's order. The default arm is the conservative one: any
// instruction not modelled here checks every operand for a ref value, so a construct this port does
// not know about fails toward reporting rather than toward silence.
func refsTransfer(
	ctx rule.Context,
	function *high_level_intermediate_representation.Function,
	env *refsEnvironment,
	instruction *high_level_intermediate_representation.Instruction,
	interpolatedAsJsx map[static_single_assignment.IdentifierId]bool,
	findings []refsFinding,
	safeBlocks *[]refsSafeBlock,
) []refsFinding {
	target := instruction.LValue.Identifier

	switch value := instruction.Value.(type) {
	// Rendering a ref value is a read. A function that merely CARRIES a ref may be rendered, which
	// is why this is the direct check rather than the general one.
	case *high_level_intermediate_representation.JsxExpression, *high_level_intermediate_representation.JsxFragment:
		high_level_intermediate_representation.EachPlace(instruction.Value, func(place high_level_intermediate_representation.Place, role high_level_intermediate_representation.PlaceRole) {
			findings = refsCheckDirectValueAccess(env, place.Identifier, findings)
		})

	// Reading a property off a ref produces a ref VALUE, which is the element the whole rule is
	// about. Reading a property off a structure yields what the structure carries.
	case *high_level_intermediate_representation.PropertyLoad:
		findings = refsPropertyLoadTransfer(ctx, function, env, instruction, value.Object.Identifier, target, findings)

	case *high_level_intermediate_representation.ComputedLoad:
		findings = refsCheckDirectValueAccess(env, value.Property.Identifier, findings)
		findings = refsPropertyLoadTransfer(ctx, function, env, instruction, value.Object.Identifier, target, findings)

	// A cast is a no-op at runtime and carries its operand's element through unchanged.
	case *high_level_intermediate_representation.TypeCastExpression:
		env.set(target, refsCarryOrSeed(ctx, function, env, value.Value.Identifier, target))

	case *high_level_intermediate_representation.LoadLocal:
		env.carryAccessNode(target, value.Place.Identifier)
		env.set(target, refsCarryOrSeed(ctx, function, env, value.Place.Identifier, target))

	case *high_level_intermediate_representation.LoadContext:
		env.carryAccessNode(target, value.Place.Identifier)
		env.set(target, refsCarryOrSeed(ctx, function, env, value.Place.Identifier, target))

	case *high_level_intermediate_representation.StoreLocal:
		stored := refsCarryOrSeed(ctx, function, env, value.Value.Identifier, value.LValue.Identifier)
		env.carryAccessNode(value.LValue.Identifier, value.Value.Identifier)
		env.carryAccessNode(target, value.Value.Identifier)
		env.set(value.LValue.Identifier, stored)
		env.set(target, refsCarryOrSeed(ctx, function, env, value.Value.Identifier, target))
		// Pin the value under the WRITTEN binding's declaration. `env.set` records against whatever
		// the lvalue resolves to through the aliasing map, which for a store into a fresh binding is
		// an unnamed temporary carrying no declaration, so the entry the declaration fallback needs
		// would never be written. See `get` for why that fallback exists at all.
		env.noteDeclaration(function, value.LValue.Identifier)
		if declaration, ok := env.declarations[value.LValue.Identifier]; ok && stored.Kind != refsNone {
			env.byDeclaration[declaration] = stored
		}

	case *high_level_intermediate_representation.StoreContext:
		env.set(value.LValue.Identifier, refsCarryOrSeed(ctx, function, env, value.Value.Identifier, value.LValue.Identifier))
		env.set(target, refsCarryOrSeed(ctx, function, env, value.Value.Identifier, target))

	// Destructuring distributes what the source structure carries to every bound name, which is how
	// `const {current} = ref` reaches the same verdict as `ref.current`.
	case *high_level_intermediate_representation.Destructure:
		var carried *refsAccessType
		if source := env.get(value.Value.Identifier); source != nil && source.Kind == refsStructure && source.Value != nil {
			carried = source.Value
		}
		if carried == nil {
			carried = refsSeedType(ctx, function, target, env)
		}
		env.set(target, carried)
		high_level_intermediate_representation.EachPlace(instruction.Value, func(place high_level_intermediate_representation.Place, role high_level_intermediate_representation.PlaceRole) {
			if role == high_level_intermediate_representation.PlaceRoleDefine {
				env.set(place.Identifier, carried)
			}
		})

	// A nested function is analysed as an inner subject and its verdict is recorded ON THE VALUE.
	// A ref touched inside it does not report there; it reports where the function is CALLED.
	case *high_level_intermediate_representation.FunctionExpression:
		findings = refsNestedFunctionTransfer(ctx, function, env, instruction, value, target, findings)

	case *high_level_intermediate_representation.CallExpression:
		findings = refsCallTransfer(ctx, function, env, instruction, value.Callee.Identifier, interpolatedAsJsx, findings)

	// A method call keeps its receiver, so `object.foo()` arrives here rather than as a property
	// load plus a call. The callee to score is the PROPERTY, but the property is a fresh temporary
	// with no lattice entry of its own: what carries the function is the RECEIVER's structure. So
	// the property is seeded from the receiver first, and only then scored.
	//
	// Without that seeding a closure stored on an object and called through it scores as an
	// ordinary call with a ref-carrying argument, which reports the wrong one of the four
	// diagnostics. Two corpus fixtures are this shape and both said `refPassedToFunction` where
	// upstream says `functionAccessesRef`.
	case *high_level_intermediate_representation.MethodCall:
		// `React.useEffect(...)` is a METHOD call, so the hook's name is the property rather than a
		// binding, and the hook test cannot see it without this. Recording it is what makes the
		// namespaced spelling behave like the bare import.
		//
		// This was not a fixture failure and could not have been: every corpus fixture calls hooks
		// bare. It was 235 findings on Kirk's tree, essentially all of them a ref touched inside a
		// `React.useEffect` or `React.useCallback` callback, which upstream is silent on (measured
		// on both shapes). A rule scoring 30 of 30 was reporting a false positive in almost every
		// React file in the repository.
		if refsMethodName(function, instruction) != "" {
			env.setName2(value.Property.Identifier, refsMethodName(function, instruction))
		}
		if receiverType := env.get(value.Receiver.Identifier); receiverType != nil &&
			receiverType.Kind == refsStructure && receiverType.Function != nil {
			env.set(value.Property.Identifier, receiverType)
		}
		findings = refsCallTransfer(ctx, function, env, instruction, value.Property.Identifier, interpolatedAsJsx, findings)

	// An aggregate becomes a structure carrying the join of what its elements carry, so a ref put
	// into an object is still findable through the object.
	case *high_level_intermediate_representation.ObjectExpression, *high_level_intermediate_representation.ArrayExpression:
		var elementTypes []*refsAccessType
		high_level_intermediate_representation.EachPlace(instruction.Value, func(place high_level_intermediate_representation.Place, role high_level_intermediate_representation.PlaceRole) {
			// An aggregate feeding a call whose RESULT is a ref is the mergeRefs shape, and
			// collecting refs into a list is the entire point of it. Upstream reaches the same
			// silence through the aliasing effects of the call rather than at the aggregate, so the
			// exemption is expressed here instead and this is a divergence in mechanism that agrees
			// in outcome.
			//
			// Measured: `mergeReferences([inputReference, ref])` is clean upstream, and it was 8
			// findings on Kirk's tree, one in every form-field component that forwards a ref. No
			// corpus fixture merges refs, so nothing imported can see this either way.
			// The exemption covers a ref OBJECT only, never a ref VALUE. Collecting refs into a
			// list or an options object and handing it to a hook is ordinary React; reading
			// `ref.current` into that same object is a render-time read whatever the object is for.
			//
			// Measured with three inputs on the executable: `useThing({r: r})` is clean,
			// `useThing({v: r.current})` reports, and `useEffect(() => {}, [ref.current])` reports
			// twice. The corpus fixture `error.hook-ref-value` is that third shape and it is the
			// test that caught this exemption when it was written one step too wide.
			isRefObject := false
			if current := env.get(place.Identifier); current != nil {
				isRefObject = refsDestructure(current).Kind == refsRef
			}
			if !isRefObject || !refsAggregateIsExempt(ctx, function, env, instruction) {
				findings = refsCheckDirectValueAccess(env, place.Identifier, findings)
			}
			elementTypes = append(elementTypes, env.get(place.Identifier))
		})
		joined := refsJoinMany(elementTypes, env.nextRefId)
		switch joined.Kind {
		case refsNone, refsGuard, refsNullable:
			env.set(target, &refsAccessType{Kind: refsNone})
		default:
			env.set(target, &refsAccessType{Kind: refsStructure, Value: joined})
		}

	case *high_level_intermediate_representation.PropertyStore:
		findings = refsStoreTransfer(env, instruction, value.Object.Identifier, &value.Value, nil, findings, safeBlocks)

	case *high_level_intermediate_representation.ComputedStore:
		findings = refsStoreTransfer(env, instruction, value.Object.Identifier, &value.Value, &value.Property, findings, safeBlocks)

	case *high_level_intermediate_representation.PropertyDelete:
		findings = refsCheckUpdate(env, value.Object.Identifier, instruction.Node, findings)

	case *high_level_intermediate_representation.ComputedDelete:
		findings = refsCheckUpdate(env, value.Object.Identifier, instruction.Node, findings)
		findings = refsCheckValueAccess(env, value.Property.Identifier, findings)

	// `undefined` and `null` are nullable, which is only interesting because a comparison against
	// one of them is what turns a ref value into a guard.
	case *high_level_intermediate_representation.LoadGlobal:
		if value.Name == "undefined" {
			env.set(target, &refsAccessType{Kind: refsNullable})
		}

	case *high_level_intermediate_representation.Primitive:
		if value.Value == nil {
			env.set(target, &refsAccessType{Kind: refsNullable})
		}

	// `!ref.current` is both a read AND a guard: it reports, and it marks the result so the write it
	// authorizes does not report a second time on the same line.
	case *high_level_intermediate_representation.UnaryExpression:
		if value.Operator == "!" {
			if operand := env.get(value.Value.Identifier); operand != nil && operand.Kind == refsRefValue && operand.HasRefId {
				env.set(target, &refsAccessType{Kind: refsGuard, RefId: operand.RefId, HasRefId: true})
				findings = append(findings, refsFinding{Kind: refsFindingValueAccess, Value: value.Value.Identifier, HasValue: true})
				break
			}
		}
		findings = refsCheckValueAccess(env, value.Value.Identifier, findings)

	// Comparing a ref value against null produces a guard rather than a finding. Comparing it
	// against anything else is an ordinary read and reports.
	case *high_level_intermediate_representation.BinaryExpression:
		findings = refsBinaryTransfer(env, instruction, value, target, findings)

	case *high_level_intermediate_representation.StartMemoize, *high_level_intermediate_representation.FinishMemoize:
		// Markers. Nothing to do, and named explicitly so they do not fall into the default arm's
		// operand check, which would report a memoized ref.

	default:
		high_level_intermediate_representation.EachPlace(instruction.Value, func(place high_level_intermediate_representation.Place, role high_level_intermediate_representation.PlaceRole) {
			findings = refsCheckValueAccess(env, place.Identifier, findings)
		})
	}

	return findings
}

// refsCallResultIsRef asks the checker what a CALL yields, through the call's own syntax.
//
// `refsIsRefBinding` reaches the checker through `Identifier.Node`, which a call's temporary lvalue
// does not have. This asks the same question of the call expression instead.
func refsCallResultIsRef(ctx rule.Context, instruction *high_level_intermediate_representation.Instruction) bool {
	if ctx.TypeChecker == nil || instruction == nil || instruction.Node == nil {
		return false
	}
	node := instruction.Node
	if node.Kind != ast.KindCallExpression {
		return false
	}
	resultType := ctx.TypeChecker.GetTypeAtLocation(node)
	if resultType == nil {
		return false
	}
	if alias := shimchecker.Type_alias(resultType); alias != nil {
		if aliasSymbol := alias.Symbol(); aliasSymbol != nil && aliasSymbol.Name == "RefCallback" {
			return true
		}
	}
	if symbol := shimchecker.Type_symbol(resultType); symbol != nil {
		// `bivarianceHack` is what the REAL `@types/react` `RefCallback` resolves to, and matching it
		// is not a workaround on our side.
		//
		// `RefCallback<T>` is declared as an indexed access over an object type whose single member
		// is named `bivarianceHack` (`@types/react/index.d.ts:176-185`), the standard idiom for
		// opting one member into bivariant parameter checking. Resolving that indexed access erases
		// the alias and leaves the method's own symbol, so the type arrives with a nil alias and the
		// symbol `bivarianceHack`.
		//
		// Measured on Kirk's tree rather than reasoned about, and a hand-written stub does NOT
		// reproduce it: a fixture declaring `type RefCallback<T> = (i: T) => void` carries the alias
		// and passes, which is precisely the failure where code and fixtures share one wrong belief.
		// The eight findings this silenced were all `mergeReferences([inner, ref])`, one in every
		// form-field component that forwards a ref, and all clean upstream.
		return symbol.Name == "RefObject" || symbol.Name == "MutableRefObject" ||
			symbol.Name == "bivarianceHack"
	}
	return false
}

// refsAggregateIsExempt reports whether an aggregate is consumed by a call that may hold refs.
//
// Asked at the AGGREGATE rather than at the call, because an argument list is evaluated first and
// would otherwise have already reported by the time the call's own exemptions are considered. Two
// consumers qualify, and both are upstream's:
//
//   - a call whose RESULT is a ref, which is the mergeRefs shape
//   - a HOOK call, which upstream exempts because a hook is independently validated
//
// Both were measured on the executable with live controls rather than inferred. Collecting refs
// into an object and handing it to `useTableColumnResize` is clean; handing the identical object to
// a plain `doThing` reports; and a plain `useRef` read during render still reports, which is the
// control proving the probe could see a finding at all.
//
// The state hooks are deliberately NOT exempt here, for the same reason they are not at the call:
// `useState`, `useReducer` and `useMemo` run what they are given during render.
func refsAggregateIsExempt(ctx rule.Context, function *high_level_intermediate_representation.Function, env *refsEnvironment, instruction *high_level_intermediate_representation.Instruction) bool {
	aggregate := instruction.LValue.Identifier
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			candidate := refsInstructionAt(function, instructionId)
			if candidate == nil || candidate.Id <= instruction.Id {
				continue
			}
			var args []high_level_intermediate_representation.Argument
			switch call := candidate.Value.(type) {
			case *high_level_intermediate_representation.CallExpression:
				args = call.Args
			case *high_level_intermediate_representation.MethodCall:
				args = call.Args
			default:
				continue
			}
			for _, argument := range args {
				if env.operandId(argument.Place.Identifier) != env.operandId(aggregate) {
					continue
				}
				if refsIsRefBinding(ctx, function, env, candidate.LValue.Identifier) ||
					refsCallResultIsRef(ctx, candidate) {
					return true
				}
				calleeName := ""
				switch call := candidate.Value.(type) {
				case *high_level_intermediate_representation.CallExpression:
					calleeName = env.nameOf(function, call.Callee.Identifier)
				case *high_level_intermediate_representation.MethodCall:
					calleeName = refsMethodName(function, candidate)
				}
				if refsIsHookName(calleeName) && calleeName != "useState" &&
					calleeName != "useReducer" && calleeName != "useMemo" {
					return true
				}
			}
		}
	}
	return false
}

// refsCarryOrSeed passes an operand's element through, falling back to its declared type.
func refsCarryOrSeed(ctx rule.Context, function *high_level_intermediate_representation.Function, env *refsEnvironment, from static_single_assignment.IdentifierId, to static_single_assignment.IdentifierId) *refsAccessType {
	if current := env.get(from); current != nil {
		return current
	}
	return refsSeedType(ctx, function, to, env)
}

// refsPropertyLoadTransfer decides what reading a property off a value yields.
//
// Off a ref, it is a ref VALUE carrying where the access happened. Off a structure, it is whatever
// the structure carries. Otherwise it falls back to the declared type.
func refsPropertyLoadTransfer(
	ctx rule.Context,
	function *high_level_intermediate_representation.Function,
	env *refsEnvironment,
	instruction *high_level_intermediate_representation.Instruction,
	object static_single_assignment.IdentifierId,
	target static_single_assignment.IdentifierId,
	findings []refsFinding,
) []refsFinding {
	objectType := env.get(object)
	switch {
	// A structure carrying a function hands that function out. Without this arm a closure stored on
	// an object property and then called through it loses its `readRefEffect`, and the call scores
	// as passing a ref rather than as calling a ref-reading function. Two corpus fixtures are this
	// exact shape (`object.foo = () => ref.current; object.foo()`), and upstream reaches it because
	// its structure carries the function in the same slot the value uses.
	case objectType != nil && objectType.Kind == refsStructure && objectType.Function != nil && objectType.Value == nil:
		env.set(target, objectType)
	case objectType != nil && objectType.Kind == refsStructure && objectType.Value != nil:
		env.set(target, objectType.Value)
	case objectType != nil && objectType.Kind == refsRef:
		if instruction.Node != nil {
			env.accessNodes[target] = instruction.Node
		}
		env.set(target, &refsAccessType{
			Kind:       refsRefValue,
			Span:       target,
			HasSpan:    true,
			RefSpan:    object,
			HasRefSpan: true,
			RefId:      objectType.RefId,
			HasRefId:   objectType.HasRefId,
		})
	default:
		env.set(target, refsSeedType(ctx, function, target, env))
	}
	return findings
}

// refsNestedFunctionTransfer analyses an inner function and records its verdict on the value.
//
// This is what makes a ref touched inside a callback report at the CALL rather than at the touch.
// `capture-ref-for-mutation` is the shipped fixture: the ref is read inside `handleKey`, and the
// finding lands on `handleKey('left')()` because that is where the reading actually happens during
// render.
//
// The inner sweep shares the OUTER environment rather than starting fresh, which is upstream's
// choice and is what lets a captured ref be visible inside the nested function at all.
func refsNestedFunctionTransfer(
	ctx rule.Context,
	function *high_level_intermediate_representation.Function,
	env *refsEnvironment,
	instruction *high_level_intermediate_representation.Instruction,
	value *high_level_intermediate_representation.FunctionExpression,
	target static_single_assignment.IdentifierId,
	findings []refsFinding,
) []refsFinding {
	inner := refsNestedFunction(function, value.Function)
	if inner == nil {
		env.set(target, &refsAccessType{Kind: refsNone})
		return findings
	}

	// The inner function gets its OWN environment, and this is the single most load-bearing line in
	// the nested analysis.
	//
	// Identifier ids are per-function and they OVERLAP. Measured by dumping the representation for
	// the corpus's `capture-ref-for-mutation` fixture: the outer function uses ids 18 through 35 and
	// the innermost closure independently uses 20 through 37. Sweeping the inner function into the
	// outer environment therefore conflates two unrelated values at every shared id, and the failure
	// is not a dropped finding but a WRONG one: the inner `pos`, a ref value, lands on an outer id
	// that the outer function then passes to a call, so the rule reported `refPassedToFunction` in a
	// function that never passed a ref anywhere. Upstream reports once at the outer call site and
	// nothing at the inner read, measured on the executable.
	//
	// This is the trap the porting brief names as "a value identifier does not survive a function
	// boundary", and it arrives here in its most misleading form: the id is VALID in both spaces, so
	// nothing is nil, nothing panics, and every message-id fixture that does not involve a nested
	// closure stays green.
	inner_env := newRefsEnvironment()
	inner_env.refIdSeed = env.refIdSeed

	// A captured value must be translated into the inner function's identifier space before the
	// inner sweep can see it. `Function.Context` is the capture edge that performs the translation:
	// the Nth capture at the call site corresponds to the Nth context value inside.
	for index, captured := range value.Captures {
		if index >= len(inner.Context) {
			continue
		}
		contextId := inner.Context[index].Identifier
		if outerType := env.get(captured.Identifier); outerType != nil {
			inner_env.data[contextId] = outerType
		}
		// The captured value's NAME crosses the boundary too. Without it a ref that the outer
		// function knows only by name rather than by type arrives inside as an anonymous context
		// value, and the name signal cannot fire on the alias chain the closure builds from it.
		// `error.invalid-aliased-ref-in-callback-invoked-during-render-` is exactly that shape:
		// `const aliasedRef = ref; aliasedRef.current` inside a callback.
		if name := env.nameOf(function, captured.Identifier); name != "" {
			inner_env.names[contextId] = name
		}
	}

	refsCollectTemporaries(ctx, inner, inner_env)
	innerFindings := refsRunOneRound(ctx, inner, inner_env)
	// Ref identity is minted from a single counter so two functions cannot collide on an id; the
	// seed travels back out rather than the whole environment.
	env.refIdSeed = inner_env.refIdSeed
	functionType := &refsFunctionType{ReturnType: &refsAccessType{Kind: refsNone}}
	if len(innerFindings) == 0 {
		if returned := inner_env.get(refsReturnedValue(inner)); returned != nil {
			functionType.ReturnType = returned
		}
	} else {
		functionType.ReadRefEffect = true
		for _, finding := range innerFindings {
			if finding.HasValue {
				functionType.RefAccessSpan, functionType.HasRefAccessSpan = finding.Value, true
				break
			}
		}
	}
	env.set(target, &refsAccessType{Kind: refsStructure, Function: functionType})
	return findings
}

// refsReturnedValue is the identifier a function actually yields.
//
// `Function.Returns` is the documented seam and it is NOT sufficient on its own: lowering populates
// it only from an explicit `return` statement (`lower.go:596`), while a concise arrow body
// terminates with `Return{Value: value}` and leaves `Returns` an empty temporary
// (`lower.go:164-166`). So `d => () => {...}` yields nothing through `Returns`, and a rule reading
// only that field loses the curried closure entirely.
//
// Measured rather than reasoned: dumping the representation for the corpus's curried fixture shows
// `handleKey`'s `Returns` as id 4 while its `FunctionExpression` writes to id 5. Reading the
// terminal recovers the real value, and `Returns` stays the fallback for the explicit-return case
// where the terminal names it anyway.
//
// This is a gap in lowering rather than in this rule, and it is worked around here rather than
// repaired there because a change to `Returns` would move every pass built on this representation
// at once. It deserves a fix upstream of this file; see the report.
func refsReturnedValue(function *high_level_intermediate_representation.Function) static_single_assignment.IdentifierId {
	for _, block := range function.Blocks {
		if terminal, isReturn := block.Terminal.(*high_level_intermediate_representation.Return); isReturn {
			return terminal.Value.Identifier
		}
	}
	return function.Returns.Identifier
}

// refsMethodName is the property being called in a method call, read from the syntax.
//
// The representation gives a method call its receiver and a Place for the property, but the
// property's NAME lives on the syntax rather than on the value, so it is read back from the node.
func refsMethodName(function *high_level_intermediate_representation.Function, instruction *high_level_intermediate_representation.Instruction) string {
	if instruction == nil || instruction.Node == nil {
		return ""
	}
	node := instruction.Node
	if node.Kind == ast.KindCallExpression {
		if expression := node.Expression(); expression != nil {
			node = expression
		}
	}
	if node.Kind != ast.KindPropertyAccessExpression {
		return ""
	}
	name := node.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.Text()
}

// refsNestedFunction resolves a function id within its parent.
func refsNestedFunction(function *high_level_intermediate_representation.Function, id high_level_intermediate_representation.FunctionId) *high_level_intermediate_representation.Function {
	if function == nil || int(id) >= len(function.Functions) {
		return nil
	}
	return function.Functions[id]
}

// refsCallTransfer decides what calling a value does, and it is the busiest arm.
//
// Three questions in order: does the callee itself read a ref (report at the callee), is the result
// a ref or is this a hook (then only direct ref values in the arguments matter), and otherwise does
// passing these arguments risk a read.
//
// The hook exemption is upstream's and is narrow: a hook call is independently validated, so a ref
// passed to one is allowed, EXCEPT for `useState` and `useReducer`, whose initializers run during
// render and so genuinely would read the ref then.
func refsCallTransfer(
	ctx rule.Context,
	function *high_level_intermediate_representation.Function,
	env *refsEnvironment,
	instruction *high_level_intermediate_representation.Instruction,
	callee static_single_assignment.IdentifierId,
	interpolatedAsJsx map[static_single_assignment.IdentifierId]bool,
	findings []refsFinding,
) []refsFinding {
	target := instruction.LValue.Identifier
	returnType := &refsAccessType{Kind: refsNone}
	didError := false

	if calleeType := env.get(callee); calleeType != nil && calleeType.Kind == refsStructure && calleeType.Function != nil {
		if calleeType.Function.ReturnType != nil {
			returnType = calleeType.Function.ReturnType
		}
		if calleeType.Function.ReadRefEffect {
			didError = true
			findings = append(findings, refsFinding{Kind: refsFindingFunctionAccessesRef, Value: callee, HasValue: true})
		}
	}

	if !didError {
		// The call's own node is consulted as well as the target value, because a call's lvalue is a
		// temporary with no `Identifier.Node`, so asking the checker about the value alone answers
		// false for every call. Handing it the CALL EXPRESSION recovers the result type, which is
		// what makes the mergeRefs shape (`mergeReferences(...)` returning `RefCallback`) exempt.
		//
		// Measured on Kirk's tree: without this the eight `mergeReferences([inner, ref])` sites all
		// report, and upstream is silent on that exact input.
		isRefResult := refsIsRefBinding(ctx, function, env, target) ||
			refsCallResultIsRef(ctx, instruction)
		hookName := env.nameOf(function, callee)
		isHook := refsIsHookName(hookName)
		// `useState`, `useReducer` and `useMemo` are NOT exempt, because each runs the argument it
		// is given during render: a state initializer and a memo callback both execute on the spot,
		// so a ref reaching one is genuinely read during render. Every other hook defers.
		//
		// Upstream's own condition names only UseState and UseReducer (`:691`), and reading it
		// would have produced a rule that lets `useMemo(() => foo(ref), [])` through. Measured on
		// the executable across ten hooks, `useMemo` REPORTS, and the fixture
		// `error.maybe-mutable-ref-not-preserved` is exactly that shape. oxc reaches the same
		// verdict by a different route (the memo callback is lowered and its body analysed rather
		// than the call being exempted), so this is a divergence in mechanism that agrees in
		// outcome, recorded here rather than silently matched.
		isStateHook := hookName == "useState" || hookName == "useReducer" || hookName == "useMemo"

		switch {
		// Producing a ref (the `mergeRefs` shape) or calling a hook: only a bare ref VALUE in the
		// arguments is a problem, because the ref object itself is what these legitimately take.
		case isRefResult || (isHook && !isStateHook):
			high_level_intermediate_representation.EachPlace(instruction.Value, func(place high_level_intermediate_representation.Place, role high_level_intermediate_representation.PlaceRole) {
				findings = refsCheckDirectValueAccess(env, place.Identifier, findings)
			})
		// The result is rendered, so a function carrying a ref is acceptable but a ref value is not.
		case interpolatedAsJsx[target]:
			high_level_intermediate_representation.EachPlace(instruction.Value, func(place high_level_intermediate_representation.Place, role high_level_intermediate_representation.PlaceRole) {
				findings = refsCheckValueAccess(env, place.Identifier, findings)
			})
		// An ordinary call: passing a ref to it may read the ref during render.
		default:
			high_level_intermediate_representation.EachPlace(instruction.Value, func(place high_level_intermediate_representation.Place, role high_level_intermediate_representation.PlaceRole) {
				if place.Identifier == callee {
					return
				}
				findings = refsCheckPassedToFunction(env, place.Identifier, findings)
			})
		}
	}

	env.set(target, returnType)
	return findings
}

// refsStoreTransfer handles writing through a property or a computed key.
//
// The safe-block consultation is what permits the guarded initialization idiom: inside the
// fallthrough of `if (ref.current == null)`, one write to that specific ref is allowed, and the
// permission is CONSUMED so a second write on the same guard still reports. That consumption is why
// `ref-initialization-linear`, which writes twice under one guard, reports exactly once.
func refsStoreTransfer(
	env *refsEnvironment,
	instruction *high_level_intermediate_representation.Instruction,
	object static_single_assignment.IdentifierId,
	stored *high_level_intermediate_representation.Place,
	computedKey *high_level_intermediate_representation.Place,
	findings []refsFinding,
	safeBlocks *[]refsSafeBlock,
) []refsFinding {
	foundSafe := false
	if target := env.get(object); target != nil && target.Kind == refsRef && target.HasRefId && computedKey == nil {
		for index, safe := range *safeBlocks {
			if safe.RefId == target.RefId {
				*safeBlocks = append((*safeBlocks)[:index], (*safeBlocks)[index+1:]...)
				foundSafe = true
				break
			}
		}
	}
	if !foundSafe {
		findings = refsCheckUpdate(env, object, instruction.Node, findings)
	}
	if computedKey != nil {
		findings = refsCheckValueAccess(env, computedKey.Identifier, findings)
	}
	if stored != nil {
		findings = refsCheckDirectValueAccess(env, stored.Identifier, findings)
		// Storing a structure into an object makes the object carry what the structure carried, so a
		// ref reaching an object through a property write is still findable.
		if storedType := env.get(stored.Identifier); storedType != nil && storedType.Kind == refsStructure {
			merged := storedType
			if target := env.get(object); target != nil {
				merged = refsJoin(merged, target, env.nextRefId)
			}
			env.set(object, merged)
		}
	}
	return findings
}

// refsBinaryTransfer recognizes the null comparison that turns a ref value into a guard.
//
// `ref.current == null` is the sanctioned way to ask whether a ref has been initialized, so it is
// clean and its result authorizes one write. Comparing a ref value against anything else is an
// ordinary read and reports.
func refsBinaryTransfer(
	env *refsEnvironment,
	instruction *high_level_intermediate_representation.Instruction,
	value *high_level_intermediate_representation.BinaryExpression,
	target static_single_assignment.IdentifierId,
	findings []refsFinding,
) []refsFinding {
	leftType := env.get(value.Left.Identifier)
	rightType := env.get(value.Right.Identifier)

	foundRefId, hasRefId := 0, false
	if leftType != nil && leftType.Kind == refsRefValue && leftType.HasRefId {
		foundRefId, hasRefId = leftType.RefId, true
	} else if rightType != nil && rightType.Kind == refsRefValue && rightType.HasRefId {
		foundRefId, hasRefId = rightType.RefId, true
	}

	nullish := (leftType != nil && leftType.Kind == refsNullable) || (rightType != nil && rightType.Kind == refsNullable)

	if hasRefId && nullish {
		env.set(target, &refsAccessType{Kind: refsGuard, RefId: foundRefId, HasRefId: true})
		return findings
	}
	findings = refsCheckValueAccess(env, value.Left.Identifier, findings)
	findings = refsCheckValueAccess(env, value.Right.Identifier, findings)
	return findings
}

// ---------------------------------------------------------------------------
// Reporting
// ---------------------------------------------------------------------------

// The four message ids, matching upstream's four distinct diagnostics for this category. They share
// one headline ("Cannot access refs during render") upstream and differ in the label under the span,
// which is the text reproduced in each Description's first sentence.
const (
	messageRefsValueAccessId         = "refValueAccess"
	messageRefsPassedToFunctionId    = "refPassedToFunction"
	messageRefsUpdateId              = "refUpdate"
	messageRefsFunctionAccessesRefId = "functionAccessesRef"
	messageRefsDidNotConvergeId      = "refTypeEnvironmentDidNotConverge"
)

// refsWhyRefsAreOutsideRender is the shared explanation, held once so a single edit moves every
// finding and the four Descriptions cannot drift apart.
const refsWhyRefsAreOutsideRender = "A ref is deliberately outside React's data flow: writing one " +
	"does not schedule a render, and during render the ref may not yet point at the element it will " +
	"point at once the tree commits. Reading it here therefore gives a value that is stale or not " +
	"yet attached, and because the read creates no subscription nothing re-renders when it changes. " +
	"Move the access into an event handler or an effect, where the ref is attached and a stale read " +
	"cannot be observed."

// refsMessageFor renders the finding.
func refsMessageFor(kind refsFindingKind) rule.Message {
	switch kind {
	case refsFindingPassedToFunction:
		return rule.Message{
			Id: messageRefsPassedToFunctionId,
			Description: "Passing a ref to a function may read its value during render. " +
				refsWhyRefsAreOutsideRender,
		}
	case refsFindingUpdate:
		return rule.Message{
			Id:          messageRefsUpdateId,
			Description: "Cannot update ref during render. " + refsWhyRefsAreOutsideRender,
		}
	case refsFindingFunctionAccessesRef:
		return rule.Message{
			Id: messageRefsFunctionAccessesRefId,
			Description: "This function accesses a ref value, and calling it during render reads " +
				"the ref. " + refsWhyRefsAreOutsideRender,
		}
	case refsFindingDidNotConverge:
		return rule.Message{
			Id: messageRefsDidNotConvergeId,
			Description: "The ref analysis did not settle after ten passes over this function, so " +
				"its result is not trustworthy and no ref finding is reported for it. This is an " +
				"internal invariant rather than a problem with the code, and it reproduces " +
				"upstream's own bound.",
		}
	default:
		return rule.Message{
			Id:          messageRefsValueAccessId,
			Description: "Cannot access ref value during render. " + refsWhyRefsAreOutsideRender,
		}
	}
}

// refsReport points a finding at the syntax its value came from.
//
// Through `ctx.ReportNode` rather than a Place's range, because a Place's range is the range of the
// node that produced the INSTRUCTION, and reporting through the node routes via `rule.TokenRange`,
// which trims leading trivia at the harness. A finding whose value is a pure temporary with no
// syntactic source falls back to the function's own node rather than being dropped.
func refsReport(ctx rule.Context, function *high_level_intermediate_representation.Function, finding refsFinding) {
	message := refsMessageFor(finding.Kind)
	if finding.Node != nil {
		ctx.ReportNode(finding.Node, message)
		return
	}
	if finding.HasValue {
		if node := refsIdentifierNode(function, finding.Value); node != nil {
			ctx.ReportNode(node, message)
			return
		}
	}
	if function.Node != nil {
		ctx.ReportNode(function.Node, message)
	}
}
