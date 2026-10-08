package lower

import "github.com/system-inc/adamic/internal/ir"

// Select a default object once, before reading any fields from its binding
// pattern. The reference's absent sentinel means an omitted/undefined argument;
// the checker rejects null for this non-null object parameter.
func (l *lowering) patternDefault(parameter patterned, heldAs ir.Type) ([]ir.Statement, int, error) {
	initializer := parameter.parameter.AsParameterDeclaration().Initializer
	if initializer == nil {
		return nil, parameter.incoming, nil
	}
	if heldAs != ir.Object {
		return nil, 0, l.notYet(parameter.parameter, "a default for a destructured "+typeName(heldAs)+" parameter")
	}
	fallback, err := l.expression(initializer)
	if err != nil {
		return nil, 0, err
	}
	if fallback.Type() != ir.Object {
		return nil, 0, l.notYet(initializer, "a destructured object's default held otherwise than its parameter")
	}
	whole := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "destructuredDefault", Type: ir.Object, Function: l.functionIndex})
	incoming := ir.Read{Local: parameter.incoming, Of: ir.Object}
	selection := ir.Coalesce{Value: incoming, Fallback: fallback, Of: ir.Object}
	return []ir.Statement{ir.Declare{Local: whole, Value: selection}}, whole, nil
}
