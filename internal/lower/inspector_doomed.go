package lower

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// A const roots at the throwing require. A function-local let also qualifies
// when every write is either rooted there or typed undefined or a string literal.
// The second result records let provenance through const aliases as well.
func (l *lowering) inspectorDoomedBinding(node *ast.Node, seen map[*ast.Symbol]bool) (*ast.Node, bool) {
	node = ast.SkipOuterExpressions(node, ast.OEKAll)
	if !ast.IsIdentifier(node) {
		return nil, false
	}
	symbol := l.symbol(node)
	if symbol == nil || seen[symbol] || len(symbol.Declarations) != 1 {
		return nil, false
	}
	seen[symbol] = true
	defer delete(seen, symbol)
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || !ast.IsIdentifier(declaration.Name()) || declaration.Parent == nil || declaration.Parent.Kind != ast.KindVariableDeclarationList {
		return nil, false
	}
	list := declaration.Parent
	initializer := declaration.Initializer()
	if list.Flags&ast.NodeFlagsConst != 0 {
		if initializer == nil {
			return nil, false
		}
		initializer = ast.SkipOuterExpressions(initializer, ast.OEKAll)
		if initializer.Kind == ast.KindCallExpression && l.nodeLibraryMember(initializer) == nodeRequireMember && nodeInspectorRequireThrows(initializer) {
			return l.nodeCommonJSRequireRoot(initializer), false
		}
		if root := inspectorReceiverRoot(initializer); root != nil {
			return l.inspectorDoomedBinding(root, seen)
		}
		return nil, false
	}
	function := ast.GetContainingFunction(declaration)
	if list.Flags&ast.NodeFlagsLet == 0 || ast.IsAssignmentTarget(list) || function == nil || function.Body() == nil || ast.HasSyntacticModifier(list.Parent, ast.ModifierFlagsExport) {
		return nil, false
	}
	// A bare declaration writes undefined only when its declared type includes it.
	if initializer == nil && !inspectorDeclaredUndefined(l.checker.GetTypeOfSymbol(symbol)) {
		return nil, false
	}
	var require *ast.Node
	classify := func(value *ast.Node) bool {
		if root := inspectorReceiverRoot(value); root != nil {
			if doomed, _ := l.inspectorDoomedBinding(root, seen); doomed != nil {
				require = doomed
				return true
			}
		}
		return inspectorOutsideArm(l.checker.GetTypeAtLocation(value))
	}
	if initializer != nil && !classify(initializer) {
		return nil, false
	}
	valid := true
	var visit ast.Visitor
	visit = func(reference *ast.Node) bool {
		if !valid {
			return true
		}
		if reference.Kind == ast.KindIdentifier && reference != declaration.Name() && ast.IsAssignmentTarget(reference) {
			targetSymbol := l.symbol(reference)
			// Shorthand names have property symbols; the assignment writes their value
			// symbol, as in the existing realpath reference walk.
			if reference.Parent != nil && ast.IsShorthandPropertyAssignment(reference.Parent) {
				targetSymbol = l.checker.GetShorthandAssignmentValueSymbol(reference.Parent)
			}
			if targetSymbol != symbol {
				return reference.ForEachChild(visit)
			}
			// Count references, including nested functions and generators. Only a plain
			// identifier on the left of = supplies a right side we can classify.
			target := reference
			for target.Parent != nil && target.Parent.Kind == ast.KindParenthesizedExpression {
				target = target.Parent
			}
			assignment := target.Parent
			if assignment == nil || assignment.Kind != ast.KindBinaryExpression || assignment.AsBinaryExpression().Left != target || assignment.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken || ast.IsAssignmentTarget(assignment) {
				valid = false
			} else {
				valid = classify(assignment.AsBinaryExpression().Right)
			}
		}
		return !valid || reference.ForEachChild(visit)
	}
	function.Body().ForEachChild(visit)
	if !valid {
		return nil, false
	}
	return require, true
}

func inspectorDeclaredUndefined(of *checker.Type) bool {
	if of.Flags()&checker.TypeFlagsUndefined != 0 {
		return true
	}
	if of.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range of.Types() {
			if inspectorDeclaredUndefined(member) {
				return true
			}
		}
	}
	return false
}

// Exactly undefined and string literal types are outside this inspector arm.
// Unions count only when every member is one of those types. Null, void, never,
// broad string, any, unknown and all object types are deliberately unclassified.
func inspectorOutsideArm(of *checker.Type) bool {
	if of.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range of.Types() {
			if !inspectorOutsideArm(member) {
				return false
			}
		}
		return len(of.Types()) != 0
	}
	return of.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsStringLiteral) != 0
}

func inspectorReceiverRoot(node *ast.Node) *ast.Node {
	for node != nil {
		node = ast.SkipOuterExpressions(node, ast.OEKAll)
		switch node.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression, ast.KindNewExpression:
			node = node.Expression()
		default:
			if ast.IsIdentifier(node) {
				return node
			}
			return nil
		}
	}
	return nil
}

// One admission predicate, shared by rule 800, merged-field contracts, and the
// trap builder. Registered profile data and every other library stay unchanged.
func (l *lowering) inspectorDoomedUse(node *ast.Node) (root, require *ast.Node) {
	node = inspectorContractNode(node)
	if node == nil {
		return nil, nil
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression, ast.KindNewExpression:
	default:
		return nil, nil
	}
	if !strings.HasPrefix(l.nodeLibraryMember(node), "node:inspector.") || nodeLibraryMemberImplemented(l.nodeLibrarySymbol(node)) {
		return nil, nil
	}
	root = inspectorReceiverRoot(node)
	if root == nil {
		return nil, nil
	}
	var viaLet bool
	require, viaLet = l.inspectorDoomedBinding(root, map[*ast.Symbol]bool{})
	if require == nil {
		return nil, nil
	}
	if viaLet {
		receiver := node
		for {
			core := ast.SkipOuterExpressions(receiver, ast.OEKAll)
			if ast.IsOptionalChain(core) {
				return nil, nil
			}
			switch core.Kind {
			case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression, ast.KindNewExpression:
				receiver = core.Expression()
			default:
				// The checker must have narrowed the receiver to its inspector object arm.
				// Keep casts and non-null assertions in the ordinary read, so their own
				// checks run before the trap rather than reading through an erased hatch.
				if l.checker.GetTypeAtLocation(receiver).Flags()&checker.TypeFlagsObject == 0 {
					return nil, nil
				}
				if inspectorReceiverRoot(receiver) == root {
					root = receiver
				}
				return root, require
			}
		}
	}
	return root, require
}

// A refusal can name a property access or its name. Both identify the same
// member operation. Calls/new stay themselves; the trap cover walk visits the
// callee property beneath them, so every recorded key is an exact AST descendant.
func inspectorContractNode(node *ast.Node) *ast.Node {
	if node != nil && node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.Name() == node {
		return node.Parent
	}
	return node
}

// Scans admit syntax, including deferred generic bodies that may never be built.
// Only a lowering entry records an obligation. The same helper is used by the
// trap and by direct member/call/builtin roads that could bypass it.
func (l *lowering) inspectorRecordBuiltUse(node *ast.Node) (root, require *ast.Node) {
	node = inspectorContractNode(node)
	root, require = l.inspectorDoomedUse(node)
	if root == nil {
		return nil, nil
	}
	if l.inspectorSkippedContracts == nil {
		l.inspectorSkippedContracts = map[*ast.Node]int{}
	}
	if _, known := l.inspectorSkippedContracts[node]; !known {
		l.inspectorSkippedContracts[node] = 0
	}
	return root, require
}

// Every admitted expression goes through this constructor. Its ordinary read
// may refuse, or throw the caught TDZ ReferenceError, before the terminal Panic.
func (l *lowering) inspectorDoomedValue(node *ast.Node) (ir.Expression, bool, error) {
	node = inspectorContractNode(node)
	root, require := l.inspectorRecordBuiltUse(node)
	if root == nil {
		return nil, false, nil
	}
	read, err := l.expression(root)
	if err != nil {
		return nil, true, err
	}
	of := ir.Object
	if held, known := l.representation(l.checker.GetTypeAtLocation(node)); known {
		of = held
	}
	// The read retains hatches; the diagnostic names only the binding identifier.
	name := inspectorReceiverRoot(root)
	message := l.nodeLibraryMember(node) + " was reached through " + name.Text() + ", whose require(\"" + nodeCommonJSRequireArgument(require).Text() + "\") throws on every host Adamic targets"
	// A fresh string slot identifies this actual construction, not just its text.
	index := len(l.result.Strings)
	l.result.Strings = append(l.result.Strings, message)
	value := ir.Effects{Body: []ir.Statement{ir.Evaluate{Value: read}, ir.Panic{Message: ir.StringConstant{Index: index}}}, Result: unreachablePlaceholder(of)}
	// The trap dominates the entire expression, including callbacks in arguments
	// which are never constructed. Separately declared callbacks are not children.
	var cover ast.Visitor
	cover = func(child *ast.Node) bool {
		if key := inspectorContractNode(child); key != nil {
			if _, skipped := l.inspectorSkippedContracts[key]; skipped {
				l.inspectorSkippedContracts[key] = index + 1
			}
		}
		return child.ForEachChild(cover)
	}
	cover(node)
	return value, true, nil
}

// Fail closed unless every skipped contract has its trap in the actual final
// IR. The marker map is owned by the paired lowering snapshot, not carried as a
// program fact into a different IR. Falling through cannot admit an ordinary call.
func (l *lowering) checkInspectorDoomedContracts() error {
	if len(l.inspectorSkippedContracts) == 0 {
		return nil
	}
	present := map[int]bool{}
	scan := func(node any) bool {
		if panic, ok := node.(ir.Panic); ok {
			if message, ok := panic.Message.(ir.StringConstant); ok {
				present[message.Index+1] = true
			}
		}
		return true
	}
	walk(l.result.Main, scan)
	for _, function := range l.result.Functions {
		walk(function.Body, scan)
	}
	for node, marker := range l.inspectorSkippedContracts {
		if marker == 0 || !present[marker] {
			return fmt.Errorf("compiler bug: skipped inspector contract without its doomed-use Panic at %s", l.program.Where(node))
		}
	}
	return nil
}
