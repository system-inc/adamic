package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) closureLength(v ir.ClosureLength) string {
	if !e.functionLengthDeclared {
		var table strings.Builder
		table.WriteString("static double adamic_function_length(adamic_closure *value) {\n")
		for index, function := range e.program.Functions {
			if function.Closure {
				fmt.Fprintf(&table, "\tif (value->code == %s) return %d;\n", e.functionName(index), function.SourceLength)
			}
		}
		// The only runtime-created closure is a collection iterator's zero-argument next.
		table.WriteString("\t(void)value;\n\treturn 0;\n}\n")
		e.declarations = append(e.declarations, table.String())
		e.functionLengthDeclared = true
	}
	value := e.value(v.Value)
	if v.Optional {
		return e.snapshot(ir.MaybeNumber, fmt.Sprintf("(%s == NULL ? %s : (adamic_maybe_number){true, adamic_function_length(%s)})", value, zero(ir.MaybeNumber), value))
	}
	return e.snapshot(ir.Number, fmt.Sprintf("adamic_function_length(%s)", value))
}
