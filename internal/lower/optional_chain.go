package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Continuous chains keep each guard outside the entire remaining continuation.
// Parentheses terminate collection. An undefined value returned by a present
// call is not a short circuit and must reach the next ordinary access.
func (l *lowering) optionalChain(node *ast.Node) (ir.Expression, bool, error) {
	if node.Flags&ast.NodeFlagsOptionalChain == 0 {
		return nil, false, nil
	}
	// Optional join already owns its receiver and lazy separator.
	if node.Kind == ast.KindCallExpression && l.isOptionalJoin(node) {
		return nil, false, nil
	}
	// Preserve the separately tested single callable and method invocation path.
	if node.Kind == ast.KindCallExpression {
		c := node.AsCallExpression()
		if c.QuestionDotToken != nil && c.Expression.Flags&ast.NodeFlagsOptionalChain == 0 {
			return nil, false, nil
		}
	}
	var steps []*ast.Node
	var collect func(*ast.Node) *ast.Node
	collect = func(n *ast.Node) *ast.Node {
		if n.Flags&ast.NodeFlagsOptionalChain == 0 {
			return n
		}
		var before *ast.Node
		switch n.Kind {
		case ast.KindPropertyAccessExpression:
			before = n.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			before = n.AsElementAccessExpression().Expression
		case ast.KindCallExpression:
			before = n.AsCallExpression().Expression
		default:
			return n
		}
		base := collect(before)
		steps = append(steps, n)
		return base
	}
	base := collect(node)
	if len(steps) == 0 {
		return nil, false, nil
	}
	// Preserve the existing constant tuple indexing representation.
	if len(steps) == 1 && node.Kind == ast.KindElementAccessExpression {
		t := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(base))
		if r, known := l.representation(t); known && r == ir.Object {
			return nil, false, nil
		}
	}
	// Preserve existing one-step reads, including RegExp properties and full
	// field contracts. Only collection size needs the new guard here.
	if len(steps) == 1 && node.Kind == ast.KindPropertyAccessExpression {
		t := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(base))
		if r, known := l.representation(t); !known || r != ir.Map || node.Name().Text() != "size" {
			return nil, false, nil
		}
	}
	discarded := false
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	discarded = at.Parent != nil && at.Parent.Kind == ast.KindExpressionStatement
	of := ir.Object
	var err error
	if !discarded {
		of, err = l.typeOf(node)
		if err != nil {
			return nil, true, err
		}
		if !of.IsReference() && !of.IsMaybe() {
			return nil, true, l.notYet(node, "an optional chain result without an undefined representation")
		}
	}
	value, err := l.expression(base)
	if err != nil {
		return nil, true, err
	}
	var next func(int, ir.Expression) (ir.Expression, error)
	finish := func(v ir.Expression) ir.Expression {
		if discarded || v.Type() == 0 {
			return ir.Effects{Body: []ir.Statement{ir.Evaluate{Value: v}}, Result: fit(ir.Undefined{}, of)}
		}
		return fit(v, of)
	}
	next = func(i int, current ir.Expression) (ir.Expression, error) {
		if i == len(steps) {
			return finish(current), nil
		}
		step := steps[i]
		optional := false
		switch step.Kind {
		case ast.KindPropertyAccessExpression:
			optional = step.AsPropertyAccessExpression().QuestionDotToken != nil
		case ast.KindElementAccessExpression:
			optional = step.AsElementAccessExpression().QuestionDotToken != nil
		case ast.KindCallExpression:
			optional = step.AsCallExpression().QuestionDotToken != nil
		}
		apply := func(saved ir.Expression) (ir.Expression, error) {
			var result ir.Expression
			var err error
			advance := 1
			switch step.Kind {
			case ast.KindPropertyAccessExpression:
				// A method's callee and receiver stay together; never read a prototype
				// method into an ordinary own-field value.
				if i+1 < len(steps) && steps[i+1].Kind == ast.KindCallExpression {
					call := steps[i+1]

					result, err = l.chainMethodCall(step, call, saved, discarded && i+2 == len(steps), of)
					advance = 2
					if err == nil && call.AsCallExpression().QuestionDotToken != nil && i+2 < len(steps) {
						selected := result.(ir.CallClosure)
						if selected.Returns == 0 {
							return nil, l.notYet(call, "a void call inside an optional chain")
						}
						selected.OptionalResult = ir.Maybe(selected.Returns)
						presence := l.chainLocal(ir.Boolean)
						selected.OptionalPresent = presence + 1
						held := l.chainLocal(selected.Type())
						continuation, err := next(i+2, ir.Read{Local: held, Of: selected.Type()})
						if err != nil {
							return nil, err
						}
						return ir.Effects{Body: []ir.Statement{
							ir.Declare{Local: presence, Value: ir.BooleanConstant{Value: false}},
							ir.Declare{Local: held, Value: selected},
						}, Result: ir.Conditional{Condition: ir.Read{Local: presence, Of: ir.Boolean}, WhenTrue: continuation, WhenNot: fit(ir.Undefined{}, of), Of: of}}, nil
					}
				} else {
					result, err = l.chainProperty(step, saved)
				}
			case ast.KindElementAccessExpression:
				result, err = l.chainElement(step, saved)
			case ast.KindCallExpression:
				result, err = l.callClosureOn(step, saved)
			}
			if err != nil {
				return nil, err
			}
			if result.Type() == 0 && i+advance != len(steps) {
				return nil, l.notYet(step, "a void call inside an optional chain")
			}
			return next(i+advance, result)
		}
		// Bound method optional calls have their own selection guard. All other
		// guards snapshot their operand and nest the full continuation inside it.
		if optional {
			if !current.Type().IsReference() || current.Type() == ir.Union {
				return nil, l.notYet(step, "an optional chain guard without a reference representation")
			}
			local := l.chainLocal(current.Type())
			read := ir.Read{Local: local, Of: current.Type()}
			present, err := apply(read)
			if err != nil {
				return nil, err
			}
			return ir.Effects{Body: []ir.Statement{ir.Declare{Local: local, Value: current}}, Result: ir.Conditional{
				Condition: ir.Unary{Operator: ir.Not, Operand: ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: read}, Right: ir.IsNull{Value: read}}},
				WhenTrue:  present, WhenNot: fit(ir.Undefined{}, of), Of: of,
			}}, nil
		}
		return apply(current)
	}
	result, err := next(0, value)
	return result, true, err
}

func (l *lowering) chainLocal(t ir.Type) int {
	function := -1
	if l.function != nil {
		function = l.functionIndex
	}
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "optional_chain", Type: t, Function: function, ExpressionAssigned: true})
	return local
}

func (l *lowering) chainProperty(node *ast.Node, object ir.Expression) (ir.Expression, error) {
	if err := l.staticProperty(node); err != nil {
		return nil, err
	}
	name := l.fieldName(node.Name())
	switch object.Type() {
	case ir.Array, ir.Uint8Array, ir.Int32Array, ir.Float64Array:
		if name == "length" {
			return ir.Length{Array: object}, nil
		}
	case ir.String:
		if name == "length" {
			return ir.StringLength{Value: object}, nil
		}
	case ir.Map:
		if name == "size" {
			return ir.MapSize{Map: object}, nil
		}
	case ir.Object:
		if l.inheritedLibraryMember(node) {
			return nil, l.prototypeRead(node, name)
		}
		if err := l.erasedLiteralMethod(node); err != nil {
			return nil, err
		}
		symbol := l.checker.GetSymbolAtLocation(node.Name())
		if symbol == nil || accessorSymbol(symbol) || l.accessorNames[node.Name().Text()] {
			return nil, l.notYet(node, "an optional chain field without a represented own slot")
		}
		stored, known := l.representation(l.checker.GetTypeOfSymbol(symbol))
		if !known || stored == ir.Union || stored == ir.Weak || censusFieldSlotless(stored) {
			return nil, l.notYet(node, "an optional chain field without a supported stored representation")
		}
		// The chain result may be widened, but the physical own slot is not.
		return l.defined(node, l.readObjectField(node, ir.Property{Object: object, Name: name, Of: stored, Class: l.classOf(node)})), nil
	}
	return nil, l.notYet(node, "an optional chain property without a represented receiver")
}

func (l *lowering) chainElement(node *ast.Node, object ir.Expression) (ir.Expression, error) {
	access := node.AsElementAccessExpression()
	index, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, err
	}
	if index.Type() != ir.Number {
		return nil, l.notYet(node, "an optional chain index that is not a number")
	}
	switch object.Type() {
	case ir.Uint8Array, ir.Int32Array, ir.Float64Array:
		// Keep main's typed-buffer read contract while sharing the saved receiver.
		read := ir.Expression(ir.ArrayIndex{Array: object, Index: index, Element: ir.Number})
		if held, _ := l.representation(l.checker.GetTypeAtLocation(node)); held == ir.Number && !l.acceptsUndefined(node) {
			read = ir.Unwrap{Value: read}
		}
		return read, nil
	case ir.String:
		return ir.StringIndex{Value: object, Index: index}, nil
	case ir.Array:
		element, err := l.elementType(access.Expression)
		if err != nil {
			return nil, err
		}
		return l.defined(node, ir.ArrayIndex{Array: object, Index: index, Element: element}), nil
	}
	return nil, l.notYet(node, "an optional chain index without array or string storage")
}

func (l *lowering) chainMethodCall(member, node *ast.Node, object ir.Expression, discarded bool, of ir.Type) (ir.Expression, error) {
	if l.libraryMember(member) {
		if node.AsCallExpression().QuestionDotToken != nil {
			return nil, l.notYet(node, "an optional library method value")
		}
		return l.chainIntrinsic(member, node, object)
	}
	if object.Type() != ir.Object {
		return nil, l.notYet(node, "an optional chain method without object storage")
	}
	if err := l.staticProperty(member); err != nil {
		return nil, err
	}
	if err := l.erasedLiteralMethod(member); err != nil {
		return nil, err
	}
	memberType := l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(member)))
	if stored, known := l.representation(memberType); !known || stored != ir.Closure || l.weakTarget(memberType) != nil {
		return nil, l.notYet(node, "an optional chain method without closure member storage")
	}
	if err := l.optionalCallableStorage(member); err != nil {
		return nil, err
	}
	symbol := l.memberSymbol(member)
	if symbol == nil || accessorSymbol(symbol) || l.accessorNames[member.Name().Text()] {
		return nil, l.notYet(node, "an optional chain callable accessor without represented selection")
	}
	name := l.fieldName(member.Name())
	if l.result.CheckedFields[name] || l.result.UninitializedFields[name] {
		return nil, l.notYet(node, "an optional method requiring a checked field contract")
	}
	property := ir.Property{Object: object, Name: name, Of: ir.Closure, Class: l.classOf(member), Method: true}
	value, err := l.callClosureOn(node, property)
	if err != nil {
		return nil, err
	}
	call := value.(ir.CallClosure)
	if node.AsCallExpression().QuestionDotToken != nil {
		if call.Returns == 0 && !discarded {
			return nil, l.notYet(node, "observing an optional void call's runtime result")
		}
		call.Optional = true
		if !discarded {
			call.OptionalResult = of
		}
	}
	return call, nil
}

// Intrinsics take the already saved receiver. Re-lowering its AST would
// duplicate effects and leave unreachable temporary ownership declarations.
func (l *lowering) chainIntrinsic(member, node *ast.Node, receiver ir.Expression) (ir.Expression, error) {
	name := member.Name().Text()
	written := node.AsCallExpression().Arguments.Nodes
	source := member.AsPropertyAccessExpression().Expression
	if receiver.Type() == ir.String {
		value, _, err := l.libraryStringMethod(node, receiver, name, written)
		return value, err
	}
	if receiver.Type() == ir.Array {
		element, err := l.elementType(source)
		if err != nil {
			return nil, err
		}
		if _, visit := visits[name]; visit {
			value, _, err := l.arrayVisit(node, receiver, element, name)
			return value, err
		}
		if name == "map" && len(written) == 1 {
			callback, known, err := l.libraryMapCallback(written[0], source, element)
			if !known {
				callback, err = l.expression(written[0])
			}
			if err != nil {
				return nil, err
			}
			if callback.Type() != ir.Closure {
				return nil, l.notYet(node, "an optional map callback without closure storage")
			}
			result, err := l.elementType(node)
			if err != nil {
				return nil, err
			}
			return ir.ArrayMap{Array: receiver, Callback: callback, Element: element, Result: result, CallbackType: int(l.concrete(l.checker.GetTypeAtLocation(written[0])).Id())}, nil
		}
		if name == "pop" && len(written) == 0 {
			return ir.ArrayPop{Array: receiver, Element: element}, nil
		}
	}
	if receiver.Type() == ir.Map && !l.isSet(source) {
		key, value, err := l.mapTypes(source)
		if err != nil {
			return nil, err
		}
		if name == "forEach" && len(written) == 1 {
			callback, err := l.expression(written[0])
			if err != nil {
				return nil, err
			}
			signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(written[0]), checker.SignatureKindCall)
			if callback.Type() != ir.Closure || len(signatures) != 1 {
				return nil, l.notYet(node, "an optional map visit without one represented callback signature")
			}
			var returns ir.Type
			result := l.checker.GetReturnTypeOfSignature(signatures[0])
			if result.Flags()&checker.TypeFlagsVoid == 0 {
				var known bool
				returns, known = l.representation(result)
				if !known || slotless(returns) {
					return nil, l.notYet(node, "an optional map visit with an unsupported callback result")
				}
			}
			return ir.MapForEach{Map: receiver, Callback: callback, Key: key, Value: value, Returns: returns, CallbackType: int(l.concrete(l.checker.GetTypeAtLocation(written[0])).Id())}, nil
		}
		if name == "clear" && len(written) == 0 {
			return ir.MapClear{Map: receiver}, nil
		}
		if name == "get" || name == "has" || name == "set" || name == "delete" {
			want := 1
			if name == "set" {
				want = 2
			}
			if len(written) != want {
				return nil, l.notYet(node, "an optional map call with unsupported arguments")
			}
			arguments, spread, err := l.callArguments(written)
			if err != nil {
				return nil, err
			}
			for _, expanded := range spread {
				if expanded {
					return nil, l.notYet(node, "a spread into an optional map intrinsic")
				}
			}
			arguments[0] = fit(arguments[0], key)
			if arguments[0].Type() != key {
				return nil, l.notYet(node, "an optional map key with a different representation")
			}
			switch name {
			case "get":
				return ir.MapGet{Map: receiver, Key: arguments[0], KeyType: key, ValueType: value}, nil
			case "has":
				return ir.MapHas{Map: receiver, Key: arguments[0], KeyType: key}, nil
			case "delete":
				return ir.MapDelete{Map: receiver, Key: arguments[0], KeyType: key}, nil
			case "set":
				arguments[1] = fit(arguments[1], value)
				if arguments[1].Type() != value {
					return nil, l.notYet(node, "an optional map value with a different representation")
				}
				return ir.MapSet{Map: receiver, Key: arguments[0], Value: arguments[1], KeyType: key, ValueType: value, Site: l.writeSite(source)}, nil
			}
		}
	}
	return nil, l.notYet(node, "an optional chain intrinsic without an explicit represented receiver")
}

// Parentheses end the short circuit. An ordinary consumer must observe its
// undefined result, including when the checker kept a narrowing across a call.
func (l *lowering) optionalChainConsumer(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	if ast.SkipParentheses(node).Flags&ast.NodeFlagsOptionalChain == 0 {
		return value, nil
	}
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	parent := at.Parent
	if parent == nil || parent.Flags&ast.NodeFlagsOptionalChain != 0 {
		return value, nil
	}
	if parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == at {
		if write := parent.Parent; write != nil && write.Kind == ast.KindBinaryExpression && write.AsBinaryExpression().Left == parent && write.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
			// The existing write checks after evaluating its right-hand side.
			return value, nil
		}
		if !value.Type().IsReference() {
			return nil, l.notYet(parent, "an ordinary consumer of a parenthesized optional scalar")
		}
		return ir.Defined{Value: value, Message: "TypeError: Cannot read properties of undefined (reading '" + parent.Name().Text() + "')"}, nil
	}
	if parent.Kind == ast.KindElementAccessExpression && parent.AsElementAccessExpression().Expression == at {
		if write := parent.Parent; write != nil && write.Kind == ast.KindBinaryExpression && write.AsBinaryExpression().Left == parent && write.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
			return nil, l.notYet(parent, "an indexed write through a parenthesized optional chain")
		}
		if !value.Type().IsReference() {
			return nil, l.notYet(parent, "an ordinary consumer of a parenthesized optional scalar")
		}
		key := ast.SkipParentheses(parent.AsElementAccessExpression().ArgumentExpression)
		if key.Kind != ast.KindNumericLiteral && key.Kind != ast.KindStringLiteral {
			return nil, l.notYet(parent, "a computed ordinary index after a parenthesized optional chain")
		}
		return ir.Defined{Value: value, Message: "TypeError: Cannot read properties of undefined (reading '" + key.Text() + "')"}, nil
	}
	if parent.Kind == ast.KindCallExpression && parent.AsCallExpression().Expression == at {
		return nil, l.notYet(parent, "a call through a parenthesized optional-chain result")
	}
	return value, nil
}
