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
		return nil, true, &Refused{Where: l.program.Where(node), What: "an unlinked typescript-go library call", Fix: "build with --tsgo <checker archive> (adamic/tsgo-link)"}
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
	returns := ir.Type(0)
	switch name {
	case "tsgoProgram":
		types, returns = []ir.Type{ir.String, ir.Array}, ir.Number
	case "tsgoInspect":
		types, returns = []ir.Type{ir.Number, ir.String, ir.Number, ir.Number, ir.String, ir.String}, ir.String
	case "tsgoTypeParts":
		types, returns = []ir.Type{ir.Number, ir.String, ir.Number, ir.Number, ir.String}, ir.String
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
		switch name {
		case "tsgoProgram":
			declared.Body = []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: 0}}}
		case "tsgoTypeParts", "tsgoInspect":
			empty := len(l.result.Strings)
			l.result.Strings = append(l.result.Strings, "")
			declared.Body = []ir.Statement{ir.Return{Value: ir.StringConstant{Index: empty}}}
		case "tsgoQuery":
			empty := len(l.result.Strings)
			l.result.Strings = append(l.result.Strings, "")
			declared.Body = []ir.Statement{ir.Return{Value: ir.ObjectLiteral{Fields: []ir.Field{
				{Name: "nodeKind", Value: ir.NumberConstant{Value: 0}},
				{Name: "symbolName", Value: ir.StringConstant{Index: empty}},
				{Name: "type", Value: ir.StringConstant{Index: empty}},
			}}}}
		}
		// Also fail loudly if an opted-in IR is sent to an ordinary backend.
		message := len(l.result.Strings)
		l.result.Strings = append(l.result.Strings, "tsgo requires the native checker library renderer")
		declared.Body = append([]ir.Statement{ir.Panic{Message: ir.StringConstant{Index: message}}}, declared.Body...)
		l.result.Functions = append(l.result.Functions, declared)
	}
	return ir.Call{Function: function, Arguments: arguments, Returns: returns}, true, nil
}
