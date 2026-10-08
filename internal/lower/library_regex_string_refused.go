package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

var regexStringMethodLengths = map[string]float64{"split": 2, "match": 1, "matchAll": 1, "search": 1, "replace": 2, "replaceAll": 2}

func (l *lowering) regexStringMethodName(node *ast.Node) string {
	name, intrinsic := l.stringPrototypeMethod(node)
	if _, known := regexStringMethodLengths[name]; intrinsic && known {
		return name
	}
	return ""
}

// Inspecting an intrinsic's own metadata does not detach it. Mutation and every other value use
// stay refused. These builtins have nonenumerable name/length and no constructor prototype.
func (l *lowering) regexStringMetadataUse(node *ast.Node) bool {
	if l.regexStringMethodName(node) == "" {
		return false
	}
	outer := stringBoxOuter(node)
	parent := outer.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindForInStatement {
		return parent.AsForInOrOfStatement().Expression == outer
	}
	if parent.Kind != ast.KindPropertyAccessExpression || parent.AsPropertyAccessExpression().Expression != outer || parent.AsPropertyAccessExpression().QuestionDotToken != nil || stringBoxWritten(parent) {
		return false
	}
	switch parent.Name().Text() {
	case "length", "name", "prototype":
		return true
	case "hasOwnProperty", "propertyIsEnumerable":
		return called(parent)
	}
	return false
}

func (l *lowering) regexStringMethodProperty(node *ast.Node) (ir.Expression, bool) {
	access := node.AsPropertyAccessExpression()
	name := l.regexStringMethodName(access.Expression)
	if name == "" || access.QuestionDotToken != nil {
		return nil, false
	}
	switch node.Name().Text() {
	case "length":
		return ir.NumberConstant{Value: regexStringMethodLengths[name]}, true
	case "name":
		return ir.StringConstant{Index: l.constant(name)}, true
	case "prototype":
		return ir.Undefined{}, true
	}
	return nil, false
}

func (l *lowering) regexStringRefusedCall(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	receiver, method := callee.AsPropertyAccessExpression().Expression, callee.Name().Text()
	args := node.AsCallExpression().Arguments.Nodes
	if name := l.regexStringMethodName(receiver); name != "" && (method == "hasOwnProperty" || method == "propertyIsEnumerable") {
		if len(args) != 1 || hasSpread(node) {
			return nil, true, l.notYet(node, "String intrinsic metadata with these arguments")
		}
		key, err := l.expression(args[0])
		if err != nil {
			return nil, true, err
		}
		if key.Type() != ir.String {
			return nil, true, l.notYet(node, "String intrinsic metadata with a non-string key")
		}
		function, reads := l.stringHelper("regex_metadata", []ir.Expression{key})
		result := ir.Expression(ir.BooleanConstant{Value: false})
		if method == "hasOwnProperty" {
			result = ir.Binary{Operator: ir.Or, Left: ir.Binary{Operator: ir.Equal, Left: reads[0], Right: ir.StringConstant{Index: l.constant("length")}}, Right: ir.Binary{Operator: ir.Equal, Left: reads[0], Right: ir.StringConstant{Index: l.constant("name")}}}
		}
		l.result.Functions[function].Returns = ir.Boolean
		l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: result}}
		return ir.Call{Function: function, Arguments: []ir.Expression{key}, Returns: ir.Boolean}, true, nil
	}
	if method == "call" {
		if name, known := l.stringPrototypeMethod(receiver); known && name == "search" && len(args) > 0 && l.libraryStringBoxOrigin(args[0]) != nil {
			receiver, method, args = args[0], name, args[1:]
		}
	}
	if l.libraryStringBoxOrigin(receiver) == nil {
		return nil, false, nil
	}
	value, err := l.libraryStringBoxSlot(receiver)
	if err != nil {
		return nil, true, err
	}
	if method == "toString" || method == "valueOf" {
		return l.libraryStringMethod(node, value, method, args)
	}
	if method != "search" || len(args) != 1 || hasSpread(node) {
		return nil, true, l.notYet(node, "String box method with these arguments")
	}
	var pattern ir.Expression
	if l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(args[0])), "RegExp") {
		if l.mayBeUndefined(args[0]) || l.includesNull(l.checker.GetTypeAtLocation(args[0])) {
			return nil, true, l.notYet(node, "String box search with a nullable RegExp")
		}
		pattern, err = l.expression(args[0])
	} else if text, constant := l.constantPattern(args[0], 0); constant {
		var evaluated ir.Expression
		evaluated, err = l.expression(args[0])
		if err == nil {
			pattern, err = l.compiledStringRegExp(node, text, "", []ir.Expression{evaluated})
		}
	} else {
		return nil, true, l.notYet(node, "String box search with an unproved string pattern")
	}
	if err != nil {
		return nil, true, err
	}
	return ir.RegExpCall{Value: value, Arguments: []ir.Expression{pattern}, Method: "search", Returns: ir.Number}, true, nil
}

// Function.prototype and Object.prototype have no enumerable own properties in Adamic's fixed
// intrinsic universe. Keep the normal loop IR (and unreachable body diagnostics) over an empty list.
func (l *lowering) regexStringMethodForIn(node *ast.Node) ([]ir.Statement, bool, error) {
	statement := node.AsForInOrOfStatement()
	if l.regexStringMethodName(statement.Expression) == "" {
		return nil, false, nil
	}
	initializer := statement.Initializer
	if initializer.Kind != ast.KindVariableDeclarationList || initializer.Flags&ast.NodeFlagsBlockScoped == 0 || len(initializer.AsVariableDeclarationList().Declarations.Nodes) != 1 {
		return nil, true, l.notYet(node, "String intrinsic enumeration without one declared name")
	}
	name := initializer.AsVariableDeclarationList().Declarations.Nodes[0].Name()
	if !ast.IsIdentifier(name) {
		return nil, true, l.notYet(node, "String intrinsic enumeration with destructuring")
	}
	local, err := l.declareLocal(name)
	if err != nil {
		return nil, true, err
	}
	body, err := l.statement(statement.Statement)
	if err != nil {
		return nil, true, err
	}
	return []ir.Statement{ir.ForOf{Local: local, Iterable: ir.ArrayLiteral{Element: ir.String}, Element: ir.String, Body: body}}, true, nil
}

// RegExp.prototype.toString follows V8's builtins-regexp.cc: slash, escaped source, slash, flags.
// Property overrides and nullable or structural RegExp views are refused by their existing guards.
func (l *lowering) regexStringConversion(node *ast.Node) (ir.Expression, bool, error) {
	if !l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node)), "RegExp") {
		return nil, false, nil
	}
	if l.mayBeUndefined(node) || l.includesNull(l.checker.GetTypeAtLocation(node)) {
		return nil, true, l.notYet(node, "String conversion of a nullable RegExp")
	}
	value, err := l.expression(node)
	if err != nil {
		return nil, true, err
	}
	function, reads := l.stringHelper("regex_to_string", []ir.Expression{value})
	slash := ir.StringConstant{Index: l.constant("/")}
	result := ir.Concat{Parts: []ir.Expression{slash, ir.Property{Object: reads[0], Name: "source", Of: ir.String}, slash, ir.Property{Object: reads[0], Name: "flags", Of: ir.String}}}
	l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: result}}
	return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: ir.String}, true, nil
}

// The string arm of + reuses the existing V8 number printer. Both operands are already evaluated
// once, in source order. Boolean and explicit undefined reuse their primitive spellings; other operands remain refused.
func (l *lowering) regexStringPrimitiveConcat(left, right ir.Expression) (ir.Expression, bool) {
	convert := func(value ir.Expression) ir.Expression {
		switch value.Type() {
		case ir.Number:
			return ir.NumberToString{Value: value}
		case ir.Boolean:
			return ir.BooleanToString{Value: value}
		}
		if _, missing := value.(ir.Undefined); missing {
			return ir.StringConstant{Index: l.constant("undefined")}
		}
		return nil
	}
	if left.Type() == ir.String {
		if converted := convert(right); converted != nil {
			return ir.Concat{Parts: []ir.Expression{left, converted}}, true
		}
	}
	if right.Type() == ir.String {
		if converted := convert(left); converted != nil {
			return ir.Concat{Parts: []ir.Expression{converted, right}}, true
		}
	}
	return nil, false
}
