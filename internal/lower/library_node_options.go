package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// nodeHostConsumesArgument proves both consumption and absence. This is not a
// general exemption for Node declarations or for readonly arguments. The fs-file
// lowerer replaces these option objects with scalar arguments; it neither stores
// nor returns the object, nor passes it to the runtime as a wider object view.
func (l *lowering) nodeHostConsumesArgument(node *ast.Node, index int) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	if !l.optionalNodeHostCall(node) {
		return false
	}
	call := node.AsCallExpression()
	if index < 0 || index >= len(call.Arguments.Nodes) {
		return false
	}
	optionIndex := 1
	switch l.nodeLibraryMember(node) {
	case "node:fs.statSync", "node:fs.mkdirSync", "node:fs.readFileSync", "node:fs.rmSync", "node:fs.mkdtempSync":
	case "node:fs.writeFileSync":
		optionIndex = 2
	default:
		// In particular, readdirSync still passes its options object onward.
		return false
	}
	if index != optionIndex {
		return false
	}
	argument := ast.SkipParentheses(call.Arguments.Nodes[index])
	if !l.nodeHostExactOptions(argument, node) {
		return false
	}
	// The consumed fs options are scalar. Do not let an exact outer literal hide
	// a nested optional widening through an object-valued field or a callback.
	for _, field := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(argument)) {
		if !nodeHostScalarOption(l.checker.GetTypeOfSymbol(field)) {
			return false
		}
	}
	return true
}

func nodeHostScalarOption(proven *checker.Type) bool {
	if proven == nil {
		return false
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if !nodeHostScalarOption(member) {
				return false
			}
		}
		return true
	}
	return proven.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0
}

// nodeHostExactOptions proves a fresh, unescaped shape at this call. A private
// module binding may be captured only by this consuming argument: all other
// object uses are rejected regardless of text order, since a function can run
// later than its declaration. Other captures remain refused. In
// particular, a structural cast is not a const assertion and cannot hide fields.
func (l *lowering) nodeHostExactOptions(argument, call *ast.Node) bool {
	unwrap := func(node *ast.Node) *ast.Node {
		for {
			node = ast.SkipParentheses(node)
			if node.Kind == ast.KindAsExpression && ast.IsConstAssertion(node) {
				node = node.AsAsExpression().Expression
				continue
			}
			return node
		}
	}
	argument = unwrap(argument)
	if argument.Kind == ast.KindObjectLiteralExpression {
		return l.exactObject(argument, 0)
	}
	if !ast.IsIdentifier(argument) {
		return false
	}
	symbol := l.symbol(argument)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	binding := declaration.AsVariableDeclaration()
	if binding.Type != nil || binding.Initializer == nil || !ast.IsIdentifier(binding.Name()) {
		return false
	}
	literal := unwrap(binding.Initializer)
	if literal.Kind != ast.KindObjectLiteralExpression || !l.exactObject(literal, 0) {
		return false
	}
	owner := func(node *ast.Node) *ast.Node {
		for node != nil && !ast.IsFunctionLike(node) && node.Kind != ast.KindSourceFile {
			node = node.Parent
		}
		return node
	}
	captured := owner(declaration) != owner(call)
	if (captured && owner(declaration).Kind != ast.KindSourceFile) || declaration.End() > call.Pos() {
		return false
	}
	// Export exposes the object to callers whose writes cannot be inspected.
	if ast.HasSyntacticModifier(declaration.Parent.Parent, ast.ModifierFlagsExport) {
		return false
	}
	// Reusing the same exact options in another identical consuming host call
	// does not expose the object. This includes the directory fixture's module
	// const used both by a private helper and by direct stat calls.
	consumingUse := func(node *ast.Node) bool {
		use := node
		for use.Parent != nil && use.Parent.Kind == ast.KindParenthesizedExpression {
			use = use.Parent
		}
		parent := use.Parent
		if parent == nil || parent.Kind != ast.KindCallExpression || l.nodeLibraryMember(parent) != l.nodeLibraryMember(call) {
			return false
		}
		index := 1
		if l.nodeLibraryMember(call) == "node:fs.writeFileSync" {
			index = 2
		}
		arguments := parent.AsCallExpression().Arguments.Nodes
		return index < len(arguments) && arguments[index] == use
	}
	valid := true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && node != binding.Name() && l.symbol(node) == symbol {
			if node != argument && consumingUse(node) {
				// The object is synchronously unpacked at this other proven host use.
			} else if node != argument && owner(node) != owner(declaration) {
				valid = false // Captures other than this consuming argument.
			} else if node != argument {
				use := node
				for use.Parent != nil && use.Parent.Kind == ast.KindParenthesizedExpression {
					use = use.Parent
				}
				parent := use.Parent
				if parent != nil && (parent.Kind == ast.KindPropertyAccessExpression || parent.Kind == ast.KindElementAccessExpression) && parent.Expression() == use {
					use = parent
					if ast.IsAssignmentTarget(use) {
						valid = false
					}
					for use.Parent != nil && use.Parent.Kind == ast.KindParenthesizedExpression {
						use = use.Parent
					}
					for parent = use.Parent; parent != nil && !ast.IsFunctionLike(parent); parent = parent.Parent {
						if parent.Kind == ast.KindDeleteExpression || (parent.Kind == ast.KindPrefixUnaryExpression && (parent.AsPrefixUnaryExpression().Operator == ast.KindPlusPlusToken || parent.AsPrefixUnaryExpression().Operator == ast.KindMinusMinusToken)) || parent.Kind == ast.KindPostfixUnaryExpression {
							valid = false
						}
						if parent.Kind == ast.KindBinaryExpression && ast.IsAssignmentOperator(parent.AsBinaryExpression().OperatorToken.Kind) {
							left := parent.AsBinaryExpression().Left
							if left.Pos() <= use.Pos() && left.End() >= use.End() {
								valid = false
							}
						}
						if parent.Kind == ast.KindCallExpression && parent.AsCallExpression().Expression.Pos() <= use.Pos() && parent.AsCallExpression().Expression.End() >= use.End() {
							valid = false // Calling a method exposes its receiver.
						}
					}
				} else if captured || node.Pos() < call.End() {
					valid = false // Aliases, casts, storage and prior calls escape.
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	ast.GetSourceFileOfNode(declaration).AsNode().ForEachChild(visit)
	return valid
}
