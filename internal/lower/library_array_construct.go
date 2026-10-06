package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Port of V8 ArrayConstructInitializeElements in src/objects/elements.cc:
// zero arguments make an empty array, one number is a length, and otherwise
// arguments become elements in order. Nonzero lengths require sparse presence.
func (l *lowering) libraryArrayConstruct(node *ast.Node) (ir.Expression, bool, error) {
	var callee *ast.Node
	var written []*ast.Node
	if node.Kind == ast.KindNewExpression {
		created := node.AsNewExpression()
		callee = created.Expression
		if created.Arguments != nil {
			written = created.Arguments.Nodes
		}
	} else {
		call := node.AsCallExpression()
		callee, written = call.Expression, call.Arguments.Nodes
	}
	callee = ast.SkipParentheses(callee)
	isOf := callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "of" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Array")
	if !isOf && !l.isLibraryGlobal(callee, "Array") {
		return nil, false, nil
	}
	for _, argument := range written {
		if argument.Kind == ast.KindSpreadElement {
			return nil, true, l.notYet(node, "Array construction with spread arguments")
		}
	}
	if !isOf && len(written) == 1 {
		proven := l.checker.GetTypeAtLocation(written[0])
		if proven.Flags()&checker.TypeFlagsUnion != 0 {
			for _, member := range proven.Types() {
				if member.Flags()&checker.TypeFlagsNumberLike != 0 {
					return nil, true, l.notYet(node, "single Array argument with an ambiguous numeric length overload")
				}
			}
		}
		if of, known := l.representation(proven); known && of == ir.Number {
			value, err := l.expression(written[0])
			if err != nil {
				return nil, true, err
			}
			if constant, ok := value.(ir.NumberConstant); !ok || constant.Value != 0 {
				return nil, true, l.notYet(node, "Array length construction requiring sparse presence (only a constant zero is dense)")
			}
			written = nil
		}
	}
	if err := l.libraryArrayConstructWrites(node, len(written)); err != nil {
		return nil, true, err
	}
	element, err := l.elementType(node)
	if err != nil {
		return nil, true, err
	}
	literal := ir.ArrayLiteral{Element: element}
	for _, argument := range written {
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		value = fit(value, element)
		if value.Type() != element {
			return nil, true, l.notYet(node, "Array construction with incompatible element representations")
		}
		literal.Elements = append(literal.Elements, value)
	}
	return literal, true, nil
}

// Newly admitted constructors must not unblock writes that need sparse presence
// or non-array integer properties. Keep these language features refused before C.
func (l *lowering) libraryArrayConstructWrites(node *ast.Node, initialLength int) error {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent == nil || outer.Parent.Kind != ast.KindVariableDeclaration || !ast.IsIdentifier(outer.Parent.Name()) {
		return nil
	}
	symbol := l.symbol(outer.Parent.Name())
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	var unsupported *ast.Node
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if unsupported != nil {
			return true
		}
		if part.Kind == ast.KindBinaryExpression && part.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
			target := ast.SkipParentheses(part.AsBinaryExpression().Left)
			if ast.IsIdentifier(target) && l.symbol(target) == symbol {
				// A reassignment invalidates the original dense prefix proof.
				initialLength = 0
			}
			if target.Kind == ast.KindElementAccessExpression {
				access := target.AsElementAccessExpression()
				receiver := ast.SkipParentheses(access.Expression)
				index := ast.SkipParentheses(access.ArgumentExpression)
				if ast.IsIdentifier(receiver) && l.symbol(receiver) == symbol && index.Kind == ast.KindNumericLiteral {
					value, err := l.expression(index)
					if err == nil {
						if constant, ok := value.(ir.NumberConstant); ok && constant.Value >= float64(initialLength) {
							unsupported = part
							return true
						}
					}
				}
			}
		}
		return part.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if unsupported != nil {
		return l.notYet(unsupported, "Array construction followed by an indexed write outside a proven dense prefix (indexed extension, sparse presence or non-array integer properties)")
	}
	return nil
}
