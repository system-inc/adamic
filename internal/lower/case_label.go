package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// caseLabel uses an imported const's declared type, rather than a narrowing at this read. Its
// literal representation is known even in a cycle. Retain the runtime read so an unfinished
// provider still raises the ordinary module dead-zone error.
func (l *lowering) caseLabel(node *ast.Node) (ir.Expression, error) {
	name := ast.SkipParentheses(node)
	if name.Kind != ast.KindIdentifier {
		return l.expression(node)
	}
	symbol := l.checker.GetSymbolAtLocation(name)
	if symbol == nil || symbol.Flags&ast.SymbolFlagsAlias == 0 {
		return l.expression(node)
	}
	symbol = l.symbol(name)
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 || declaration.AsVariableDeclaration().Initializer == nil {
			continue
		}
		proven := l.checker.GetTypeOfSymbol(symbol)
		if proven.Flags()&(checker.TypeFlagsNumberLiteral|checker.TypeFlagsStringLiteral) != 0 {
			if constant, _, known := l.literalConstant(proven); known {
				if local, known := l.local(name); known {
					return ir.Read{Local: local, Of: constant.Type(), Checked: l.checkedModuleRead(name, local), Readiness: sourceExpression(name)}, nil
				}
				return nil, l.notYet(name, "an imported case label without runtime storage")
			}
		}
		return nil, l.notYet(name, "a case label using const "+declaration.Name().Text()+" with type "+l.checker.TypeToString(proven)+" (not a number or string literal type)")
	}
	return l.expression(node)
}
