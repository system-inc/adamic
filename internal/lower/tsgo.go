package lower

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/bridge/tsgo/spec"
	"github.com/system-inc/adamic/internal/ir"
)

// tsgo lowers direct calls only. Library functions borrow their inputs, mutate no
// Adamic value, and return fresh results. Ordinary IR calls preserve evaluation
// order and reference ownership. The bodies describe those effects for the
// analyses; the opt-in native library renderer supplies their implementation.
func (l *lowering) tsgo(node *ast.Node) (ir.Expression, bool, error) {
	callee := node.AsCallExpression().Expression
	name := ""
	for _, candidate := range []string{"tsgoProgram", "tsgoQuery", "tsgoRelease", "tsgoTypeParts", "tsgoInspect"} {
		if l.isPreludeFunction(callee, candidate) {
			name = candidate
			break
		}
	}
	if name == "" {
		return nil, false, nil
	}
	if !l.program.TSGoEnabled() {
		return nil, true, &Refused{Where: l.program.Where(node), What: "an unlinked typescript-go library call", Fix: "build with --tsgo <checker archive>"}
	}
	arguments := []ir.Expression{}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		arguments = append(arguments, value)
	}
	types := []ir.Type{ir.Number}
	returns := ir.Object
	switch name {
	case "tsgoProgram":
		types, returns = []ir.Type{ir.String, ir.Array}, ir.Object
	case "tsgoInspect":
		types, returns = []ir.Type{ir.Number, ir.String, ir.Number, ir.Number, ir.String, ir.String}, ir.Object
	case "tsgoTypeParts":
		types, returns = []ir.Type{ir.Number, ir.String, ir.Number, ir.Number, ir.String}, ir.Object
	case "tsgoQuery":
		types, returns = []ir.Type{ir.Number, ir.String, ir.Number}, ir.Object
	}
	if len(arguments) != len(types) {
		return nil, true, fmt.Errorf("lower: %s: invalid %s argument count", l.program.Where(node), name)
	}
	for index, argument := range arguments {
		if argument.Type() != types[index] {
			return nil, true, fmt.Errorf("lower: %s: invalid %s argument type", l.program.Where(node), name)
		}
	}
	function := -1
	for index, declared := range l.result.Functions {
		if declared.Name == spec.Prefix+name {
			function = index
			break
		}
	}
	if function < 0 {
		function = len(l.result.Functions)
		declared := ir.Function{Name: spec.Prefix + name, Returns: returns}
		for index, of := range types {
			local := len(l.result.Locals)
			l.result.Locals = append(l.result.Locals, ir.Local{Name: fmt.Sprintf("argument%d", index), Type: of, Function: function, Borrowed: of.IsReference()})
			declared.Parameters = append(declared.Parameters, local)
		}
		// A fresh result describes ownership to the analyses. Backend adapters
		// replace this body; ordinary rendering returns an honest refusal.
		kind := len(l.result.Strings)
		l.result.Strings = append(l.result.Strings, "Error", "tsgo: checker library renderer unavailable")
		declared.Body = []ir.Statement{ir.Return{Value: ir.ObjectLiteral{Fields: []ir.Field{
			{Name: "kind", Value: ir.StringConstant{Index: kind}},
			{Name: "message", Value: ir.StringConstant{Index: kind + 1}},
		}}}}

		l.result.Functions = append(l.result.Functions, declared)
	}
	return ir.Call{Function: function, Arguments: arguments, Returns: returns}, true, nil
}
