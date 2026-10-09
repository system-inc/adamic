package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) sourceFunctionSurface(node *ast.Node) *ir.FunctionSurface {
	surface := &ir.FunctionSurface{}
	if name := node.Name(); name != nil {
		if name.Kind == ast.KindComputedPropertyName {
			resolved, known := l.methodName(node)
			if !known {
				return nil
			}
			if resolved == iteratorSlot {
				resolved = "[Symbol.iterator]"
			}
			surface.Name = resolved
		} else {
			surface.Name = name.Text()
		}
	} else {
		outer := node
		for outer.Parent != nil {
			switch outer.Parent.Kind {
			case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression:
				outer = outer.Parent
			default:
				goto named
			}
		}
	named:
		if parent := outer.Parent; parent != nil {
			switch parent.Kind {
			case ast.KindVariableDeclaration, ast.KindPropertyAssignment, ast.KindPropertyDeclaration, ast.KindBindingElement, ast.KindParameter:
				if name := parent.Name(); name != nil && (ast.IsIdentifier(name) || name.Kind == ast.KindPrivateIdentifier) {
					surface.Name = name.Text()
				}
				if name := parent.Name(); name != nil && name.Kind == ast.KindStringLiteral {
					surface.Name = name.Text()
				}
				if name := parent.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
					resolved, known := l.methodName(parent)
					if !known {
						return nil
					}
					if resolved == iteratorSlot {
						resolved = "[Symbol.iterator]"
					}
					surface.Name = resolved
				}
			case ast.KindBinaryExpression:
				binary := parent.AsBinaryExpression()
				operator := binary.OperatorToken.Kind
				if (operator == ast.KindEqualsToken || operator == ast.KindBarBarEqualsToken || operator == ast.KindAmpersandAmpersandEqualsToken || operator == ast.KindQuestionQuestionEqualsToken) && binary.Right == outer && ast.IsIdentifier(binary.Left) {
					surface.Name = binary.Left.Text()
				}
			}
		}
	}
	for _, parameter := range node.Parameters() {
		if ast.IsIdentifier(parameter.Name()) && parameter.Name().Text() == "this" {
			continue
		}
		declared := parameter.AsParameterDeclaration()
		if declared.Initializer != nil || declared.DotDotDotToken != nil {
			break
		}
		surface.Length++
	}
	return surface
}

func (l *lowering) closureSurfaceRead(node *ast.Node) (ir.Expression, bool, error) {
	access := node.AsPropertyAccessExpression()
	name := access.Name().Text()
	if name != "length" && name != "name" {
		return nil, false, nil
	}
	of, known := l.representation(l.checker.GetTypeAtLocation(access.Expression))
	if !known || of != ir.Closure {
		return nil, false, nil
	}
	// Library constructor identities and opaque intrinsic tokens use their existing
	// observation adapters, rather than pretending their shim code is the source.
	if l.librarySymbol(l.symbol(access.Expression)) {
		return nil, false, nil
	}
	if access.QuestionDotToken != nil {
		return nil, true, l.notYet(node, "optional reflection on a function value")
	}
	value, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	result := ir.String
	if name == "length" {
		result = ir.Number
	}
	return ir.Property{Object: value, Name: name, Of: result}, true, nil
}
