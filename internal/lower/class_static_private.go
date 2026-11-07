package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (lowering *lowering) privateStaticMember(name *ast.Node) *ast.Node {
	if name.Kind != ast.KindPrivateIdentifier {
		return nil
	}
	symbol := lowering.checker.GetSymbolAtLocation(name)
	if symbol == nil ||
		len(symbol.Declarations) == 0 ||
		!ast.HasSyntacticModifier(symbol.Declarations[0], ast.ModifierFlagsStatic) {
		return nil
	}
	return symbol.Declarations[0]
}

// A static private brand is exact, rather than inherited. Keep the check as a call so
// exception, ownership and reuse analyses see the throw and the receiver's lifetime.
func (lowering *lowering) privateStaticReceiver(name *ast.Node, object ir.Expression, write bool) ir.Expression {
	member := lowering.privateStaticMember(name)
	if member == nil {
		return object
	}
	class := lowering.statics[lowering.symbol(member.Parent.Name())]
	message := "Receiver must be class " + member.Parent.Name().Text()
	if member.Kind == ast.KindPropertyDeclaration {
		message = "Cannot read private member " + name.Text() + " from an object whose class did not declare it"
		if write {
			message = "Cannot write private member " + name.Text() + " to an object whose class did not declare it"
		}
	}
	index := len(lowering.result.Functions)
	self := len(lowering.result.Locals)
	lowering.result.Locals = append(lowering.result.Locals, ir.Local{Name: "this", Type: ir.Object, Function: index})
	value := ir.Read{Local: self, Of: ir.Object}
	lowering.result.Functions = append(
		lowering.result.Functions,
		ir.Function{Name: "static_private_brand", Parameters: []int{self}, Returns: ir.Object, Body: []ir.Statement{
			ir.If{
				Condition: ir.InstanceOf{Value: value, Class: class.class, Exact: true},
				Then:      []ir.Statement{ir.Return{Value: value}},
			},
			ir.Throw{Value: ir.MakeError{
				Message: ir.StringConstant{Index: lowering.constant(message)},
				Name:    ir.StringConstant{Index: lowering.constant("TypeError")},
			}},
		}},
	)
	return ir.Call{Function: index, Arguments: []ir.Expression{object}, Returns: ir.Object}
}

// A private assignment checks its brand after the RHS has executed, as JavaScript does.
func (lowering *lowering) privateStaticStore(target *ast.Node, object, value ir.Expression) (ir.Expression, bool) {
	if lowering.privateStaticMember(target.Name()) == nil {
		return nil, false
	}
	index := len(lowering.result.Functions)
	self := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "this", Type: ir.Object, Function: index},
		ir.Local{Name: "value", Type: value.Type(), Function: index},
	)
	lowering.result.Functions = append(
		lowering.result.Functions,
		ir.Function{Name: "static_private_write", Parameters: []int{self, self + 1}},
	)
	receiver := lowering.privateStaticReceiver(target.Name(), ir.Read{Local: self, Of: ir.Object}, true)
	lowering.result.Functions[index].Body = []ir.Statement{ir.SetProperty{
		Object: receiver,
		Name:   lowering.fieldName(target.Name()),
		Value:  ir.Read{Local: self + 1, Of: value.Type()},
		Site:   lowering.writeSite(target.AsPropertyAccessExpression().Expression),
	}}
	return ir.Call{Function: index, Arguments: []ir.Expression{object, value}}, true
}
