package lower

import (
	"errors"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// input lowers the program's doors in from outside, opened in 0.2: readTextFile(path) and
// programArguments() from 'adamic', and the door out, writeTextFile(path, text). isInput is false for any other call.
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
	case l.isPreludeFunction(callee, "writeTextFile"):
		if len(arguments) != 2 {
			return nil, true, errors.New("lower: " + l.program.Where(node) + ": writeTextFile takes a path and a text, and the checker let another count through")
		}
		path, err := l.expression(arguments[0])
		if err != nil {
			return nil, true, err
		}
		text, err := l.expression(arguments[1])
		if err != nil {
			return nil, true, err
		}
		if path.Type() != ir.String || text.Type() != ir.String {
			return nil, true, errors.New("lower: " + l.program.Where(node) + ": writeTextFile's path or text isn't a string, and the checker let it through")
		}
		return ir.WriteTextFile{Path: path, Text: text}, true, nil
	case l.isPreludeFunction(callee, "programArguments"):
		if len(arguments) != 0 {
			return nil, true, errors.New("lower: " + l.program.Where(node) + ": programArguments takes nothing, and the checker let arguments through")
		}
		return ir.ProgramArguments{}, true, nil
	}
	return nil, false, nil
}
