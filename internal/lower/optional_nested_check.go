package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func (l *lowering) checkedOptionalNested(node *ast.Node, value ir.Expression, contract *load.OptionalObjectContract) (ir.Expression, error) {
	expected := ir.Object
	if contract.Kind == "array" {
		expected = ir.Array
	}
	if contract.Kind == "function" {
		expected = ir.Closure
		result, known := l.representation(contract.Result)
		if !known || result != ir.Object {
			return nil, l.notYet(node, "optional callback requires a presence-capable object return ABI")
		}
	}
	if value.Type() != expected {
		return nil, l.notYet(node, "nested optional contract requires its presence-capable storage ABI")
	}
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "optionalContract", Type: expected, Function: l.functionIndex})
	l.noteLocal(local, l.concrete(l.checker.GetTypeAtLocation(node)), node)
	read := ir.Read{Local: local, Of: expected}
	var guards []ir.Statement
	var err error
	switch contract.Kind {
	case "function":
		guards = []ir.Statement{ir.Evaluate{Value: ir.ObjectCall{Method: "optionalFunctionStorage", Arguments: []ir.Expression{read}, Returns: ir.Boolean, Readiness: l.program.Where(node)}}}
	case "array":
		element, known := l.representation(contract.Element.Source)
		if !known || element != ir.Object || len(contract.Element.Children) > 0 || contract.Element.Kind != "" {
			return nil, l.notYet(node, "optional array contract requires plain object elements with fixed own fields")
		}
		keys := ir.ArrayLiteral{Element: ir.String}
		for _, name := range contract.Element.Fields {
			keys.Elements = append(keys.Elements, ir.StringConstant{Index: l.constant(name)})
		}
		guards = []ir.Statement{ir.Evaluate{Value: ir.ObjectCall{Method: "optionalArrayPresence", Arguments: []ir.Expression{read, ir.BooleanConstant{Value: contract.Nullable}, keys, ir.BooleanConstant{Value: contract.Element.Nullable}}, Returns: ir.Boolean, Readiness: l.program.Where(node)}}}
	default:
		guards, err = l.optionalNestedGuards(node, read, contract)
	}
	if err != nil {
		return nil, err
	}
	body := append([]ir.Statement{ir.Declare{Local: local, Value: value}}, guards...)
	l.program.RecordOptionalLiteral(node)
	return ir.Effects{Body: body, Result: read}, nil
}

func (l *lowering) optionalNestedGuards(node *ast.Node, read ir.Expression, contract *load.OptionalObjectContract) ([]ir.Statement, error) {
	body := []ir.Statement{ir.Evaluate{Value: ir.ObjectCall{Method: "optionalViewStorage", Arguments: []ir.Expression{read, ir.BooleanConstant{Value: contract.Nullable}}, Returns: ir.Boolean, Readiness: l.program.Where(node)}}}
	fields := []ir.Statement{}
	own := func(name string) ir.Statement {
		return ir.Evaluate{Value: ir.ObjectCall{Method: "optionalWritePresence", Arguments: []ir.Expression{read, ir.StringConstant{Index: l.constant(name)}}, Returns: ir.Boolean, Readiness: l.program.Where(node)}}
	}
	for _, name := range contract.Fields {
		fields = append(fields, own(name))
	}
	for _, child := range contract.Children {
		of, known := l.representation(child.Contract.Source)
		if !known || of != ir.Object {
			return nil, l.notYet(node, "nested optional field requires own-presence object storage")
		}
		if !child.Optional {
			fields = append(fields, own(child.Name))
		}
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "optionalNestedField", Type: ir.Object, Function: l.functionIndex})
		l.noteLocal(local, l.concrete(child.Contract.Source), node)
		value := ir.Property{Object: read, Name: child.Name, Of: ir.Object, Absent: child.Optional}
		fields = append(fields, ir.Declare{Local: local, Value: value})
		nested, err := l.optionalNestedGuards(node, ir.Read{Local: local, Of: ir.Object}, child.Contract)
		if err != nil {
			return nil, err
		}
		fields = append(fields, nested...)
	}
	if contract.Nullable {
		body = append(body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: read}}, Then: fields})
	} else {
		body = append(body, fields...)
	}
	return body, nil
}
