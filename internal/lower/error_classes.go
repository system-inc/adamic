package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

var errorNames = []string{"Error", "TypeError", "RangeError", "ReferenceError", "SyntaxError", "URIError", "EvalError"}

func (l *lowering) errorGlobal(node *ast.Node) string {
	for _, name := range errorNames {
		if l.isLibraryGlobal(node, name) {
			return name
		}
	}
	return ""
}

func (l *lowering) errorType(proven *checker.Type) bool {
	if l.isLibraryType(proven, errorNames...) {
		return true
	}
	if proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if base := l.checker.GetBaseConstraintOfType(proven); base != nil {
			return l.errorType(base)
		}
		return false
	}
	if !isClassInstance(proven) {
		return false
	}
	for _, base := range l.classBases(proven) {
		if l.errorType(base) {
			return true
		}
	}
	return false
}

// Built-in errors use the same prefix layout and nominal dispatch as user classes.
// Reserved definition IDs cannot collide with an erased source class identity.
func (l *lowering) errorInstance(name string) *instance {
	key := "builtin-error:" + name
	if found := l.instances[key]; found != nil {
		return found
	}
	var base *instance
	if name != "Error" {
		base = l.errorInstance("Error")
	}
	if l.instances == nil {
		l.instances = map[string]*instance{}
	}
	lowered := &instance{class: len(l.result.Classes) + 1, constructor: len(l.result.Functions), initializer: len(l.result.Functions) + 1, methods: map[string]int{}, slots: map[string]int{}, staticMethods: map[string]bool{}, base: base, builtinError: name}
	l.instances[key] = lowered
	definition := 1 << 30
	for i, kind := range errorNames {
		if kind == name {
			definition += i
		}
	}
	fields := []ir.Field{{Name: "name", Value: ir.Undefined{Of: ir.String}}, {Name: "message", Value: ir.Undefined{Of: ir.String}}, {Name: "cause", Value: ir.Undefined{Of: ir.Union}}}
	metadata := ir.Class{Name: name, Definition: definition, Constructor: lowered.constructor, Fields: fields}
	if base != nil {
		metadata.Base = base.class
		metadata.OwnStart = len(fields)
	}
	l.result.Classes = append(l.result.Classes, metadata)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: name + "_new", Returns: ir.Object}, ir.Function{Name: name + "_initialize", LibraryGuarded: true})
	parameter := func(owner int, name string, of ir.Type) int {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: of, Function: owner})
		return local
	}
	initThis := parameter(lowered.initializer, "this", ir.Object)
	initMessage := parameter(lowered.initializer, "message", ir.String)
	initCause := parameter(lowered.initializer, "cause", ir.Union)
	l.result.Functions[lowered.initializer].Parameters = []int{initThis, initMessage, initCause}
	message := ir.Coalesce{Value: ir.Read{Local: initMessage, Of: ir.String}, Fallback: ir.StringConstant{Index: l.constant("")}, Of: ir.String}
	// The built-in initializer is called only on a newly allocated error or a
	// fresh subclass receiver, before user code can freeze it. LibraryGuarded
	// keeps these internal writes out of the frozen-object refusal boundary.
	// A synthesized initializer has no source receiver whose checker type can be
	// recorded. Keep the holder unknown, conservatively, and let the existing
	// interprocedural proof establish that construction writes into a fresh object.
	write := func() int {
		l.writeSites = append(l.writeSites, writeSite{node: l.program.Files()[0].AsNode()})
		return len(l.writeSites)
	}
	l.result.Functions[lowered.initializer].Body = []ir.Statement{
		ir.SetProperty{Object: ir.Read{Local: initThis, Of: ir.Object}, Name: "name", Value: ir.StringConstant{Index: l.constant(name)}, Site: write()},
		ir.SetProperty{Object: ir.Read{Local: initThis, Of: ir.Object}, Name: "message", Value: message, Site: write()},
		ir.SetProperty{Object: ir.Read{Local: initThis, Of: ir.Object}, Name: "cause", Value: ir.Read{Local: initCause, Of: ir.Union}, Site: write()},
	}
	method := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: name + "_toString", Returns: ir.String})
	this := parameter(method, "this", ir.Object)
	l.result.Functions[method].Parameters = []int{this}
	receiver := ir.Read{Local: this, Of: ir.Object}
	n := ir.Property{Object: receiver, Name: "name", Of: ir.String}
	m := ir.Property{Object: receiver, Name: "message", Of: ir.String}
	empty := ir.StringConstant{Index: l.constant("")}
	result := ir.Conditional{Condition: ir.Binary{Operator: ir.Equal, Left: n, Right: empty}, WhenTrue: m, WhenNot: ir.Conditional{Condition: ir.Binary{Operator: ir.Equal, Left: m, Right: empty}, WhenTrue: n, WhenNot: ir.Concat{Parts: []ir.Expression{n, ir.StringConstant{Index: l.constant(": ")}, m}}}}
	l.result.Functions[method].Body = []ir.Statement{ir.Return{Value: result}}
	lowered.methods["toString"] = method
	lowered.slots["toString"] = 0
	l.result.Classes[lowered.class-1].Methods = []int{method}
	incoming := parameter(lowered.constructor, "message", ir.String)
	incomingCause := parameter(lowered.constructor, "cause", ir.Union)
	object := parameter(lowered.constructor, "this", ir.Object)
	l.result.Functions[lowered.constructor].Parameters = []int{incoming, incomingCause}
	l.result.Functions[lowered.constructor].Body = []ir.Statement{
		ir.Declare{Local: object, Value: ir.ObjectLiteral{Class: lowered.class, Fields: fields, Methods: lowered.methodList()}},
		ir.Evaluate{Value: ir.Call{Function: lowered.initializer, Arguments: []ir.Expression{ir.Read{Local: object, Of: ir.Object}, ir.Read{Local: incoming, Of: ir.String}, ir.Read{Local: incomingCause, Of: ir.Union}}}},
		ir.Return{Value: ir.Read{Local: object, Of: ir.Object}},
	}
	return lowered
}

func (l *lowering) errorMessage(node *ast.Node, args []*ast.Node) (ir.Expression, error) {
	if len(args) > 2 {
		return nil, l.notYet(node, "an Error constructor with more than two arguments")
	}
	if len(args) == 0 {
		return ir.Undefined{Of: ir.String}, nil
	}
	value, err := l.expression(args[0])
	if err != nil {
		return nil, err
	}
	if value.Type() != ir.String {
		if _, ok := value.(ir.Undefined); ok {
			return ir.Undefined{Of: ir.String}, nil
		}
		return nil, l.notYet(node, "an Error message that needs JavaScript ToString coercion")
	}
	return value, nil
}

func (l *lowering) errorMethod(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "toString" {
		return nil, false, nil
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	if !l.errorType(l.checker.GetTypeAtLocation(receiver)) {
		return nil, false, nil
	}
	symbol := l.checker.GetSymbolAtLocation(callee.Name())
	if symbol != nil && len(symbol.Declarations) > 0 && symbol.Declarations[0].Kind == ast.KindMethodDeclaration {
		return nil, false, nil
	}
	if len(nodesOf(node.AsCallExpression().Arguments)) != 0 {
		return nil, true, l.notYet(node, "Error.toString with arguments")
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	instance := l.errorInstance("Error")
	return ir.Call{Function: instance.methods["toString"], Virtual: 1, Returns: ir.String, Arguments: []ir.Expression{value}}, true, nil
}

// ErrorOptions is inspected only while its literal still exposes the cause's real type.
// This keeps erased unknown out of ordinary object layouts.
func (l *lowering) errorCause(node *ast.Node, args []*ast.Node) (ir.Expression, error) {
	if len(args) < 2 {
		return ir.Undefined{Of: ir.Union}, nil
	}
	options := ast.SkipParentheses(args[1])
	if ast.IsIdentifier(options) && options.Text() == "undefined" {
		return ir.Undefined{Of: ir.Union}, nil
	}
	if options.Kind != ast.KindObjectLiteralExpression {
		return nil, l.notYet(options, "ErrorOptions not written as a literal: the erased cause's representation and ownership must be known where it is boxed")
	}
	var result ir.Expression = ir.Undefined{Of: ir.Union}
	for _, field := range options.AsObjectLiteralExpression().Properties.Nodes {
		if field.Name() == nil || field.Name().Text() != "cause" || (field.Kind != ast.KindPropertyAssignment && field.Kind != ast.KindShorthandPropertyAssignment) {
			return nil, l.notYet(field, "an ErrorOptions literal with computed, spread or other fields")
		}
		var value ir.Expression
		var err error
		var proven *checker.Type
		if field.Kind == ast.KindPropertyAssignment {
			value, err = l.expression(field.AsPropertyAssignment().Initializer)
			proven = l.checker.GetTypeAtLocation(field.AsPropertyAssignment().Initializer)
		} else {
			proven = l.checker.GetTypeOfSymbol(l.checker.GetShorthandAssignmentValueSymbol(field))
			if proven.Flags()&checker.TypeFlagsUnknown != 0 {
				return nil, l.notYet(field, "an erased unknown cause: narrow it before constructing the error so its possible reference cycles can be proven")
			}
			value, err = l.shorthand(field)
		}
		if err != nil {
			return nil, err
		}
		if proven.Flags()&checker.TypeFlagsUnknown != 0 {
			return nil, l.notYet(field, "an erased unknown cause: narrow it before constructing the error so its possible reference cycles can be proven")
		}
		l.errorInstance("Error").errorCauses = append(l.errorInstance("Error").errorCauses, l.concrete(proven))
		result = fit(value, ir.Union)
		if boxed, ok := result.(ir.Box); ok && l.includesNull(proven) {
			boxed.Nullable = true
			result = boxed
		}
	}
	return result, nil
}

func (l *lowering) isErrorCause(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "cause" && l.errorType(l.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression))
}

// The standard method is generic, but never reads a prototype method as an own slot.
func (l *lowering) errorPrototypeCall(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "call" {
		return nil, false, nil
	}
	method := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	if method.Kind != ast.KindPropertyAccessExpression || method.Name().Text() != "toString" {
		return nil, false, nil
	}
	prototype := ast.SkipParentheses(method.AsPropertyAccessExpression().Expression)
	if prototype.Kind != ast.KindPropertyAccessExpression || prototype.Name().Text() != "prototype" || l.errorGlobal(prototype.AsPropertyAccessExpression().Expression) == "" {
		return nil, false, nil
	}
	args := nodesOf(node.AsCallExpression().Arguments)
	if len(args) != 1 {
		return nil, true, l.notYet(node, "Error.prototype.toString.call without exactly one receiver")
	}
	value, err := l.expression(args[0])
	if err != nil {
		return nil, true, err
	}
	if value.Type() == ir.Object && l.includesUndefined(l.checker.GetTypeAtLocation(args[0])) {
		return nil, true, l.notYet(node, "Error.prototype.toString on a possibly undefined object: generic field lookup needs a proven receiver")
	}
	index := len(l.result.Functions)
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "receiver", Type: value.Type(), Function: index})
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "Error_prototype_toString", Parameters: []int{local}, Returns: ir.String})
	read := ir.Read{Local: local, Of: value.Type()}
	_, nullReceiver := value.(ir.Null)
	if value.Type() != ir.Object || nullReceiver {
		var text ir.Expression
		switch {
		case nullReceiver:
			text = ir.StringConstant{Index: l.constant("null")}
		case value.Type() == ir.Number:
			text = ir.NumberToString{Value: read}
		case value.Type() == ir.Boolean:
			text = ir.BooleanToString{Value: read}
		case value.Type() == ir.String:
			text = ir.Coalesce{Value: read, Fallback: ir.StringConstant{Index: l.constant("undefined")}, Of: ir.String}
		default:
			return nil, true, l.notYet(node, "Error.prototype.toString on this receiver representation")
		}
		// Arrays and functions are objects in the specification; they need generic property lookup.
		message := ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: l.constant("Method Error.prototype.toString called on incompatible receiver ")}, text}}
		builtin := l.errorInstance("TypeError")
		l.result.Functions[index].Body = []ir.Statement{ir.Throw{Value: ir.Call{Function: builtin.constructor, Returns: ir.Object, Arguments: []ir.Expression{message, ir.Undefined{Of: ir.Union}}}}}
	} else {
		proven := l.checker.GetTypeAtLocation(args[0])
		field := func(name, fallback string) (ir.Expression, error) {
			symbol := l.checker.GetPropertyOfType(proven, name)
			if symbol == nil {
				return ir.StringConstant{Index: l.constant(fallback)}, nil
			}
			of, known := l.representation(l.checker.GetTypeOfSymbol(symbol))
			if !known || symbol.Flags&ast.SymbolFlagsOptional != 0 {
				return nil, l.notYet(node, "Error.prototype.toString with an optional or erased field: own property absence is not represented")
			}
			var result ir.Expression = ir.Property{Object: read, Name: name, Of: of}
			switch of {
			case ir.String:
				if l.includesUndefined(l.checker.GetTypeOfSymbol(symbol)) {
					result = ir.Coalesce{Value: result, Fallback: ir.StringConstant{Index: l.constant(fallback)}, Of: ir.String}
				}
			case ir.Number:
				result = ir.NumberToString{Value: result}
			case ir.Boolean:
				result = ir.BooleanToString{Value: result}
			default:
				return nil, l.notYet(node, "Error.prototype.toString with a field whose ToString coercion is not lowered yet")
			}
			return result, nil
		}
		name, err := field("name", "Error")
		if err != nil {
			return nil, true, err
		}
		message, err := field("message", "")
		if err != nil {
			return nil, true, err
		}
		empty := ir.StringConstant{Index: l.constant("")}
		result := ir.Conditional{Condition: ir.Binary{Operator: ir.Equal, Left: name, Right: empty}, WhenTrue: message, WhenNot: ir.Conditional{Condition: ir.Binary{Operator: ir.Equal, Left: message, Right: empty}, WhenTrue: name, WhenNot: ir.Concat{Parts: []ir.Expression{name, ir.StringConstant{Index: l.constant(": ")}, message}}}}
		l.result.Functions[index].Body = []ir.Statement{ir.Return{Value: result}}
	}
	return ir.Call{Function: index, Returns: ir.String, Arguments: []ir.Expression{value}}, true, nil
}

func (l *lowering) errorPrototypeRead(node *ast.Node) bool {
	if node.Name().Text() != "toString" {
		return false
	}
	prototype := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
	if prototype.Kind != ast.KindPropertyAccessExpression || prototype.Name().Text() != "prototype" || l.errorGlobal(prototype.AsPropertyAccessExpression().Expression) == "" {
		return false
	}
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	return parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.Name().Text() == "call" && called(parent)
}

// Generated class failures have nominal identities just like source Error values.
// Their constant message and absent cause need no user constructor call.
func (l *lowering) generatedError(name, message string) ir.Expression {
	instance := l.errorInstance(name)
	return ir.ObjectLiteral{Class: instance.class, Methods: instance.methodList(), Fields: []ir.Field{
		{Name: "name", Value: ir.StringConstant{Index: l.constant(name)}},
		{Name: "message", Value: ir.StringConstant{Index: l.constant(message)}},
		{Name: "cause", Value: ir.Undefined{Of: ir.Union}},
	}}
}
