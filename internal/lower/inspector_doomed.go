package lower

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Only a const whose initializer cannot complete has no value at every read,
// closures included. Early reads still use the ordinary TDZ/refusal path.
func (l *lowering) inspectorDoomedBinding(node *ast.Node, seen map[*ast.Symbol]bool) *ast.Node {
	node = ast.SkipOuterExpressions(node, ast.OEKAll)
	if !ast.IsIdentifier(node) {
		return nil
	}
	symbol := l.symbol(node)
	if symbol == nil || seen[symbol] {
		return nil
	}
	seen[symbol] = true
	defer delete(seen, symbol)
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Kind != ast.KindVariableDeclarationList || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
			continue
		}
		initializer := declaration.Initializer()
		if initializer == nil {
			continue
		}
		initializer = ast.SkipOuterExpressions(initializer, ast.OEKAll)
		if initializer.Kind == ast.KindCallExpression && l.nodeLibraryMember(initializer) == nodeRequireMember && nodeInspectorRequireThrows(initializer) {
			return l.nodeCommonJSRequireRoot(initializer)
		}
		switch initializer.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression, ast.KindNewExpression:
			if root := inspectorReceiverRoot(initializer); root != nil {
				return l.inspectorDoomedBinding(root, seen)
			}
		}
	}
	return nil
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
	if node.Parent != nil && node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.Name() == node {
		node = node.Parent
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
	require = l.inspectorDoomedBinding(root, map[*ast.Symbol]bool{})
	if require == nil {
		return nil, nil
	}
	return root, require
}

// Every admitted expression goes through this constructor. Its ordinary read
// may refuse, or throw the caught TDZ ReferenceError, before the terminal Panic.
func (l *lowering) inspectorDoomedValue(node *ast.Node) (ir.Expression, bool, error) {
	root, require := l.inspectorDoomedUse(node)
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
	message := l.nodeLibraryMember(node) + " was reached through " + root.Text() + ", whose require(\"" + nodeCommonJSRequireArgument(require).Text() + "\") throws on every host Adamic targets"
	// A fresh string slot identifies this actual construction, not just its text.
	index := len(l.result.Strings)
	l.result.Strings = append(l.result.Strings, message)
	value := ir.Effects{Body: []ir.Statement{ir.Evaluate{Value: read}, ir.Panic{Message: ir.StringConstant{Index: index}}}, Result: unreachablePlaceholder(of)}
	// The trap dominates the entire expression, including callbacks in arguments
	// which are never constructed. Separately declared callbacks are not children.
	var cover ast.Visitor
	cover = func(child *ast.Node) bool {
		if _, skipped := l.inspectorSkippedContracts[child]; skipped {
			l.inspectorSkippedContracts[child] = index + 1
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
