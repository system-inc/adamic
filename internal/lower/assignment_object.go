package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Hold the source once, then read and assign each named property in source order.
// Defaults, rest, computed names and nested targets retain their own stops.
func (l *lowering) objectDestructuringAssignment(pattern, source *ast.Node) ([]ir.Statement, error) {
	value, err := l.expression(source)
	if err != nil {
		return nil, err
	}
	if value.Type() != ir.Object {
		return nil, l.notYet(source, "object destructuring a "+typeName(value.Type()))
	}
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "destructure", Type: ir.Object, Function: l.functionIndex})
	body := []ir.Statement{ir.Declare{Local: held, Value: value}}
	for _, property := range pattern.AsObjectLiteralExpression().Properties.Nodes {
		var name, target *ast.Node
		switch property.Kind {
		case ast.KindShorthandPropertyAssignment:
			short := property.AsShorthandPropertyAssignment()
			if short.ObjectAssignmentInitializer != nil {
				return nil, l.notYet(property, "a default in an object assignment")
			}
			name, target = property.Name(), property.Name()
		case ast.KindPropertyAssignment:
			name, target = property.Name(), ast.SkipParentheses(property.AsPropertyAssignment().Initializer)
		default:
			return nil, l.notYet(property, "an object assignment property other than a named field")
		}
		if name.Kind != ast.KindIdentifier && name.Kind != ast.KindStringLiteral && name.Kind != ast.KindNumericLiteral {
			return nil, l.notYet(name, "a computed object assignment name")
		}
		symbol := l.symbol(target)
		if property.Kind == ast.KindShorthandPropertyAssignment {
			symbol = l.checker.GetShorthandAssignmentValueSymbol(property)
			if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
				symbol = l.checker.GetAliasedSymbol(symbol)
			}
			if symbol != nil {
				symbol = l.checker.GetExportSymbolOfSymbol(symbol)
			}
		}
		local, known := l.locals[symbol]
		if !ast.IsIdentifier(target) || !known || l.alwaysUndefined[symbol] || l.caught[symbol] {
			return nil, l.notYet(target, "assigning to "+describe(target)+" in an object assignment")
		}
		l.touch(local)
		if l.result.Locals[local].Captured && slotless(l.result.Locals[local].Type) {
			return nil, l.notYet(target, "a captured optional scalar in an object assignment")
		}
		field := l.checker.GetTypeOfPropertyOfType(l.checker.GetTypeAtLocation(source), name.Text())
		of, represented := l.representation(field)
		if !represented || slotless(of) || of == ir.Union || of == ir.Weak {
			return nil, l.notYet(property, "an object assignment field requiring another representation")
		}
		read := ir.Expression(ir.Property{Object: ir.Read{Local: held, Of: ir.Object}, Name: name.Text(), Of: of})
		to := l.result.Locals[local].Type
		read = fit(read, to)
		if read.Type() != to {
			return nil, l.notYet(property, "an object assignment field requiring another representation")
		}
		body = append(body, ir.Assign{Local: local, Value: read, Checked: l.checked(local)})
	}
	return []ir.Statement{ir.Block{Body: body}}, nil
}
