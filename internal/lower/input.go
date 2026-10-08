package lower

import (
	"errors"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// input lowers the program's doors in from outside, opened in 0.2: readTextFile(path) and
// programArguments() from 'adamic', the door out, writeTextFile(path, text), and what a walk of the
// file system needs, readDirectory(path) and fileStatus(path). isInput is false for any other call.
func (l *lowering) input(node *ast.Node) (ir.Expression, bool, error) {
	if value, found, err := l.nodeFSFile(node); found {
		return value, true, err
	}
	if value, found, err := l.tsgo(node); found {
		return value, true, err
	}
	callee := node.AsCallExpression().Expression
	arguments := node.AsCallExpression().Arguments.Nodes
	switch {
	case l.isPreludeFunction(callee, "parallelMap"):
		value, err := l.parallelMap(node)
		return value, true, err
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
	case l.isPreludeFunction(callee, "utf8Length") || l.isPreludeFunction(callee, "utf8At"):
		lowered := []ir.Expression{}
		for _, argument := range arguments {
			value, err := l.expression(argument)
			if err != nil {
				return nil, true, err
			}
			lowered = append(lowered, value)
		}
		if l.isPreludeFunction(callee, "utf8Length") {
			if len(lowered) != 1 || lowered[0].Type() != ir.String {
				return nil, true, errors.New("lower: " + l.program.Where(node) + ": utf8Length takes one string, and the checker let something else through")
			}
			return ir.Utf8Length{Text: lowered[0]}, true, nil
		}
		if len(lowered) != 2 || lowered[0].Type() != ir.String || lowered[1].Type() != ir.Number {
			return nil, true, errors.New("lower: " + l.program.Where(node) + ": utf8At takes a string and a number, and the checker let something else through")
		}
		return ir.Utf8At{Text: lowered[0], Index: lowered[1]}, true, nil
	case l.isPreludeFunction(callee, "readDirectory"), l.isPreludeFunction(callee, "fileStatus"), l.isPreludeFunction(callee, "realPath"):
		// By what it is, not what it's called here: an import may name it anything.
		name := "fileStatus"
		if l.isPreludeFunction(callee, "readDirectory") {
			name = "readDirectory"
		}
		if l.isPreludeFunction(callee, "realPath") {
			name = "realPath"
		}
		if len(arguments) != 1 {
			return nil, true, errors.New("lower: " + l.program.Where(node) + ": " + name + " takes one path, and the checker let another count through")
		}
		path, err := l.expression(arguments[0])
		if err != nil {
			return nil, true, err
		}
		if path.Type() != ir.String {
			return nil, true, errors.New("lower: " + l.program.Where(arguments[0]) + ": " + name + "'s path isn't a string, and the checker let it through")
		}
		if name == "readDirectory" {
			return ir.ReadDirectory{Path: path}, true, nil
		}
		if name == "realPath" {
			return ir.RealPath{Path: path}, true, nil
		}
		return ir.FileStatus{Path: path}, true, nil
	case l.isPreludeFunction(callee, "programArguments"):
		if len(arguments) != 0 {
			return nil, true, errors.New("lower: " + l.program.Where(node) + ": programArguments takes nothing, and the checker let arguments through")
		}
		return ir.ProgramArguments{}, true, nil
	}
	return nil, false, nil
}
