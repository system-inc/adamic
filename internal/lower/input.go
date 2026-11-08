package lower

import (
	"errors"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// inputFunctions are the prelude functions input lowers.
var inputFunctions = []string{
	"readTextFile",
	"writeTextFile",
	"utf8Length",
	"utf8At",
	"readDirectory",
	"fileStatus",
	"programArguments",
}

// input lowers the program's doors in from outside, opened in 0.2: readTextFile(path) and
// programArguments() from 'adamic', the door out, writeTextFile(path, text), and what a walk of the
// file system needs, readDirectory(path) and fileStatus(path). isInput is false for any other call.
func (lowering *lowering) input(node *ast.Node) (ir.Expression, bool, error) {
	if value, found, nodeHostError := lowering.nodeHost(node); found {
		return value, true, nodeHostError
	}
	if value, found, nodeFileSystemFileError := lowering.nodeFileSystemFile(node); found {
		return value, true, nodeFileSystemFileError
	}
	if value, found, encodeJsonError := lowering.encodeJson(node); found {
		return value, true, encodeJsonError
	}
	if value, found, decodeJsonError := lowering.decodeJson(node); found {
		return value, true, decodeJsonError
	}
	if value, found, tsgoError := lowering.tsgo(node); found {
		return value, true, tsgoError
	}
	callee := node.AsCallExpression().Expression
	arguments := node.AsCallExpression().Arguments.Nodes
	// A spread of a tuple is a count the checker accepts and the cases below can't see, so it is
	// refused here, before any of them reads it as a checker bug.
	for _, name := range inputFunctions {
		if !lowering.isPreludeFunction(callee, name) {
			continue
		}
		for _, argument := range arguments {
			if argument.Kind == ast.KindSpreadElement {
				return nil, true, lowering.notYet(argument, "a spread argument to "+name)
			}
		}
	}
	switch {
	case lowering.isPreludeFunction(callee, "parallelMap"):
		value, parallelMapError := lowering.parallelMap(node)
		return value, true, parallelMapError
	case lowering.isPreludeFunction(callee, "readTextFile"):
		if len(arguments) != 1 {
			return nil, true, errors.New("lower: " +
				lowering.program.Where(node) +
				": readTextFile takes one path, and the checker let another count through")
		}
		path, expressionError := lowering.expression(arguments[0])
		if expressionError != nil {
			return nil, true, expressionError
		}
		if path.Type() != ir.String {
			return nil, true, errors.New("lower: " +
				lowering.program.Where(arguments[0]) +
				": readTextFile's path isn't a string, and the checker let it through")
		}
		return ir.ReadTextFile{Path: path}, true, nil
	case lowering.isPreludeFunction(callee, "writeTextFile"):
		if len(arguments) != 2 {
			return nil, true, errors.New("lower: " +
				lowering.program.Where(node) +
				": writeTextFile takes a path and a text, and the checker let another count through")
		}
		path, expressionError := lowering.expression(arguments[0])
		if expressionError != nil {
			return nil, true, expressionError
		}
		text, expressionError := lowering.expression(arguments[1])
		if expressionError != nil {
			return nil, true, expressionError
		}
		if path.Type() != ir.String || text.Type() != ir.String {
			return nil, true, errors.New("lower: " +
				lowering.program.Where(node) +
				": writeTextFile's path or text isn't a string, and the checker let it through")
		}
		return ir.WriteTextFile{Path: path, Text: text}, true, nil
	case lowering.isPreludeFunction(callee, "utf8Length") || lowering.isPreludeFunction(callee, "utf8At"):
		lowered := []ir.Expression{}
		for _, argument := range arguments {
			value, expressionError := lowering.expression(argument)
			if expressionError != nil {
				return nil, true, expressionError
			}
			lowered = append(lowered, value)
		}
		if lowering.isPreludeFunction(callee, "utf8Length") {
			if len(lowered) != 1 || lowered[0].Type() != ir.String {
				return nil, true, errors.New("lower: " +
					lowering.program.Where(node) +
					": utf8Length takes one string, and the checker let something else through")
			}
			return ir.Utf8Length{Text: lowered[0]}, true, nil
		}
		if len(lowered) != 2 || lowered[0].Type() != ir.String || lowered[1].Type() != ir.Number {
			return nil, true, errors.New("lower: " +
				lowering.program.Where(node) +
				": utf8At takes a string and a number, and the checker let something else through")
		}
		return ir.Utf8At{Text: lowered[0], Index: lowered[1]}, true, nil
	case lowering.isPreludeFunction(callee, "readDirectory"), lowering.isPreludeFunction(callee, "fileStatus"):
		// By what it is, not what it's called here: an import may name it anything.
		name := "fileStatus"
		if lowering.isPreludeFunction(callee, "readDirectory") {
			name = "readDirectory"
		}
		if len(arguments) != 1 {
			return nil, true, errors.New("lower: " +
				lowering.program.Where(node) +
				": " +
				name +
				" takes one path, and the checker let another count through")
		}
		path, expressionError := lowering.expression(arguments[0])
		if expressionError != nil {
			return nil, true, expressionError
		}
		if path.Type() != ir.String {
			return nil, true, errors.New("lower: " +
				lowering.program.Where(arguments[0]) +
				": " +
				name +
				"'s path isn't a string, and the checker let it through")
		}
		if name == "readDirectory" {
			return ir.ReadDirectory{Path: path}, true, nil
		}
		return ir.FileStatus{Path: path}, true, nil
	case lowering.isPreludeFunction(callee, "programArguments"):
		if len(arguments) != 0 {
			return nil, true, errors.New("lower: " +
				lowering.program.Where(node) +
				": programArguments takes nothing, and the checker let arguments through")
		}
		return ir.ProgramArguments{}, true, nil
	}
	return nil, false, nil
}
