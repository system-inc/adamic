package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Observing an intrinsic's descriptors does not detach its this or materialize
// a mutable prototype. Every root is checked against the bundled declaration.
var dateMethodLengths = map[string]int{
	"toString": 0, "toDateString": 0, "toTimeString": 0, "toLocaleString": 0, "toLocaleDateString": 0, "toLocaleTimeString": 0,
	"valueOf": 0, "getTime": 0, "getFullYear": 0, "getUTCFullYear": 0, "getMonth": 0, "getUTCMonth": 0, "getDate": 0, "getUTCDate": 0,
	"getDay": 0, "getUTCDay": 0, "getHours": 0, "getUTCHours": 0, "getMinutes": 0, "getUTCMinutes": 0, "getSeconds": 0, "getUTCSeconds": 0,
	"getMilliseconds": 0, "getUTCMilliseconds": 0, "getTimezoneOffset": 0, "getYear": 0,
	"setTime": 1, "setMilliseconds": 1, "setUTCMilliseconds": 1, "setSeconds": 2, "setUTCSeconds": 2, "setMinutes": 3, "setUTCMinutes": 3,
	"setHours": 4, "setUTCHours": 4, "setDate": 1, "setUTCDate": 1, "setMonth": 2, "setUTCMonth": 2, "setFullYear": 3, "setUTCFullYear": 3, "setYear": 1,
	"toUTCString": 0, "toGMTString": 0, "toISOString": 0, "toJSON": 1,
}

func (l *lowering) dateIntrinsic(node *ast.Node) (string, int, bool) {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression {
		return "", 0, false
	}
	access := node.AsPropertyAccessExpression()
	name := node.Name().Text()
	if l.isLibraryGlobal(access.Expression, "Date") {
		switch name {
		case "parse":
			return name, 1, true
		case "UTC":
			return name, 7, true
		}
	}
	if l.datePrototype(access.Expression) && l.libraryMember(node) {
		length, ok := dateMethodLengths[name]
		return name, length, ok
	}
	return "", 0, false
}
func (l *lowering) libraryDateObservedMethod(node *ast.Node) bool {
	if _, _, ok := l.dateIntrinsic(node); !ok {
		return false
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	switch parent.Name().Text() {
	case "name", "length":
		return true
	case "hasOwnProperty":
		return called(parent)
	}
	return false
}
func (l *lowering) libraryDateProperty(node *ast.Node) (ir.Expression, bool) {
	access := node.AsPropertyAccessExpression()
	name := node.Name().Text()
	if l.isLibraryGlobal(access.Expression, "Date") {
		if name == "length" {
			return ir.NumberConstant{Value: 7}, true
		}
		if name == "name" {
			return ir.StringConstant{Index: l.constant("Date")}, true
		}
	}
	method, length, known := l.dateIntrinsic(access.Expression)
	if !known {
		return nil, false
	}
	if name == "length" {
		return ir.NumberConstant{Value: float64(length)}, true
	}
	if name == "name" {
		if method == "toGMTString" {
			method = "toUTCString"
		}
		return ir.StringConstant{Index: l.constant(method)}, true
	}
	return nil, false
}
func (l *lowering) libraryDateMetadataCall(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "hasOwnProperty" {
		return nil, false, nil
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	target := -1
	if l.isLibraryGlobal(receiver, "Date") {
		target = 0
	} else if l.datePrototype(receiver) {
		target = 1
	} else if _, _, ok := l.dateIntrinsic(receiver); ok {
		target = 2
	}
	if target < 0 {
		return nil, false, nil
	}
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) != 1 || hasSpread(node) {
		return nil, true, l.notYet(node, "Date.hasOwnProperty with other than one string key")
	}
	key, err := l.expression(written[0])
	if err != nil {
		return nil, true, err
	}
	if key.Type() != ir.String || l.includesNull(l.checker.GetTypeAtLocation(written[0])) || l.includesUndefined(l.checker.GetTypeAtLocation(written[0])) {
		return nil, true, l.notYet(node, "Date.hasOwnProperty with a non-string key")
	}
	return ir.DateCall{Method: "hasOwnProperty", Arguments: []ir.Expression{ir.NumberConstant{Value: float64(target)}, key}, Returns: ir.Boolean}, true, nil
}
