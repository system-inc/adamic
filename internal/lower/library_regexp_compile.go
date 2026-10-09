package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Dynamic construction uses the same checked compiler as the Go reference.
// Its implementation is included only by a module that reaches this operation.
func (l *lowering) dynamicRegExp(node *ast.Node, args []*ast.Node) (ir.Expression, error) {
	arguments := make([]ir.Expression, 0, 2)
	for _, arg := range args {
		value, err := l.expression(arg)
		if err != nil {
			return nil, err
		}
		if _, undefined := value.(ir.Undefined); undefined {
			value = ir.StringConstant{Index: l.constant("")}
		}
		if value.Type() != ir.String && value.Type() != ir.Maybe(ir.String) {
			return nil, l.notYet(arg, "dynamic RegExp argument other than a string or undefined")
		}
		arguments = append(arguments, value)
	}
	for len(arguments) < 2 {
		arguments = append(arguments, ir.StringConstant{Index: l.constant("")})
	}
	return l.runtimeRegExpNew(arguments), nil
}

func (l *lowering) runtimeRegExpNew(arguments []ir.Expression) ir.Expression {
	const marker = "#define ADAMIC_REGEXP_RUNTIME_COMPILER 1\n"
	found := false
	for _, program := range l.result.Regexps {
		if strings.HasPrefix(program.Declarations, marker) {
			found = true
			break
		}
	}
	if !found {
		l.result.Regexps = append(l.result.Regexps, ir.RegExpProgram{Declarations: marker + `#include "regexp_compile_parser.c"
#include "regexp_compile_properties.c"
#include "regexp_compile_sets.c"
#include "regexp_compile_bytecode.c"
#include "regexp_compile_v8.c"
#include "regexp_compile_runtime.c"
`})
	}
	return ir.RegExpNew{Index: -1, Arguments: arguments}
}
