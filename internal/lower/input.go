package lower

import (
	"errors"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// input lowers the program's doors in from outside, opened in 0.2: readTextFile(path) and
// programArguments() from 'adamic'. isInput is false for any other call.
func (l *lowering) input(node *ast.Node) (ir.Expression, bool, error) {
	callee := node.AsCallExpression().Expression
	arguments := node.AsCallExpression().Arguments.Nodes
	switch {
	case l.isPreludeFunction(callee, "readTextFile"):
		if len(arguments) != 1 {
			return nil, true, errors.New("lower: " + l.program.Where(node) + ": readTextFile takes one path, and the checker let another count through")
		}
		path, err := l.expression(arguments[0])
		if err != nil {
			return nil, true, err
		}
		if path.Type() != ir.String {
			return nil, true, errors.New("lower: " + l.program.Where(arguments[0]) + ": readTextFile's path isn't a string, and the checker let it through")
		}
		return ir.ReadTextFile{Path: path}, true, nil
	case l.isPreludeFunction(callee, "programArguments"):
		if len(arguments) != 0 {
			return nil, true, errors.New("lower: " + l.program.Where(node) + ": programArguments takes nothing, and the checker let arguments through")
		}
		return ir.ProgramArguments{}, true, nil
	}
	return nil, false, nil
}
