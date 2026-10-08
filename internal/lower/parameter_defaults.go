package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Each parameter is bound before the next initializer runs. An object pattern
// selects its undefined-only default once, then reads fields from that holder.
func (l *lowering) parameterPrologue(declaration *ast.Node, defaults []defaulted, patterns []patterned) ([]ir.Statement, error) {
	var statements []ir.Statement
	for _, source := range declaration.Parameters() {
		for _, parameter := range defaults {
			if parameter.initializer.Parent != source {
				continue
			}
			if l.uninitializedInitializer(parameter.initializer) {
				l.result.Locals[parameter.local].Uninitialized = true
				incoming := ir.Read{Local: parameter.incoming, Of: l.result.Locals[parameter.incoming].Type}
				statements = append(statements, ir.Declare{Local: parameter.local, Uninitialized: true}, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: incoming}}, Then: []ir.Statement{ir.Assign{Local: parameter.local, Value: fit(incoming, l.result.Locals[parameter.local].Type)}}})
				continue
			}

			if assertionInitializer(parameter.initializer) {
				prefix, present, value, lazyErr := l.lazyAssertion(parameter.initializer, l.result.Locals[parameter.local].Type)
				if lazyErr != nil {
					return nil, lazyErr
				}
				l.result.Locals[parameter.local].Uninitialized = true
				l.result.Locals[parameter.local].InitializerExpression = sourceExpression(parameter.initializer)
				incoming := ir.Read{Local: parameter.incoming, Of: l.result.Locals[parameter.incoming].Type}
				fallback := append(prefix, ir.If{Condition: present, Then: []ir.Statement{ir.Assign{Local: parameter.local, Value: value}}})
				statements = append(statements, ir.Declare{Local: parameter.local, Uninitialized: true}, ir.If{Condition: ir.IsUndefined{Value: incoming}, Then: fallback, Else: []ir.Statement{ir.Assign{Local: parameter.local, Value: fit(incoming, l.result.Locals[parameter.local].Type)}}})
				continue
			}
			fallback, err := l.expression(parameter.initializer)
			if err != nil {
				return nil, err
			}
			of := l.result.Locals[parameter.local].Type
			fallback = fit(fallback, of)
			if fallback.Type() != of {
				return nil, l.notYet(parameter.initializer, "a default of another type than its parameter")
			}
			incoming := ir.Read{Local: parameter.incoming, Of: l.result.Locals[parameter.incoming].Type}
			statements = append(statements, ir.Declare{Local: parameter.local, Value: ir.Coalesce{Value: incoming, Fallback: fallback, Of: of, UndefinedOnly: true}})
		}
		for _, parameter := range patterns {
			if parameter.parameter != source {
				continue
			}
			held := parameter.incoming
			if initializer := source.AsParameterDeclaration().Initializer; initializer != nil {
				initialization, local, err := l.objectPatternDefault(parameter, initializer)
				if err != nil {
					return nil, err
				}
				statements = append(statements, initialization...)
				held = local
			}
			proven := l.checker.GetTypeAtLocation(source)
			of, _ := l.representation(proven)
			destructured, err := l.destructureFrom(parameter.pattern, proven, of, held)
			if err != nil {
				return nil, err
			}
			statements = append(statements, destructured...)
		}
	}
	return statements, nil
}

func (l *lowering) objectPatternDefault(parameter patterned, initializer *ast.Node) ([]ir.Statement, int, error) {
	if l.uninitializedInitializer(initializer) || assertionInitializer(initializer) {
		return nil, 0, l.notYet(initializer, "a destructured object default with an unproven initializer")
	}
	fallback, err := l.expression(initializer)
	if err != nil {
		return nil, 0, err
	}
	if fallback.Type() != ir.Object {
		return nil, 0, l.notYet(initializer, "a destructured object default held as another representation")
	}
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "parameter_default", Type: ir.Object, Function: l.functionIndex})
	incoming := ir.Read{Local: parameter.incoming, Of: ir.Object}
	return []ir.Statement{ir.Declare{Local: local, Value: ir.Coalesce{Value: incoming, Fallback: fallback, Of: ir.Object, UndefinedOnly: true}}}, local, nil
}
