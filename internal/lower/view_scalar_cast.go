package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A scalar assertion checks its owned value at the cast. This does not admit
// phantom brands, intersections, any or unknown, nor invent facts about objects.
func (l *lowering) scalarCastCandidate(_ *ast.Node, source, target *checker.Type) bool {
	if !interfaceScalar(target) || l.includesUndefined(target) {
		return false
	}
	to, known := l.representation(target)
	if !known || (to != ir.Number && to != ir.String && to != ir.Boolean) {
		return false
	}
	from, known := l.representation(source)
	if !known || (from != ir.Union && from != ir.String && from != ir.Number && from != ir.Boolean && from != ir.MaybeNumber && from != ir.MaybeBoolean) {
		return false
	}
	// Proven scalar upcasts keep their existing proof, including enum branding.
	return !l.checker.IsTypeAssignableTo(source, target)
}

func (l *lowering) scalarCast(node *ast.Node, value ir.Expression, source, target *checker.Type) (ir.Expression, error) {
	if !l.scalarCastCandidate(node, source, target) {
		return nil, nil
	}
	to, known := l.representation(target)
	if !known {
		return nil, nil
	}
	name := map[ir.Type]string{ir.String: "string", ir.Number: "number", ir.Boolean: "boolean"}[to]
	function := len(l.result.Functions)
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "cast_value", Type: ir.Union, Function: function})
	read := ir.Read{Local: local, Of: ir.Union}
	prefix := "cast failed: " + sourceExpression(node.AsAsExpression().Expression) + " expected " + l.checker.TypeToString(target) + ", found "
	failure := ir.Panic{Message: ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: l.constant(prefix)}, ir.TypeOf{Value: read}}}}
	narrowed := ir.Narrow{Value: read, To: to}
	success := []ir.Statement{ir.Return{Value: narrowed}}
	if allowed := l.viewLiterals(target); len(allowed) > 0 {
		var condition ir.Expression
		for _, literal := range allowed {
			test := ir.Binary{Operator: ir.Equal, Left: narrowed, Right: literal}
			if condition == nil {
				condition = test
			} else {
				condition = ir.Binary{Operator: ir.Or, Left: condition, Right: test}
			}
		}
		success = []ir.Statement{ir.If{Condition: condition, Then: success}, failure}
	}
	kind := ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: read}, Right: ir.StringConstant{Index: l.constant(name)}}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "checked_scalar_cast", Parameters: []int{local}, Returns: to, Body: []ir.Statement{ir.If{Condition: kind, Then: success}, failure}})
	return ir.Call{Function: function, Arguments: []ir.Expression{fit(value, ir.Union)}, Returns: to}, nil
}
