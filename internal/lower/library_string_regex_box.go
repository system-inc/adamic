package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// StringData is private storage, never an observable own property. The complete use proof below
// admits only intrinsic search/conversion calls and typeof. Escaping, reflection,
// writes and structural String views stay refused before the object can be emitted.
const libraryStringData = "\x01StringData"

func (l *lowering) libraryStringBoxOrigin(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindNewExpression && l.isLibraryGlobal(node.AsNewExpression().Expression, "String") {
		return node
	}
	if ast.IsIdentifier(node) {
		symbol := l.checker.GetSymbolAtLocation(node)
		if symbol != nil && len(symbol.Declarations) == 1 && symbol.Declarations[0].Kind == ast.KindVariableDeclaration {
			initializer := symbol.Declarations[0].AsVariableDeclaration().Initializer
			if initializer != nil {
				initializer = ast.SkipParentheses(initializer)
				if initializer.Kind == ast.KindNewExpression && l.isLibraryGlobal(initializer.AsNewExpression().Expression, "String") {
					return initializer
				}
			}
		}
	}
	return nil
}

func libraryStringBoxMethod(name string) bool {
	switch name {
	case "search", "toString", "valueOf":
		return true
	}
	return false
}

func stringBoxOuter(node *ast.Node) *ast.Node {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node
}

func (l *lowering) libraryStringBoxUse(node *ast.Node) bool {
	node = stringBoxOuter(node)
	parent := node.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindTypeOfExpression {
		return true
	}
	if parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == node {
		return libraryStringBoxMethod(parent.Name().Text()) && called(parent)
	}
	if parent.Kind == ast.KindCallExpression {
		call := parent.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if len(call.Arguments.Nodes) == 1 && call.Arguments.Nodes[0] == node && l.isLibraryGlobal(callee, "String") {
			return true
		}
		if len(call.Arguments.Nodes) > 0 && call.Arguments.Nodes[0] == node && callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "call" {
			name, intrinsic := l.stringPrototypeMethod(callee.AsPropertyAccessExpression().Expression)
			return intrinsic && libraryStringBoxMethod(name)
		}
	}
	return false
}

func stringBoxWritten(node *ast.Node) bool {
	node = stringBoxOuter(node)
	if node.Parent == nil {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindBinaryExpression:
		binary := node.Parent.AsBinaryExpression()
		return binary.Left == node && ast.IsAssignmentOperator(binary.OperatorToken.Kind)
	case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression, ast.KindDeleteExpression:
		return true
	}
	return false
}

func (l *lowering) newLibraryStringBox(node *ast.Node) (ir.Expression, error) {
	refusal := func() (ir.Expression, error) {
		return nil, l.notYet(node, "new String internal slots with escaping, reassignment, reflection or unsupported observations")
	}
	outer := stringBoxOuter(node)
	if outer.Parent != nil && outer.Parent.Kind == ast.KindVariableDeclaration && outer.Parent.AsVariableDeclaration().Initializer == outer {
		declaration := outer.Parent
		if !ast.IsIdentifier(declaration.Name()) || (declaration.Parent != nil && declaration.Parent.Parent != nil && ast.HasSyntacticModifier(declaration.Parent.Parent, ast.ModifierFlagsExport)) {
			return refusal()
		}
		symbol := l.checker.GetSymbolAtLocation(declaration.Name())
		valid := true
		var visit ast.Visitor
		visit = func(read *ast.Node) bool {
			if !valid {
				return true
			}
			if ast.IsIdentifier(read) && read != declaration.Name() {
				reference := l.checker.GetSymbolAtLocation(read)
				if read.Parent != nil && read.Parent.Kind == ast.KindShorthandPropertyAssignment {
					reference = l.checker.GetShorthandAssignmentValueSymbol(read.Parent)
				}
				if reference == symbol && !l.libraryStringBoxUse(read) {
					valid = false
					return true
				}
			}
			read.ForEachChild(visit)
			return false
		}
		ast.GetSourceFileOfNode(node).AsNode().ForEachChild(visit)
		if !valid {
			return refusal()
		}
	} else if !l.libraryStringBoxUse(node) {
		return refusal()
	}
	arguments := node.AsNewExpression().Arguments
	value := ir.Expression(ir.StringConstant{Index: l.constant("")})
	if arguments != nil && len(arguments.Nodes) > 0 {
		if len(arguments.Nodes) != 1 || arguments.Nodes[0].Kind == ast.KindSpreadElement {
			return refusal()
		}
		var err error
		value, err = l.stringConversion(arguments.Nodes[0])
		if err != nil {
			return nil, err
		}
	}
	return ir.ObjectLiteral{NoReuse: true, Fields: []ir.Field{{Name: libraryStringData, Value: value}}}, nil
}

func (l *lowering) libraryStringBoxSlot(node *ast.Node) (ir.Expression, error) {
	object, err := l.expression(node)
	if err != nil {
		return nil, err
	}
	return ir.Property{Object: object, Name: libraryStringData, Of: ir.String}, nil
}
