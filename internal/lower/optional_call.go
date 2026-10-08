package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// lowerOptionalCall keeps the argument work inside a branch of an immediately
// called capture helper. Existing IR supplies evaluation order and ownership in
// both backends; no argument runs until the saved callee or receiver is present.
func (l *lowering) lowerOptionalCall(node *ast.Node, discarded bool) (ir.Expression, bool, error) {
	if node.Kind != ast.KindCallExpression || node.Flags&ast.NodeFlagsOptionalChain == 0 {
		return nil, false, nil
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind == ast.KindPropertyAccessExpression {
		property := callee.AsPropertyAccessExpression()
		if property.QuestionDotToken == nil && property.Expression.Flags&ast.NodeFlagsOptionalChain != 0 {
			return nil, true, l.notYet(node, "an optional call continuing through an unguarded intermediate property")
		}
	}
	explicit := call.QuestionDotToken != nil
	library := false
	if callee.Kind == ast.KindPropertyAccessExpression {
		symbol := l.checker.GetSymbolAtLocation(callee)
		library = symbol != nil && len(symbol.Declarations) > 0 && load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0]))
	}
	if !explicit && !library {
		return nil, false, nil
	}
	if explicit && !library {
		value, err := l.optionalCallable(node, discarded)
		return value, true, err
	}
	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "optional_call", Closure: true})
	outer := l.functionIndex
	l.functionIndex = index
	l.closures = append(l.closures, index)
	defer func() { l.functionIndex = outer; l.closures = l.closures[:len(l.closures)-1] }()
	var saved, value ir.Expression
	var err error
	held := len(l.result.Locals)
	read := ir.Read{Local: held}
	if library {
		var handled bool
		value, handled, err = l.optionalMapCall(node)
		if !handled {
			value, handled, err = l.builtin(node)
		}
		if err != nil {
			return nil, true, err
		}
		if !handled {
			return nil, true, l.notYet(node, "an optional library call without an intrinsic")
		}
		if call, ok := value.(ir.Call); ok {
			targets := l.result.CallTargets(call)
			if len(targets) != 1 || !strings.HasPrefix(l.result.Functions[targets[0]].Name, "library_string_") || len(call.Arguments) == 0 || call.Arguments[0].Type() != ir.String {
				return nil, true, l.notYet(node, "an optional library call without one proven receiver operand")
			}
		}
		saved, value = optionalIntrinsicReceiver(value, &read)
		if saved == nil {
			return nil, true, l.notYet(node, "an optional library call whose receiver is not a single intrinsic operand")
		}
	}
	if !saved.Type().IsReference() {
		return nil, true, l.notYet(node, "an optional receiver not held as a reference")
	}
	// A helper's arguments may create locals while lowering. Allocate the saved
	// receiver after that work, and point every replacement at the final local.
	final := len(l.result.Locals)
	replacement := ir.Read{Local: final, Of: saved.Type()}
	_, value = optionalIntrinsicReceiver(value, &replacement)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "optional_callee", Type: saved.Type(), Function: index})
	resultType := ir.Object
	if value.Type() != 0 {
		if proven, known := l.representation(l.checker.GetTypeAtLocation(node)); known {
			resultType = proven
		} else {
			return nil, true, l.notYet(node, "an optional call result without a representation")
		}
	} else if !discarded {
		return nil, true, l.notYet(node, "a void optional call used as a value")
	}
	absent := fit(ir.Undefined{}, resultType)
	body := []ir.Statement{ir.Declare{Local: final, Value: saved}, ir.If{Condition: ir.IsUndefined{Value: ir.Read{Local: final, Of: saved.Type()}}, Then: []ir.Statement{ir.Return{Value: absent}}}}
	if value.Type() == 0 {
		body = append(body, ir.Evaluate{Value: value}, ir.Return{Value: absent})
	} else {
		body = append(body, ir.Return{Value: fit(value, resultType)})
	}
	l.result.Functions[index].Returns = resultType
	l.result.Functions[index].Body = body
	return ir.CallClosure{Closure: ir.MakeClosure{Function: index}, Returns: resultType}, true, nil
}

// Only these intrinsics have one receiver operand. Unknown shapes stay stopped.
func optionalIntrinsicReceiver(value ir.Expression, replacement *ir.Read) (ir.Expression, ir.Expression) {
	replace := func(receiver ir.Expression) ir.Expression { replacement.Of = receiver.Type(); return *replacement }
	switch v := value.(type) {
	case ir.Call:
		old := v.Arguments[0]
		v.Arguments = append([]ir.Expression(nil), v.Arguments...)
		v.Arguments[0] = replace(old)
		return old, v
	case ir.MapGet:
		old := v.Map
		v.Map = replace(old)
		return old, v
	case ir.MapSet:
		old := v.Map
		v.Map = replace(old)
		return old, v
	case ir.MapHas:
		old := v.Map
		v.Map = replace(old)
		return old, v
	case ir.MapDelete:
		old := v.Map
		v.Map = replace(old)
		return old, v
	case ir.MapClear:
		old := v.Map
		v.Map = replace(old)
		return old, v
	case ir.MapForEach:
		old := v.Map
		v.Map = replace(old)
		return old, v
	case ir.SetAdd:
		old := v.Set
		v.Set = replace(old)
		return old, v
	case ir.ArrayMap:
		old := v.Array
		v.Array = replace(old)
		return old, v
	case ir.ArrayVisit:
		old := v.Array
		v.Array = replace(old)
		return old, v
	case ir.ArraySort:
		old := v.Array
		v.Array = replace(old)
		return old, v
	case ir.ArrayPush:
		old := v.Array
		v.Array = replace(old)
		return old, v
	case ir.ArrayPop:
		old := v.Array
		v.Array = replace(old)
		return old, v
	case ir.ArraySearch:
		old := v.Array
		v.Array = replace(old)
		return old, v
	case ir.ArrayJoin:
		old := v.Array
		v.Array = replace(old)
		return old, v
	case ir.ArraySlice:
		old := v.Array
		v.Array = replace(old)
		return old, v
	case ir.StringCall:
		old := v.Value
		v.Value = replace(old)
		return old, v
	case ir.Trim:
		old := v.Value
		v.Value = replace(old)
		return old, v
	case ir.CharCodeAt:
		old := v.Value
		v.Value = replace(old)
		return old, v
	}
	return nil, value
}

// A nullable Map's type arguments belong to its present member. Keep the
// ordinary Map helper unchanged for the other workers using it tonight.
func (l *lowering) optionalMapCall(node *ast.Node) (ir.Expression, bool, error) {
	access := ast.SkipParentheses(node.AsCallExpression().Expression).AsPropertyAccessExpression()
	receiver := access.Expression
	proven := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver))
	set := l.isLibraryType(proven, "Set", "ReadonlySet")
	if !set && !l.isLibraryType(proven, "Map", "ReadonlyMap") {
		return nil, false, nil
	}
	name := access.Name().Text()
	if name != "get" && name != "set" && name != "has" && name != "delete" && name != "clear" && name != "forEach" && !(set && name == "add") {
		return nil, false, nil
	}
	types := l.typeArguments(proven)
	wantTypes := 2
	if set {
		wantTypes = 1
	}
	if len(types) != wantTypes {
		return nil, true, l.notYet(node, "an optional collection without concrete element types")
	}
	key, keyKnown := l.representation(types[0])
	mapped, valueKnown := key, keyKnown
	if !set {
		mapped, valueKnown = l.kept(types[1])
	}
	if !keyKnown || !keyable(key) || !valueKnown || slotless(mapped) {
		return nil, true, l.notYet(node, "an optional collection without representable slots")
	}
	object, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	values := []ir.Expression{}
	for _, arg := range node.AsCallExpression().Arguments.Nodes {
		v, err := l.expression(arg)
		if err != nil {
			return nil, true, err
		}
		values = append(values, v)
	}
	if name == "forEach" {
		if len(values) != 1 || values[0].Type() != ir.Closure {
			return nil, true, l.notYet(node, "optional forEach without one callback")
		}
		signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node.AsCallExpression().Arguments.Nodes[0]), checker.SignatureKindCall)
		if len(signatures) != 1 {
			return nil, true, l.notYet(node, "optional forEach without one callback signature")
		}
		returns := ir.Type(0)
		if result := l.checker.GetReturnTypeOfSignature(signatures[0]); result.Flags()&checker.TypeFlagsVoid == 0 {
			var known bool
			returns, known = l.representation(result)
			if !known || slotless(returns) {
				return nil, true, l.notYet(node, "optional forEach without a represented callback result")
			}
		}
		return ir.MapForEach{Map: object, Callback: values[0], Key: key, Value: mapped, Set: set, Returns: returns}, true, nil
	}
	if name == "clear" {
		if len(values) != 0 {
			return nil, true, l.notYet(node, "optional Map clear with arguments")
		}
		return ir.MapClear{Map: object}, true, nil
	}
	want := 1
	if name == "set" {
		want = 2
	}
	if len(values) != want {
		return nil, true, l.notYet(node, "optional Map call with a different argument count")
	}
	values[0] = fit(values[0], key)
	if name == "set" {
		values[1] = fit(values[1], mapped)
	}
	if values[0].Type() != key || (name == "set" && values[1].Type() != mapped) {
		return nil, true, l.notYet(node, "optional Map call with unrepresented arguments")
	}
	switch name {
	case "add":
		return ir.SetAdd{Set: object, Value: values[0], Element: key, Site: l.writeSite(receiver)}, true, nil
	case "get":
		return ir.MapGet{Map: object, Key: values[0], KeyType: key, ValueType: mapped}, true, nil
	case "has":
		return ir.MapHas{Map: object, Key: values[0], KeyType: key}, true, nil
	case "set":
		return ir.MapSet{Map: object, Key: values[0], Value: values[1], KeyType: key, ValueType: mapped, Site: l.writeSite(receiver)}, true, nil
	}
	return ir.MapDelete{Map: object, Key: values[0], KeyType: key}, true, nil
}

// optionalCallable keeps a method's receiver attached to its callee. Returns
// describes the present call; CallClosure.Optional adds the missing result.
func (l *lowering) optionalCallable(node *ast.Node, discarded bool) (ir.Expression, error) {
	call := node.AsCallExpression()
	proven := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(call.Expression))
	signatures := l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)
	if len(signatures) != 1 || l.censusNeverRestSignature(signatures[0]) {
		return nil, l.notYet(node, "an optional call without one concrete callable signature")
	}
	signature := signatures[0]
	for _, parameter := range signature.Parameters() {
		for _, declaration := range parameter.Declarations {
			if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil {
				return nil, l.notYet(node, "a rest parameter in an optional callable")
			}
		}
	}
	var callee ir.Expression
	var err error
	access := ast.SkipParentheses(call.Expression)
	if access.Kind == ast.KindPropertyAccessExpression {
		property := access.AsPropertyAccessExpression()
		symbol := l.checker.GetSymbolAtLocation(access)
		if symbol != nil && symbol.Flags&(ast.SymbolFlagsGetAccessor|ast.SymbolFlagsSetAccessor) != 0 {
			return nil, l.notYet(node, "an optional callee read through an accessor")
		}
		if err := l.erasedMethodField(access, l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(property.Expression)), property.Name().Text()); err != nil {
			return nil, err
		}
		receiver, failure := l.expression(property.Expression)
		if failure != nil {
			return nil, failure
		}
		if receiver.Type() != ir.Object {
			return nil, l.notYet(node, "an optional callable receiver not held as an object")
		}
		callee = ir.Property{Object: receiver, Name: property.Name().Text(), Of: ir.Closure, Method: true, Optional: property.QuestionDotToken != nil}
	} else {
		callee, err = l.expression(call.Expression)
		if err != nil {
			return nil, err
		}
		if callee.Type() != ir.Closure {
			return nil, l.notYet(node, "an optional callee not held as a closure")
		}
	}
	arguments := []ir.Expression{}
	for position, argument := range call.Arguments.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			return nil, l.notYet(argument, "a spread argument to an optional callable")
		}
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, err
		}
		if position < len(signature.Parameters()) {
			if takes, known := l.censusCallableParameter(signature.Parameters()[position]); known {
				lowered = fit(lowered, takes)
			}
		}
		if censusCallableSlotless(lowered.Type()) {
			return nil, l.notYet(argument, "an optional callable argument without a closure slot")
		}
		arguments = append(arguments, lowered)
	}
	result := l.checker.GetReturnTypeOfSignature(signature)
	returns := ir.Type(0)
	if result.Flags()&checker.TypeFlagsVoid == 0 {
		var known bool
		returns, known = l.representation(result)
		if !known || censusCallableSlotless(returns) {
			return nil, l.notYet(node, "an optional callable result without a closure slot")
		}
	} else if !discarded {
		return nil, l.notYet(node, "a void optional call used as a value")
	}
	return ir.CallClosure{Closure: callee, Arguments: arguments, Returns: returns, Optional: true}, nil
}
