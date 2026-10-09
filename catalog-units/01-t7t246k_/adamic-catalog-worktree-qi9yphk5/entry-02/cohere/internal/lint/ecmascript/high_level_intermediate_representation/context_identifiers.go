// Which bindings a function shares with the closures inside it, decided before anything is lowered.
//
// React's `FindContextIdentifiers`, which runs over the AST before `BuildHIR` and produces the set
// its lowering consults when it decides whether a write is a `StoreLocal` or a `StoreContext`.
//
// # Why the question cannot be asked during lowering
//
// A write lowers as `StoreContext` here when the WRITER is a nested function reaching outward, which
// is decidable at that point because `captureOf` has already seen the binding come in from an
// enclosing scope. Upstream asks a different question: is this binding reassigned anywhere, and does
// any inner function reference it? That is a property of the whole function, and at the moment an
// outer write lowers, the closure that captures it has not been walked yet -- `builder.captured`
// fills as the nested function lowers, which is after.
//
// So the outer reassignment in
//
//	let x = []; const cb = useCallback(() => [x], [x]); x = makeArray();
//
// lowered as an ordinary local store, upstream's lowers as a context store, and their
// `StoreContext` arm emits a `Mutate` effect that widens `x`'s mutable range past the memo. Without
// that widening the scope holding `x` closes before the marker, and
// `error.invalid-useCallback-captures-reassigned-context.ts` stays silent where upstream reports
// "This dependency may be mutated later, which could cause the value to change unexpectedly."
//
// # The rule, which is upstream's exactly
//
//	reassignedByInnerFn                     -> context
//	reassigned AND referencedByInnerFn      -> context
//
// `FindContextIdentifiers.ts:106`. The first case this tree already handles by the writer's
// position; the second is what this pass adds.
package high_level_intermediate_representation

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// contextIdentifiers is the set of symbols a function shares with the closures inside it.
type contextIdentifiers map[*ast.Symbol]bool

// identifierUsage is what the walk records about one binding before the rule is applied.
type identifierUsage struct {
	// reassigned is whether anything writes to the binding after its declaration.
	reassigned bool
	// referencedByInnerFunction is whether a function nested inside the subject reads it.
	referencedByInnerFunction bool
	// reassignedByInnerFunction is whether a function nested inside the subject writes it.
	reassignedByInnerFunction bool
}

// findContextIdentifiers returns the bindings that must lower as context rather than as locals.
//
// `node` is the function being lowered and `checker` resolves names to symbols, so a shadowed name
// is never confused with the binding it shadows -- the same reason `lowerExpression` resolves by
// symbol rather than by a scope tree this package would have to maintain.
func findContextIdentifiers(node *ast.Node, checker *checker.Checker) contextIdentifiers {
	context := contextIdentifiers{}
	if node == nil || checker == nil {
		return context
	}
	body := functionBody(node)
	if body == nil {
		return context
	}

	usage := map[*ast.Symbol]*identifierUsage{}
	record := func(symbol *ast.Symbol) *identifierUsage {
		if symbol == nil {
			return nil
		}
		if existing, found := usage[symbol]; found {
			return existing
		}
		fresh := &identifierUsage{}
		usage[symbol] = fresh
		return fresh
	}

	// `depth` counts function boundaries crossed from the subject, so zero is the subject's own body
	// and anything greater is a closure inside it.
	var walk func(current *ast.Node, depth int)
	walk = func(current *ast.Node, depth int) {
		if current == nil {
			return
		}
		if written := assignmentTarget(current); written != nil {
			if entry := record(checker.GetSymbolAtLocation(written)); entry != nil {
				entry.reassigned = true
				if depth > 0 {
					entry.reassignedByInnerFunction = true
				}
			}
		}
		if depth > 0 && ast.IsIdentifier(current) {
			if entry := record(checker.GetSymbolAtLocation(current)); entry != nil {
				entry.referencedByInnerFunction = true
			}
		}
		next := depth
		if ast.IsFunctionLike(current) && current != node {
			next++
		}
		current.ForEachChild(func(child *ast.Node) bool {
			walk(child, next)
			return false
		})
	}
	walk(body, 0)

	// # A mutation dropping the `referencedByInnerFunction` half SURVIVES
	//
	// Measured: accepting every reassigned binding scores identically on all five instruments --
	// goldens 26, false positives 31, scope `under` 5 fixtures / 7 scopes, `exact` 26. No corpus
	// fixture reassigns a binding that no closure reads, so the two rules select the same set here.
	//
	// Kept as upstream's because the direction matters and this one is not symmetric. Marking a
	// binding contextual makes SSA define it once, which suppresses the versioning that every later
	// pass reads -- so a binding no closure captures would lose its phis for no reason, and the
	// analysis below it would see one value where the program has several. The verdict EXPIRES the
	// moment a fixture reassigns a purely local binding, which any loop counter would produce.
	for symbol, entry := range usage {
		if entry.reassignedByInnerFunction ||
			(entry.reassigned && entry.referencedByInnerFunction) {
			context[symbol] = true
		}
	}
	return context
}

// assignmentTarget returns the identifier a node writes to, or nil when it writes to none.
//
// Only the shapes that name a binding directly. A write through a property or an element is a
// mutation of the object rather than a reassignment of the binding, which is the distinction
// upstream's `isAssignmentTarget` draws and the one that decides whether the range must widen.
func assignmentTarget(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindBinaryExpression:
		expression := node.AsBinaryExpression()
		if expression == nil || !ast.IsAssignmentOperator(expression.OperatorToken.Kind) {
			return nil
		}
		if expression.Left != nil && ast.IsIdentifier(expression.Left) {
			return expression.Left
		}
	case ast.KindPrefixUnaryExpression:
		expression := node.AsPrefixUnaryExpression()
		if expression == nil {
			return nil
		}
		if expression.Operator != ast.KindPlusPlusToken &&
			expression.Operator != ast.KindMinusMinusToken {
			return nil
		}
		if expression.Operand != nil && ast.IsIdentifier(expression.Operand) {
			return expression.Operand
		}
	case ast.KindPostfixUnaryExpression:
		expression := node.AsPostfixUnaryExpression()
		if expression == nil {
			return nil
		}
		if expression.Operand != nil && ast.IsIdentifier(expression.Operand) {
			return expression.Operand
		}
	}
	return nil
}
