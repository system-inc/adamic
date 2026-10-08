package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Extracted from regexBuiltin at codex/regex-protocol ffb3ae0. All String entry
// points share its argument admission and RegExpCall dispatch; runtime protocols
// and callback collection remain the imported implementation.
func (l *lowering) stringRegExpMethod(node *ast.Node, value ir.Expression, name string, args []*ast.Node, provided []ir.Expression) (ir.Expression, bool, error) {
	if len(args) == 0 || !l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(args[0])), "RegExp") {
		return l.stringRegExpCreate(node, value, name, args, provided)
	}
	result := ir.Type(0)
	switch name {
	case "match", "split":
		result = ir.Array
	case "matchAll":
		result = ir.Object
	case "search":
		result = ir.Number
	case "replace", "replaceAll":
		result = ir.String
	default:
		return nil, false, nil
	}
	method := name
	arguments := provided
	if provided == nil {
		for _, arg := range args {
			v, err := l.expression(arg)
			if err != nil {
				return nil, true, err
			}
			arguments = append(arguments, v)
		}
	}
	if name == "replace" || name == "replaceAll" {
		if len(arguments) != 2 {
			return nil, true, l.notYet(node, "regex replacement arity")
		}
		if arguments[1].Type() == ir.Closure && l.regexReplacementCallback(args[1]) {
			method += "Callback"
		} else if arguments[1].Type() != ir.String {
			return nil, true, l.notYet(node, "regex replacement callback with unproved capture/index/input/group parameters or result type")
		}
	} else if name == "split" {
		if len(arguments) == 1 {
			arguments = append(arguments, ir.Undefined{})
		}
		undefined := false
		if len(arguments) == 2 {
			_, undefined = arguments[1].(ir.Undefined)
		}
		if len(arguments) != 2 || !undefined && arguments[1].Type() != ir.Number && arguments[1].Type() != ir.MaybeNumber {
			return nil, true, l.notYet(node, "regex split limit other than a number")
		}
		if !l.regexSplitLimitProven(args[0], arguments[1]) {
			return nil, true, l.notYet(node, "String RegExp split limit whose V8 Smi/HeapNumber tag is unproved")
		}
	} else if len(arguments) != 1 {
		return nil, true, l.notYet(node, "String RegExp protocol with extra arguments")
	}
	return ir.RegExpCall{Value: value, Arguments: arguments, Method: method, Returns: result}, true, nil
}

// match, matchAll and search create an intrinsic RegExp when no protocol object
// was supplied. Only a proven constant primitive pattern can use the existing
// compiler; object hooks and a dynamic ECMAScript pattern compiler remain refused.
func (l *lowering) stringRegExpCreate(node *ast.Node, value ir.Expression, name string, args []*ast.Node, provided []ir.Expression) (ir.Expression, bool, error) {
	result := ir.Array
	flags := ""
	switch name {
	case "match":
	case "search":
		result = ir.Number
	case "matchAll":
		result, flags = ir.Object, "g"
	default:
		return nil, false, nil
	}
	if len(args) > 1 {
		return nil, true, l.notYet(node, "String RegExpCreate with extra arguments")
	}
	pattern := ""
	var evaluated []ir.Expression
	if len(args) == 1 {
		if !l.constantUndefined(args[0], 0) {
			var proven bool
			pattern, proven = l.constantPattern(args[0], 0)
			if !proven {
				return nil, true, l.notYet(node, "String RegExpCreate with a nonconstant or observable pattern")
			}
		}
		if provided != nil {
			evaluated = provided
		} else {
			argument, err := l.expression(args[0])
			if err != nil {
				return nil, true, err
			}
			evaluated = []ir.Expression{argument}
		}
	}
	regex, err := l.regexCompiled(node, pattern, flags, evaluated)
	if err != nil {
		return nil, true, err
	}
	return ir.RegExpCall{Value: value, Arguments: []ir.Expression{regex}, Method: name, Returns: result}, true, nil
}
