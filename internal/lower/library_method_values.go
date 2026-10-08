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
		for _, name := range []string{"String", "Number", "parseInt", "parseFloat"} {
			if l.isLibraryGlobal(node, name) {
				return libraryMethod{"global", name, node}, true
			}
		}
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
	for _, family := range []string{"Array", "String", "Number", "Object", "Math"} {
		if l.isLibraryGlobal(receiver, family) {
			return libraryMethod{family, node.Name().Text(), node}, true
		}
		if receiver.Kind == ast.KindPropertyAccessExpression && receiver.Name().Text() == "prototype" && l.isLibraryGlobal(receiver.AsPropertyAccessExpression().Expression, family) {
			return libraryMethod{family + ".prototype", node.Name().Text(), node}, true
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
		if l.detachedOwnAlias(node) != nil {
			return ir.Read{Local: local, Of: ir.Boolean, Checked: l.checked(local)}, nil
		}
		return ir.Read{Local: local, Of: ir.Closure, Checked: l.checked(local)}, nil
	}
	return nil, nil
}

func (l *lowering) libraryMethodCall(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	target := callee
	mode := "direct"
	if callee.Kind == ast.KindPropertyAccessExpression && (callee.Name().Text() == "call" || callee.Name().Text() == "apply" || callee.Name().Text() == "bind") {
		mode = callee.Name().Text()
		target = callee.AsPropertyAccessExpression().Expression
	}
	method, known := l.libraryMethod(target, map[*ast.Symbol]bool{})
	if !known {
		return nil, false, nil
	}
	// Leave existing immediate primitive prototype calls on their established paths.
	if mode == "call" && method.source == ast.SkipParentheses(target) && (method.family == "String.prototype" || method.family == "Number.prototype") {
		return nil, false, nil
	}
	// Leave the intrinsic's ordinary direct call on the established path.
	if mode == "direct" && !ast.IsIdentifier(target) {
		return nil, false, nil
	}
	if mode == "direct" && method.source == target {
		return nil, false, nil
	}
	if mode == "bind" {
		value, err := l.libraryMethodBind(node, target, method)
		return value, true, err
	}
	if hasSpread(node) {
		return nil, true, l.notYet(node, "spread into a delayed library call (argument arity is not proven)")
	}
	written := call.Arguments.Nodes
	var receiver *ast.Node
	if mode != "direct" {
		if len(written) == 0 {
			return nil, true, l.notYet(node, "a delayed library call without a receiver")
		}
		receiver, written = written[0], written[1:]
	}
	if mode == "apply" {
		if len(written) != 1 || ast.SkipParentheses(written[0]).Kind != ast.KindArrayLiteralExpression {
			return nil, true, l.notYet(node, "apply without a dense argument literal (length, presence and argument representations must be proven)")
		}
		written = ast.SkipParentheses(written[0]).AsArrayLiteralExpression().Elements.Nodes
		for _, argument := range written {
			if argument.Kind == ast.KindSpreadElement || argument.Kind == ast.KindOmittedExpression {
				return nil, true, l.notYet(node, "apply with holes or spread (argument presence is not proven)")
			}
		}
	}
	prefix := []ir.Expression{}
	if alias, err := l.methodAliasRead(target); err != nil {
		return nil, true, err
	} else if alias != nil {
		prefix = append(prefix, alias)
	}
	if method.family == "global" || method.family == "Number" || method.family == "Math" || method.family == "Object" || method.family == "String" {
		if receiver != nil {
			value, err := l.expression(receiver)
			if err != nil {
				return nil, true, err
			}
			if value.Type() == 0 {
				value = fit(value, ir.Object)
			}
			if slotless(value.Type()) {
				return nil, true, l.notYet(receiver, "a delayed static call with this receiver representation")
			}
			prefix = append(prefix, value)
		}
	}
	var value ir.Expression
	var handled bool
	var err error
	switch method.family {
	case "Array.prototype":
		if receiver == nil {
			return nil, true, l.notYet(node, "an unbound Array method call (this is undefined); supply an array to call/apply")
		}
		if l.mayBeUndefined(receiver) {
			return nil, true, l.notYet(receiver, "an Array method receiver that may be undefined")
		}
		if of, _ := l.representation(l.checker.GetTypeAtLocation(receiver)); of != ir.Array {
			return nil, true, l.notYet(receiver, "an Array method receiver whose dense array representation is not proven")
		}
		switch method.name {
		case "indexOf", "includes", "lastIndexOf":
			value, handled, err = l.libraryArrayMethodArguments(node, receiver, method.name, written)
		case "slice", "join", "at", "push", "reverse", "concat", "fill", "splice":
			if ast.SkipParentheses(receiver).Kind == ast.KindNewExpression && method.name == "fill" {
				return nil, true, l.notYet(node, "delayed fill of an array with holes; write the direct filled construction")
			}
			value, handled, err = l.arrayMethodArguments(node, receiver, method.name, written)
		}
	case "String.prototype":
		if receiver == nil || l.mayBeUndefined(receiver) {
			return nil, true, l.notYet(node, "an unbound method read as a value without a proven present string receiver")
		}
		if of, _ := l.representation(l.checker.GetTypeAtLocation(receiver)); of != ir.String {
			return nil, true, l.notYet(receiver, "a String method receiver whose primitive string representation is not proven")
		}
		var text ir.Expression
		text, err = l.expression(receiver)
		if err == nil {
			value, handled, err = l.libraryStringMethod(node, text, method.name, written)
		}
	case "Number.prototype":
		value, err = l.delayedNumberMethod(node, receiver, method.name, written)
		handled = true
	case "Math":
		count, supported := mathFunctions[method.name]
		if !supported || count < 0 || len(written) != count {
			return nil, true, l.notYet(node, "a delayed Math call without a proven fixed argument count")
		}
		arguments := []ir.Expression{}
		for _, argument := range written {
			number, failure := l.expression(argument)
			if failure != nil {
				return nil, true, failure
			}
			if number.Type() != ir.Number {
				return nil, true, l.notYet(argument, "a delayed Math argument that is not a number")
			}
			arguments = append(arguments, number)
		}
		value, handled = ir.MathCall{Function: method.name, Arguments: arguments}, true
	case "String":
		if method.name == "fromCharCode" || method.name == "fromCodePoint" {
			value, handled, err = l.stringFromCodesArguments(node, method.name == "fromCodePoint", written)
		}
	case "Object.prototype":
		if receiver == nil {
			return nil, true, l.notYet(node, "an Object method without a proven receiver")
		}
		if method.name == "toString" {
			value, err = l.objectTagCall(node, receiver, written)
			handled = true
		} else if method.name == "hasOwnProperty" || method.name == "propertyIsEnumerable" || method.name == "isPrototypeOf" {
			value, handled, err = l.objectPrototypeCallArguments(node, receiver, method.name, written)
		}
	case "Number":
		value, handled, err = l.numberCallArguments(node, method.name, written)
	case "Object":
		if method.name == "freeze" || method.name == "assign" {
			return nil, true, l.notYet(node, "delayed Object mutation with an erased result shape; use a direct call")
		}
		value, handled, err = l.objectCallArguments(node, method.name, written)
	case "global":
		switch method.name {
		case "String":
			handled = true
			if len(written) == 1 {
				value, err = l.stringConversion(written[0])
			} else {
				err = l.notYet(node, "delayed String with other than one argument")
			}
		case "Number":
			handled = true
			if len(written) == 1 {
				value, err = l.libraryNumber(written[0])
			} else {
				err = l.notYet(node, "delayed Number with other than one argument")
			}
		case "parseInt", "parseFloat":
			value, handled, err = l.numberCallArguments(node, method.name, written)
		}
	}
	if !handled {
		return nil, true, l.notYet(node, "delayed "+method.family+"."+method.name+" (no adapter proves this receiver and argument ABI); use an arrow")
	}
	if err != nil {
		return nil, true, err
	}
	if contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil {
		expected, proven := l.representation(contextual)
		if !proven || expected != value.Type() && expected != ir.Maybe(value.Type()) && !(expected == ir.Union && l.writable(contextual) && (value.Type() == ir.String || value.Type() == ir.Number || value.Type() == ir.Boolean || value.Type() == ir.MaybeNumber)) {
			return nil, true, l.notYet(node, "a delayed library result whose contextual representation is not proven")
		}
		if value.Type() == ir.Array && method.family == "Array.prototype" {
			from := l.checker.GetElementTypeOfArrayType(l.checker.GetTypeAtLocation(receiver))
			to := l.checker.GetElementTypeOfArrayType(contextual)
			if from == nil || to == nil || from != to || !l.sameKeeping(from, to, map[[2]*checker.Type]bool{}) {
				return nil, true, l.notYet(node, "a delayed Array result with an erased or different element type; preserve the receiver's exact element type")
			}
		}
	}
	if len(prefix) > 0 {
		if value.Type() == 0 || slotless(value.Type()) {
			return nil, true, l.notYet(node, "a delayed library call returning this representation")
		}
		value = l.methodSequence(append(prefix, value))
	}
	return value, true, nil
}

func (l *lowering) objectTagCall(node, receiver *ast.Node, written []*ast.Node) (ir.Expression, error) {
	if len(written) != 0 || l.mayBeUndefined(receiver) {
		return nil, l.notYet(node, "Object.prototype.toString with arguments or a possibly undefined receiver")
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, err
	}
	tag := ""
	if _, null := value.(ir.Null); null {
		return l.methodSequence([]ir.Expression{value, ir.StringConstant{Index: l.constant("[object Null]")}}), nil
	}
	switch value.Type() {
	case ir.Number:
		tag = "Number"
	case ir.Boolean:
		tag = "Boolean"
	case ir.String:
		tag = "String"
	case ir.Array:
		tag = "Array"
	case ir.Map:
		if l.isSet(receiver) {
			tag = "Set"
		} else {
			tag = "Map"
		}
	case ir.Closure:
		tag = "Function"
	case ir.Object:
		if reason := l.prototypeHazard(receiver, "toString"); reason != "" {
			return nil, l.notYet(node, "Object.prototype.toString through a view ("+reason+")")
		}
		tag = "Object"
	default:
		return nil, l.notYet(node, "Object.prototype.toString on this representation")
	}
	return l.methodSequence([]ir.Expression{value, ir.StringConstant{Index: l.constant("[object " + tag + "]")}}), nil
}

func (l *lowering) libraryMapCallback(node, receiver *ast.Node, element ir.Type) (ir.Expression, bool, error) {
	method, known := l.libraryMethod(node, map[*ast.Symbol]bool{})
	if !known {
		return nil, false, nil
	}
	index := len(l.result.Functions)
	function := ir.Function{Name: "library_map_" + method.name, Closure: true}
	reads := []ir.Expression{}
	for _, of := range []ir.Type{element, ir.Number, ir.Array} {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "argument", Type: of, Function: index})
		function.Parameters = append(function.Parameters, local)
		reads = append(reads, ir.Read{Local: local, Of: of})
	}
	// Reference-valued optional elements use NULL. String(undefined) is a real string, not an
	// absent result slot, so normalize presence before passing it to the intrinsic adapter.
	sourceElement := l.checker.GetElementTypeOfArrayType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver)))
	if element == ir.String && sourceElement != nil && l.includesUndefined(sourceElement) {
		reads[0] = ir.Coalesce{Value: reads[0], Fallback: ir.StringConstant{Index: l.constant("undefined")}, Of: ir.String}
	}
	var result ir.Expression
	if method.family == "global" && method.name == "String" {
		switch element {
		case ir.Number:
			result = ir.NumberToString{Value: reads[0]}
		case ir.Boolean:
			result = ir.BooleanToString{Value: reads[0]}
		case ir.String:
			result = reads[0]
		case ir.MaybeNumber:
			result = ir.MaybeToString{Value: reads[0]}
		}
	} else if (method.family == "Number" || method.family == "global") && (method.name == "parseInt" || method.name == "parseFloat") && element == ir.String {
		arguments := []ir.Expression{reads[0]}
		if method.name == "parseInt" {
			arguments = append(arguments, reads[1])
		}
		result = ir.NumberCall{Function: method.name, Arguments: arguments}
	}
	if result == nil {
		return nil, true, l.notYet(node, "a library map callback without a proven primitive adapter (map supplies value, index and array); use an arrow")
	}
	function.Returns = result.Type()
	function.Body = []ir.Statement{ir.Return{Value: result}}
	l.result.Functions = append(l.result.Functions, function)
	value := ir.Expression(ir.MakeClosure{Function: index})
	if alias, err := l.methodAliasRead(node); err != nil {
		return nil, true, err
	} else if alias != nil {
		value = l.methodSequence([]ir.Expression{alias, value})
	}
	return value, true, nil
}

// A bound primitive string cannot form a reference cycle. The ordinary captured parameter cell
// snapshots this once and outlives the factory call, rather than rereading a mutable source name.
func (l *lowering) libraryMethodBind(node, target *ast.Node, method libraryMethod) (ir.Expression, error) {
	written := node.AsCallExpression().Arguments.Nodes
	if method.family != "String.prototype" || len(written) != 1 || hasSpread(node) {
		return nil, l.notYet(node, "bind without a supported captured receiver and fixed argument ABI (only zero-argument String methods are lowered); use an arrow")
	}
	switch method.name {
	case "trim", "trimStart", "trimEnd", "toUpperCase", "toLowerCase", "toString", "valueOf":
	default:
		return nil, l.notYet(node, "bind of a method with optional, required or partial arguments (its argument adapter is not lowered); use an arrow")
	}
	receiver, err := l.expression(written[0])
	if err != nil {
		return nil, err
	}
	if receiver.Type() != ir.String || l.mayBeUndefined(written[0]) {
		return nil, l.notYet(node, "bind without a proven present primitive string receiver")
	}
	maker := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "library_bind_factory", Returns: ir.Closure})
	captured := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "bound_receiver", Type: ir.String, Function: maker, Captured: true})
	closure := len(l.result.Functions)
	read := ir.Expression(ir.Read{Local: captured, Of: ir.String})
	result := read
	switch method.name {
	case "trim":
		result = ir.Trim{Value: read}
	case "toString", "valueOf":
	default:
		result = ir.StringCall{Method: method.name, Value: read}
	}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "library_bound_" + method.name, Closure: true, Environment: []int{captured}, Returns: ir.String, Body: []ir.Statement{ir.Return{Value: result}}})
	l.closureRecords = append(l.closureRecords, closureRecord{proven: l.concrete(l.checker.GetTypeAtLocation(node)), function: closure, node: node})
	l.result.Functions[maker].Parameters = []int{captured}
	l.result.Functions[maker].Body = []ir.Statement{ir.Return{Value: ir.MakeClosure{Function: closure}}}
	value := ir.Expression(ir.Call{Function: maker, Arguments: []ir.Expression{receiver}, Returns: ir.Closure})
	if alias, err := l.methodAliasRead(target); err != nil {
		return nil, err
	} else if alias != nil {
		value = l.methodSequence([]ir.Expression{alias, value})
	}
	return value, nil
}

func (l *lowering) delayedNumberMethod(node, receiver *ast.Node, name string, written []*ast.Node) (ir.Expression, error) {
	if receiver == nil {
		return nil, l.notYet(node, "an unbound Number method without a numeric receiver")
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, err
	}
	if value.Type() != ir.Number {
		return nil, l.notYet(receiver, "a Number method receiver without a proven number internal slot")
	}
	if name == "valueOf" && len(written) == 0 {
		return value, nil
	}
	if name != "toString" && name != "toFixed" && name != "toExponential" && name != "toPrecision" {
		return nil, l.notYet(node, "this delayed Number prototype method")
	}
	if len(written) > 1 {
		return nil, l.notYet(node, "a delayed Number format with extra arguments")
	}
	var argument ir.Expression
	if len(written) == 1 {
		argument, err = l.expression(written[0])
		if err != nil {
			return nil, err
		}
		if _, missing := argument.(ir.Undefined); missing {
			argument = nil
		} else if argument.Type() != ir.Number {
			return nil, l.notYet(node, "a delayed Number format argument that is not a number")
		}
	}
	if name == "toFixed" {
		if argument == nil {
			argument = ir.NumberConstant{}
		}
		return ir.ToFixed{Value: value, Digits: argument}, nil
	}
	return ir.NumberFormat{Method: name, Value: value, Argument: argument}, nil
}

// [] has element type never. No invocation occurs, but the callback value is still read, including
// its temporal dead zone check; it is not replaced by an arbitrary guessed element conversion.
func (l *lowering) libraryEmptyMap(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "map" || !l.libraryMember(callee) || len(call.Arguments.Nodes) != 1 {
		return nil, false, nil
	}
	receiver := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	if receiver.Kind != ast.KindArrayLiteralExpression || len(receiver.AsArrayLiteralExpression().Elements.Nodes) != 0 {
		return nil, false, nil
	}
	method, known := l.libraryMethod(call.Arguments.Nodes[0], map[*ast.Symbol]bool{})
	if !known {
		return nil, false, nil
	}
	var element ir.Type
	if method.family == "global" && method.name == "String" {
		element = ir.String
	}
	if (method.family == "Number" || method.family == "global") && (method.name == "parseInt" || method.name == "parseFloat") {
		element = ir.Number
	}
	if element == 0 {
		return nil, true, l.notYet(node, "an empty map with an unsupported library callback value")
	}
	value := ir.Expression(ir.ArrayLiteral{Element: element})
	if alias, err := l.methodAliasRead(call.Arguments.Nodes[0]); err != nil {
		return nil, true, err
	} else if alias != nil {
		value = l.methodSequence([]ir.Expression{alias, value})
	}
	return value, true, nil
}
