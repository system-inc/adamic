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
	if !l.nodeHostConsumesOptionsPosition(node, index) {
		return false
	}
	call := node.AsCallExpression()
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

// Separate the lowerer's consuming position from the binding proof so every
// reference can be checked without recursively proving the same binding.
func (l *lowering) nodeHostConsumesOptionsPosition(node *ast.Node, index int) bool {
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
	return index == optionIndex
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

// nodeHostExactOptions proves a fresh shape whose every reference is a scalar
// consuming host argument. Captures obey the same rule as direct references.
// Exports and all non-consuming references fail, regardless of execution order.
// Structural casts are not const assertions and cannot hide fields.
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
	// Export exposes the object to callers whose writes cannot be inspected.
	if ast.HasSyntacticModifier(declaration.Parent.Parent, ast.ModifierFlagsExport) {
		return false
	}
	valid := true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && node != binding.Name() {
			reference := l.symbol(node)
			if node.Parent != nil && ast.IsShorthandPropertyAssignment(node.Parent) {
				reference = l.checker.GetShorthandAssignmentValueSymbol(node.Parent)
			}
			if reference == symbol {
				use := node
				for use.Parent != nil && use.Parent.Kind == ast.KindParenthesizedExpression {
					use = use.Parent
				}
				consumed := false
				if parent := use.Parent; parent != nil && parent.Kind == ast.KindCallExpression {
					for index, argument := range parent.AsCallExpression().Arguments.Nodes {
						if argument == use && l.nodeHostConsumesOptionsPosition(parent, index) {
							consumed = true
							for _, field := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(argument)) {
								if !nodeHostScalarOption(l.checker.GetTypeOfSymbol(field)) {
									consumed = false
								}
							}
						}
					}
				}
				if !consumed {
					valid = false
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	if l.program == nil {
		return false // No complete reference inventory is available.
	}
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		if !file.IsDeclarationFile {
			file.AsNode().ForEachChild(visit)
		}
	}
	return valid
}
