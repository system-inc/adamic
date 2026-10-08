package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Intrinsic aliases are specialized at their uses. Their token is deliberately opaque: it may
// only be copied to another const, called with a proven receiver, or used by a supported callback
// adapter. It cannot escape into a slot whose callable ABI would erase the receiver's type.
type libraryMethod struct {
	family, name string
	source       *ast.Node
}

func (l *lowering) libraryMethod(node *ast.Node, seen map[*ast.Symbol]bool) (libraryMethod, bool) {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol == nil || seen[symbol] || len(symbol.Declarations) != 1 {
			return libraryMethod{}, false
		}
		seen[symbol] = true
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
			return libraryMethod{}, false
		}
		initializer := declaration.AsVariableDeclaration().Initializer
		if initializer == nil {
			return libraryMethod{}, false
		}
		return l.libraryMethod(initializer, seen)
	}
	if node.Kind != ast.KindPropertyAccessExpression || !l.libraryMember(node) || len(l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindCall)) == 0 {
		return libraryMethod{}, false
	}
	access := node.AsPropertyAccessExpression()
	if access.QuestionDotToken != nil || node.Flags&ast.NodeFlagsOptionalChain != 0 {
		return libraryMethod{}, false
	}
	if l.isLibraryGlobal(access.Expression, "Math") && node.Name().Text() == "random" {
		return libraryMethod{}, false
	}
	receiver := ast.SkipParentheses(access.Expression)
	for _, family := range []string{"Object"} {
		if l.isLibraryGlobal(receiver, family) {
			if node.Name().Text() == "keys" || node.Name().Text() == "values" || node.Name().Text() == "entries" {
				return libraryMethod{family, node.Name().Text(), node}, true
			}
		}
		if receiver.Kind == ast.KindPropertyAccessExpression && receiver.Name().Text() == "prototype" && l.isLibraryGlobal(receiver.AsPropertyAccessExpression().Expression, family) {
			if node.Name().Text() == "hasOwnProperty" {
				return libraryMethod{family + ".prototype", node.Name().Text(), node}, true
			}
		}
	}
	return libraryMethod{}, false
}

func constMethodInitializer(node *ast.Node) bool {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	return parent != nil && parent.Kind == ast.KindVariableDeclaration && parent.AsVariableDeclaration().Initializer == outer && parent.Parent != nil && parent.Parent.Flags&ast.NodeFlagsConst != 0
}

func (l *lowering) libraryMethodReadAllowed(node *ast.Node) bool {
	if _, known := l.libraryMethod(node, map[*ast.Symbol]bool{}); !known {
		return false
	}
	if constMethodInitializer(node) {
		return true
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	if parent != nil && parent.Kind == ast.KindPropertyAccessExpression && (parent.Name().Text() == "call" || parent.Name().Text() == "apply" || parent.Name().Text() == "bind") && called(parent) {
		return true
	}
	return l.libraryMapArgument(node)
}

func (l *lowering) libraryMapArgument(node *ast.Node) bool {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	call := parent.AsCallExpression()
	if len(call.Arguments.Nodes) != 1 || call.Arguments.Nodes[0] != outer {
		return false
	}
	callee := ast.SkipParentheses(call.Expression)
	return callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "map" && l.libraryMember(callee)
}

func (l *lowering) libraryMethodValue(node *ast.Node) (ir.Expression, bool, error) {
	_, known := l.libraryMethod(node, map[*ast.Symbol]bool{})
	if !known {
		return nil, false, nil
	}
	// Ordinary direct calls and existing observations are dispatched elsewhere.
	if ast.IsIdentifier(node) && (l.isLibraryGlobal(node, "String") || l.isLibraryGlobal(node, "Number")) {
		parent := node.Parent
		for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
			parent = parent.Parent
		}
		if parent != nil && parent.Kind == ast.KindBinaryExpression {
			operator := parent.AsBinaryExpression().OperatorToken.Kind
			if operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken {
				return nil, false, nil
			}
		}
	}
	if called(node) {
		return nil, false, nil
	}
	if !constMethodInitializer(node) {
		return nil, true, l.notYet(node, "a library method value outside a const alias, typed call/apply, or supported map callback (its receiver and callable ABI are not proven); wrap the call in an arrow")
	}
	if ast.IsIdentifier(node) && !l.isLibraryGlobal(node, "String") && !l.isLibraryGlobal(node, "Number") && !l.isLibraryGlobal(node, "parseInt") && !l.isLibraryGlobal(node, "parseFloat") {
		return nil, false, nil
	}
	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "library_method_token", Closure: true, Returns: ir.Number, Body: []ir.Statement{ir.Return{Value: ir.NumberConstant{}}}})
	return ir.MakeClosure{Function: index}, true, nil
}

// sequence returns the last operand while keeping all preceding evaluations and their checks.
func (l *lowering) methodSequence(values []ir.Expression) ir.Expression {
	index := len(l.result.Functions)
	function := ir.Function{Name: "library_method_sequence", Returns: values[len(values)-1].Type()}
	for _, value := range values {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "argument", Type: value.Type(), Function: index})
		function.Parameters = append(function.Parameters, local)
	}
	function.Body = []ir.Statement{ir.Return{Value: ir.Read{Local: function.Parameters[len(function.Parameters)-1], Of: function.Returns}}}
	l.result.Functions = append(l.result.Functions, function)
	return ir.Call{Function: index, Arguments: values, Returns: function.Returns}
}

func (l *lowering) methodAliasRead(node *ast.Node) (ir.Expression, error) {
	if !ast.IsIdentifier(ast.SkipParentheses(node)) {
		return nil, nil
	}
	if local, known := l.local(ast.SkipParentheses(node)); known {
		return ir.Read{Local: local, Of: l.result.Locals[local].Type, Checked: l.checked(local)}, nil
	}
	return nil, nil
}

// Only Object's shape-preserving inspection aliases are carried over from the
// named-record worker. Other families retain main's existing dispatch and refusals.
func (l *lowering) libraryMethodCall(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	target := ast.SkipParentheses(call.Expression)
	mode := "direct"
	if target.Kind == ast.KindPropertyAccessExpression && (target.Name().Text() == "call" || target.Name().Text() == "apply") {
		mode = target.Name().Text()
		target = target.AsPropertyAccessExpression().Expression
	}
	method, known := l.libraryMethod(target, map[*ast.Symbol]bool{})
	if !known {
		return nil, false, nil
	}
	if mode == "direct" && !ast.IsIdentifier(target) {
		return nil, false, nil
	}
	if method.family == "Object.prototype" && mode == "direct" {
		return nil, false, nil
	}
	if hasSpread(node) {
		return nil, true, l.notYet(node, "spread into an Object inspection alias")
	}
	written := call.Arguments.Nodes
	var receiver *ast.Node
	if mode != "direct" {
		if len(written) == 0 {
			return nil, true, l.notYet(node, "Object inspection alias without a receiver")
		}
		receiver, written = written[0], written[1:]
	}
	if mode == "apply" {
		if len(written) != 1 || ast.SkipParentheses(written[0]).Kind != ast.KindArrayLiteralExpression {
			return nil, true, l.notYet(node, "Object inspection apply without a dense literal")
		}
		written = ast.SkipParentheses(written[0]).AsArrayLiteralExpression().Elements.Nodes
		for _, argument := range written {
			if argument.Kind == ast.KindSpreadElement || argument.Kind == ast.KindOmittedExpression {
				return nil, true, l.notYet(node, "Object inspection apply with holes or spread")
			}
		}
	}
	prefix := []ir.Expression{}
	if alias, err := l.methodAliasRead(target); err != nil {
		return nil, true, err
	} else if alias != nil {
		prefix = append(prefix, alias)
	}
	var value ir.Expression
	var handled bool
	var err error
	if method.family == "Object" {
		if receiver != nil {
			r, e := l.expression(receiver)
			if e != nil {
				return nil, true, e
			}
			if r.Type() == 0 {
				r = fit(r, ir.Object)
			}
			prefix = append(prefix, r)
		}
		value, handled, err = l.objectCallArguments(node, method.name, written)
	} else {
		if receiver == nil || len(written) != 1 {
			return nil, true, l.notYet(node, "detached hasOwnProperty without one receiver and key")
		}
		of, known := l.representation(l.checker.GetTypeAtLocation(receiver))
		if !known || l.includesUndefined(l.checker.GetTypeAtLocation(receiver)) {
			return nil, true, l.notYet(receiver, "hasOwnProperty receiver with unproven storage")
		}
		if of != ir.Object && of != ir.Record {
			value, handled, err = l.nonObjectOwnPropertyArguments(node, receiver, "hasOwnProperty", of, written)
		} else {
			if of == ir.Object {
				if reason := l.prototypeHazard(receiver, ""); reason != "" {
					return nil, true, l.notYet(receiver, "hasOwnProperty through an object view ("+reason+")")
				}
			}
			r, e := l.expression(receiver)
			if e != nil {
				return nil, true, e
			}
			k, e := l.recordKey(written[0], false)
			if e != nil {
				return nil, true, e
			}
			if of == ir.Record {
				slot, e := l.recordSlot(receiver)
				if e != nil {
					return nil, true, e
				}
				value = ir.RecordCall{Method: "hasOwn", Arguments: []ir.Expression{r, k}, Element: slot, Returns: ir.Boolean}
			} else {
				value = ir.HasOwn{Object: r, Key: k}
			}
			handled = true
		}
	}
	if !handled || err != nil {
		return value, handled, err
	}
	if len(prefix) != 0 {
		value = l.methodSequence(append(prefix, value))
	}
	return value, true, nil
}
