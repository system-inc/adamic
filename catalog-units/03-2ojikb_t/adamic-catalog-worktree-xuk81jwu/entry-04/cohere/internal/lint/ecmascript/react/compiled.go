package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Which functions React Compiler compiles, answered once for every rule that needs to know.
//
// Lifted from `react-hooks/unsupported-syntax`, where it was written against the compiler's own
// driver and measured case by case, and where seven other react rules already reached for it. It
// moved here when a core rule needed the same answer: `logical-assignment-operators` must stay silent
// inside compiled functions, because the compiler refuses the `||=` it would suggest, and a core rule
// cannot import a react rule package. A second detector written for that one rule would have drifted
// from this one the first time either was corrected.
//
// The two name predicates are `IsCompilerComponentName` and `IsCompilerHookName` rather than this
// shelf's `IsLikelyComponentName` and `IsHookName`, deliberately: the compiler's tests are ASCII
// and accept a digit after `use`, and each doc comment below records the measurement that showed
// the shelf's general predicates answer differently.

// IsInsideComponentOrHook reports whether a node sits inside a function React Compiler compiles.
//
// # The rule, which took seventeen probe rounds against React's own rule to state correctly
//
// Two conditions, and both are necessary. A function is a compilation root when it is
// component-or-hook-like on its own merits, **and** its position is reachable: every step from it
// out to the source file passes through either a component-or-hook-NAMED function or a construct
// that forms no scope of its own. Once a root is chosen its whole subtree is lowered, so anything
// nested inside it is judged whether or not the nested thing looks like a component.
//
// The second condition is the one that is invisible from `getComponentOrHookLike`, which knows
// nothing about where its argument sits, and it is where a port goes wrong. It comes from
// `findFunctionsToCompile`, which walks the program picking roots.
//
// The distinction between "like" and "NAMED" in that sentence is load-bearing and was measured
// rather than reasoned. Varying only the enclosing function's name over eight spellings, with the
// body held byte-identical:
//
//	function useOuter() { const Component = p => { eval('x'); return <div />; }; ... }   reports
//	function Outer()    { ... same body ... }                                            reports
//	function Xyz()      { ... same body ... }                                            reports
//	function use2()     { ... same body ... }                                            reports
//	function plainOuter() { ... same body ... }                                          silent
//	function outer()    { ... same body ... }                                            silent
//	function abc()      { ... same body ... }                                            silent
//	function use()      { ... same body ... }                                            silent
//
// The enclosing function has no evidence of its own in any of those, so it never qualifies as a
// root itself. Its NAME still decides whether the traversal reaches the inner arrow. That is why
// the position walk asks only the name question of ancestors while asking the full question of the
// candidate itself.
//
// # The measurements that pin the rest of it
//
// Position, holding shape fixed:
//
//	{ function Component(props) { eval('x'); return <div />; } }                 silent
//	if (true) { function Component(props) { eval('x'); return <div />; } }       silent
//	class K { m() { function Component(props) { eval('x'); ... } } }             silent
//	function Component(props) { eval('x'); return <div />; }                     reports
//
// Evidence, holding position and name fixed:
//
//	function Component(props) { eval('x'); const inner = () => <div />; ... }    silent
//	function Component(props) { eval('x'); return <div />; }                     reports
//
// The first of those is the nested-function skip in the evidence walk: JSX that only appears inside
// an inner closure does not count for the enclosing function. And once a root does qualify on its
// own evidence, a plain nested function inside it is still judged:
//
//	function Component(props) { const inner = () => { eval('x'); ... }; return <div>{inner()}</div>; }  reports
func IsInsideComponentOrHook(node *ast.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		if IsComponentOrHookLike(current) && IsReachableRootPosition(current) {
			return true
		}
	}
	return false
}

// IsReachableRootPosition reports whether the traversal picking compilation roots can reach a
// function-like node at all.
//
// Everything between the candidate and the source file must be either transparent, meaning it forms
// no scope and is just how the function is written down, or a component-or-hook-NAMED function,
// through which the traversal continues looking for roots.
//
// A name test rather than the full `IsComponentOrHookLike` on ancestors, for the reason measured at
// the caller: an enclosing function with a qualifying name and no evidence of its own still lets the
// traversal through, while the same body under a non-qualifying name does not.
//
// The transparent set is an allow list rather than a deny list on purpose. A deny list silently
// admits every node kind nobody thought of, and admitting one wrongly makes this rule report inside
// functions upstream never compiles, which is the expensive direction to be wrong in.
func IsReachableRootPosition(node *ast.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindSourceFile:
			return true
		case ast.KindVariableDeclaration, ast.KindVariableDeclarationList,
			ast.KindVariableStatement, ast.KindExportAssignment, ast.KindExportDeclaration,
			ast.KindParenthesizedExpression:
			// The spelling of a top-level binding, an export wrapper, and a redundant parenthesis
			// our parser keeps where Babel's tree drops it. None of these moves the function into a
			// nested scope, and omitting the parenthesis arm made `const Component = ((props) => ...)`
			// unreachable while React reports it.
			continue
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
			if !hasComponentOrHookName(current) {
				return false
			}
			continue
		case ast.KindBlock:
			// A function body block is transparent, because the function above it has already been
			// judged by the arm above. A bare block, an `if` body or a loop body is not, and both
			// spell as KindBlock, so the parent is what separates them.
			parent := current.Parent
			if parent == nil || !isSkippedNestedFunction(parent) {
				return false
			}
			continue
		default:
			return false
		}
	}
	return false
}

// SkipParenthesesUpward walks out through redundant parentheses to the first real parent.
//
// The mirror of `ast.SkipParentheses`, which walks inward. Our parser keeps a parenthesized
// expression as a live node, so a function written `((props) => {})` sits one level deeper than the
// same function written without them, and every ancestor test has to account for it.
func SkipParenthesesUpward(node *ast.Node) *ast.Node {
	current := node.Parent
	for current != nil && current.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	return current
}

// hasComponentOrHookName reports whether a function-like node carries a name that claims to be a
// component or a hook, without asking whether it has any evidence of being one.
//
// The name half of the gate on its own. See the table at `IsInsideComponentOrHook` for the eight
// spellings that establish this is genuinely the name and not the shape.
func hasComponentOrHookName(node *ast.Node) bool {
	name, named := inferredFunctionName(node)
	if !named {
		return false
	}
	return IsCompilerComponentName(name) || IsCompilerHookName(name)
}

// IsComponentOrHookLike answers React's `getComponentOrHookLike` for the two naming forms this port
// recognizes.
//
// The name is read from the declaration's own identifier or, for a function expression and an arrow,
// from the variable declarator that holds it. See the note on the rule for why the object-property,
// assignment, member-expression, `forwardRef` and `memo` paths are not recognized: all of them were
// measured silent under React's own rule on this rule's diagnostics.
func IsComponentOrHookLike(node *ast.Node) bool {
	name, named := inferredFunctionName(node)
	if !named {
		return false
	}
	if IsCompilerHookName(name) {
		return callsHooksOrCreatesJsx(node)
	}
	if !IsCompilerComponentName(name) {
		return false
	}
	return callsHooksOrCreatesJsx(node) && hasComponentShapedParameters(node)
}

// inferredFunctionName returns the name React would infer for a function-like node.
//
// A function declaration carries its own identifier. A function expression or an arrow takes the
// name of the variable declarator initializing it, which is the `parent.isVariableDeclarator() &&
// parent.get('init').node === path.node` branch of upstream's `getFunctionName`. The identity check
// on the initializer matters: without it a declarator whose initializer merely *contains* the
// function, such as `const Component = memo(props => ...)`, would hand the arrow the outer name,
// and that case is measured silent upstream.
func inferredFunctionName(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		name := node.AsFunctionDeclaration().Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return "", false
		}
		return name.Text(), true
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		// Parentheses are skipped on the way OUT to the declarator, and this direction was measured
		// rather than assumed. Our parser keeps `const Component = ((props) => {...});` as a real
		// KindParenthesizedExpression between the declaration and the arrow, where Babel's tree has
		// already dropped it, so an unskipped walk finds a parenthesized expression as the parent
		// and answers unnamed. Measured on React's own rule, that input **reports**, so withholding
		// the skip here would be a false negative rather than a harmless narrowing. It was the only
		// disagreement in a 111-input differential run against React and it is the port brief's
		// "measure which tree drops the parens" rule producing a live hit.
		parent := SkipParenthesesUpward(node)
		if parent == nil || parent.Kind != ast.KindVariableDeclaration {
			return "", false
		}
		declaration := parent.AsVariableDeclaration()
		if ast.SkipParentheses(declaration.Initializer) != node {
			return "", false
		}
		name := declaration.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return "", false
		}
		return name.Text(), true
	}
	return "", false
}

// IsCompilerComponentName answers React's `/^[A-Z]/`.
//
// Written here rather than taken from `react.IsLikelyComponentName`, and the reason is measured
// rather than stylistic. That shelf helper tests `unicode.IsUpper`, so `Émile` and `Ωmega` are
// component names to it. Probed against React's own rule, `function Émile(props) { eval('x');
// return <div />; }` is **silent**, because React's test is an ASCII regex. Building on the shelf
// helper would have reported an input upstream passes.
//
// The shelf's own doc comment already records that divergence and says nothing has established
// which answer we want. For this rule the answer is established: React is the authority and React
// is ASCII, so the ASCII test is what ships here.
func IsCompilerComponentName(name string) bool {
	if name == "" {
		return false
	}
	first := name[0]
	return first >= 'A' && first <= 'Z'
}

// IsCompilerHookName answers React's `/^use[A-Z0-9]/`.
//
// Also written here rather than taken from `react.IsHookName`, and also for a measured reason
// running the other way. The shelf helper rejects a digit after the prefix, so `use2Things` is not a
// hook to it. Probed against React's own rule, `function use2Things(a) { eval('x'); return <div />; }`
// **reports**, because React's character class is `[A-Z0-9]`. Building on the shelf helper would
// have gone silent on an input upstream reports.
//
// Both shelf name helpers are therefore wrong for this rule, in opposite directions, and both have
// doc comments accurately describing behavior that does not match this authority. That is the port
// brief's shelf-probe rule producing two hits in one file.
func IsCompilerHookName(name string) bool {
	if len(name) < 4 {
		return false
	}
	if name[0:3] != "use" {
		return false
	}
	fourth := name[3]
	return (fourth >= 'A' && fourth <= 'Z') || (fourth >= '0' && fourth <= '9')
}

// callsHooksOrCreatesJsx reports whether a function's own body writes JSX or calls a hook.
//
// Upstream's `callsHooksOrCreatesJsx` traverses the body and installs skip handlers on every nested
// function kind, so a nested closure's contents do not count toward the enclosing function's
// evidence. That skip is reproduced by not descending into a nested function-like node, and it is
// load-bearing rather than an optimization: measured on React's rule, a capitalized function whose
// only JSX lives inside a nested arrow is silent, and so is one whose only hook call does.
//
// The parameter list is not walked either, matching upstream, which traverses the body path.
func callsHooksOrCreatesJsx(node *ast.Node) bool {
	body := FunctionBody(node)
	if body == nil {
		return false
	}
	found := false
	var walk func(current *ast.Node)
	walk = func(current *ast.Node) {
		if current == nil || found {
			return
		}
		if isJsxNode(current) {
			found = true
			return
		}
		if current.Kind == ast.KindCallExpression && isCompilerHookCallee(current.AsCallExpression().Expression) {
			found = true
			return
		}
		if isSkippedNestedFunction(current) {
			return
		}
		current.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	// The body itself may be a function-like node's block; walking its children rather than the
	// body node avoids the skip test declining the very function being judged.
	body.ForEachChild(func(child *ast.Node) bool {
		walk(child)
		return found
	})
	// An arrow with an expression body is the one shape whose body is not a block, and its
	// expression is the thing to judge rather than something to descend past.
	if !found && body.Kind != ast.KindBlock {
		walk(body)
	}
	return found
}

// FunctionBody returns the body of a function-like node, or nil when it has none.
//
// `Node.Body()` is an accessor and the port brief's standing hazard is that every `Node.Xxx()`
// accessor panics off its kind, so the kind is established by the switch before the typed accessor
// is reached rather than by trusting a general helper. A nil node answers nil, which is what the
// constant-casing rule's copy did before it was folded into this one.
func FunctionBody(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Body
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Body
	case ast.KindArrowFunction:
		return node.AsArrowFunction().Body
	}
	return nil
}

// isSkippedNestedFunction reports the node kinds upstream installs skip handlers for.
//
// Exactly three, matching upstream's `ArrowFunctionExpression`, `FunctionExpression` and
// `FunctionDeclaration` skip handlers. A method or an accessor is deliberately absent: upstream
// does not skip them either, and a class body inside a component is already the inline-class
// finding rather than something to search for evidence in.
//
// Not named `isFunctionLike`, which `error_boundaries.go` in this package already defines for a
// different question. Its set is the one that answers "is this a function", while this one is the
// narrower set upstream declines to descend through, and collapsing them would silently widen or
// narrow one caller.
func isSkippedNestedFunction(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		return true
	}
	return false
}

// isJsxNode reports whether a node is any JSX construct.
//
// Upstream's traversal registers the `JSX` alias, which Babel expands to every JSX node type, so
// the test here is over the kinds our parser produces for the same source rather than over element
// kinds alone. A fragment is included because `return <></>;` was measured as sufficient evidence
// under React's rule.
func isJsxNode(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment,
		ast.KindJsxOpeningElement, ast.KindJsxOpeningFragment, ast.KindJsxExpression,
		ast.KindJsxText, ast.KindJsxAttribute, ast.KindJsxAttributes, ast.KindJsxSpreadAttribute:
		return true
	}
	return false
}

// isCompilerHookCallee reports whether a call's callee reads as a hook to the React Compiler.
//
// Upstream's `isHook` accepts a bare identifier with a hook name and a non-computed member
// expression whose property is a hook name and whose object is a capitalized namespace. Both are
// reproduced; the namespace test is what makes `React.useState(...)` count while `obj.useThing(...)`
// on a lowercase receiver does not.
//
// # Why this does not reuse `isHookCallee`, which is in this package and answers the same shape
//
// `rules_of_hooks.go` defines `isHookCallee` with an identical structure, and reusing it was the
// first instinct. Its name predicate is `isHookIdentifierName`, whose own doc comment records a
// deliberate divergence: it accepts **bare `use`**, because both of that rule's upstreams treat
// `use(...)` as a hook and its corpus turns on the case.
//
// The React Compiler does not. Its `isHookName` is `/^use[A-Z0-9]/`, which bare `use` fails.
// Measured on React's own rule rather than argued from the regex: a capitalized function whose only
// evidence is a `use(props.p)` call is **silent**, and so is `function use(a) { eval('x'); return
// <div />; }`, while both `use42` spellings **report**. Routing this rule through the neighbouring
// helper would have reported two inputs React passes.
//
// So the two live side by side, each correct for its own authority, and this is the port brief's
// "read the body of anything whose name matches your question" warning producing a hit inside a
// single package rather than out on a shelf.
func isCompilerHookCallee(callee *ast.Node) bool {
	if callee == nil {
		return false
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		return IsCompilerHookName(callee.Text())
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier || !IsCompilerHookName(name.Text()) {
			return false
		}
		object := access.Expression
		return object != nil && object.Kind == ast.KindIdentifier &&
			IsCompilerComponentName(object.Text())
	}
	return false
}

// hasComponentShapedParameters answers React's `isValidComponentParams`.
//
// Three rules, all measured on React's own rule rather than read off the source alone. No
// parameters is valid. One or two are valid when the first is not a rest element and is not
// annotated as a primitive type. A second parameter must be an identifier whose name contains
// `ref` or `Ref`, which is upstream's own `name.includes('ref') || name.includes('Ref')`. Three or
// more is never valid.
//
// This applies to components only. Upstream asks it inside the `isComponentName` branch and not
// inside the hook branch, so a hook with any parameter shape still qualifies, which is why
// `function use2Things(a)` reports.
func hasComponentShapedParameters(node *ast.Node) bool {
	parameters := FunctionParameters(node)
	switch len(parameters) {
	case 0:
		return true
	case 1:
		first := parameters[0]
		return !isSpreadParameter(first) && !hasPrimitiveTypeAnnotation(first)
	case 2:
		first := parameters[0]
		if isSpreadParameter(first) || hasPrimitiveTypeAnnotation(first) {
			return false
		}
		second := parameters[1].AsParameterDeclaration().Name()
		if second == nil || second.Kind != ast.KindIdentifier {
			return false
		}
		name := second.Text()
		return containsRef(name)
	}
	return false
}

// FunctionParameters returns a function-like node's parameters.
//
// `Node.ParameterList()` panics off function-like kinds, which the port brief names, so the kind is
// established here before any typed accessor is reached.
func FunctionParameters(node *ast.Node) []*ast.Node {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Parameters.Nodes
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Parameters.Nodes
	case ast.KindArrowFunction:
		return node.AsArrowFunction().Parameters.Nodes
	}
	return nil
}

// isSpreadParameter reports whether a parameter is written with a spread.
//
// Not named `isRestParameter`, which `error_boundaries.go` in this package already defines.
func isSpreadParameter(parameter *ast.Node) bool {
	return parameter.AsParameterDeclaration().DotDotDotToken != nil
}

// containsRef answers upstream's `name.includes('ref') || name.includes('Ref')`.
//
// Written as two substring tests rather than as a case-insensitive comparison, because that is what
// upstream does and the two differ: a parameter named `REF` contains neither spelling and is
// therefore not a valid second parameter upstream, while a case-insensitive test would accept it.
func containsRef(name string) bool {
	return containsSubstring(name, "ref") || containsSubstring(name, "Ref")
}

// containsSubstring is `strings.Contains`, spelled out to keep this file's imports to the two
// packages the rule genuinely needs.
func containsSubstring(haystack string, needle string) bool {
	if len(needle) > len(haystack) {
		return false
	}
	for index := 0; index+len(needle) <= len(haystack); index++ {
		if haystack[index:index+len(needle)] == needle {
			return true
		}
	}
	return false
}

// hasPrimitiveTypeAnnotation answers upstream's `isValidPropsAnnotation` inverted.
//
// Upstream lists the TypeScript annotation kinds that disqualify a first parameter from being
// props, and everything else, including no annotation at all, is valid. The list is reproduced
// rather than approximated as "not an object type", because upstream's own set admits shapes an
// approximation would decline: a type reference, an intersection, a union and a mapped type are all
// valid props annotations there.
//
// Measured both ways on React's rule: `function Component(props: string)` is silent and
// `function Component(props: {a: number})` reports.
func hasPrimitiveTypeAnnotation(parameter *ast.Node) bool {
	annotation := parameter.AsParameterDeclaration().Type
	if annotation == nil {
		return false
	}
	switch annotation.Kind {
	case ast.KindArrayType, ast.KindBigIntKeyword, ast.KindBooleanKeyword,
		ast.KindConstructorType, ast.KindFunctionType, ast.KindLiteralType,
		ast.KindNeverKeyword, ast.KindNumberKeyword, ast.KindStringKeyword,
		ast.KindSymbolKeyword, ast.KindTupleType:
		return true
	}
	return false
}
