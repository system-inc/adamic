package lower

import (
	"github.com/system-inc/adamic/internal/ir"
)

// String(v) and a template's ${v}, where v is a union that may hold an object, an unknown among them
// (#09dvwny). Node takes the object apart by ToPrimitive with the string hint: toString, then valueOf,
// each called when it's a function, the first primitive either gives taken, and a TypeError when
// neither gives one. A call the runtime made by name would be one no pass sees (effects, exceptions,
// borrowing, regions), so lowering makes each call itself, as ordinary IR, in one helper per program,
// to_string_hint, which every such site calls. finishToPrimitive writes its body once every class is
// lowered, from the closed world: every class the program makes has an arm, and every plain object's
// own toString and valueOf are functions the program wrote where it can say which. Whatever the
// helper can't answer as Node does stops loudly as not yet, never as a guess.
//
// A symbol is where String and a template part: String(symbol) is "Symbol(description)", and a
// template of one throws a TypeError. No symbol has a representation (representation), so none can
// reach a union, and one helper serves both. The day one can, the two forms need two arms here.

// toStringHint is String(value) for a union that may hold an object: a call to the program's one
// to_string_hint helper.
func (l *lowering) toStringHint(value ir.Expression) ir.Expression {
	if l.toPrimitive == 0 {
		index := len(l.result.Functions)
		parameter := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "value", Type: ir.Union, Function: index})
		l.result.Functions = append(l.result.Functions, ir.Function{Name: "to_string_hint", Parameters: []int{parameter}, Returns: ir.String})
		l.toPrimitive = index + 1
	}
	return ir.Call{Function: l.toPrimitive - 1, Arguments: []ir.Expression{value}, Returns: ir.String}
}

// toPrimitiveStop is what the helper says where it doesn't answer yet: exit 70, naming the kind of
// value, as every not yet does at run time.
func (l *lowering) toPrimitiveStop(what string) ir.Statement {
	return ir.Panic{Message: ir.StringConstant{Index: l.constant("converting " + what + " to a string is not yet lowered")}}
}

// finishToPrimitive writes to_string_hint's body, an arm for each kind of value ir.UnknownKind
// names, in the order adamic_union_kind tells them apart. It runs after finishClassCalls and
// finishAccessors, so every class and its method slots are known and the accessor pass leaves its
// calls alone, and before exceptions(), so MayThrow sees every method it calls.
func (l *lowering) finishToPrimitive() {
	if l.toPrimitive == 0 {
		return
	}
	helper := l.toPrimitive - 1
	value := ir.Read{Local: l.result.Functions[helper].Parameters[0], Of: ir.Union}
	kind := l.toPrimitiveLocal(helper, "kind", ir.Number)
	is := func(number int) ir.Expression {
		return ir.Binary{Operator: ir.Equal, Left: ir.Read{Local: kind, Of: ir.Number}, Right: ir.NumberConstant{Value: float64(number)}}
	}
	text := func(written string) ir.Expression { return ir.StringConstant{Index: l.constant(written)} }
	object := ir.Narrow{Value: value, To: ir.Object}
	arm := func(number int, body ...ir.Statement) ir.Statement {
		return ir.If{Condition: is(number), Then: body}
	}
	l.result.Functions[helper].Body = []ir.Statement{
		ir.Declare{Local: kind, Value: ir.UnknownKind{Value: value}},
		arm(ir.UnknownUndefined, ir.Return{Value: text("undefined")}),
		arm(ir.UnknownNull, ir.Return{Value: text("null")}),
		arm(ir.UnknownNumber, ir.Return{Value: ir.UnionToString{Value: value}}),
		arm(ir.UnknownBoolean, ir.Return{Value: ir.UnionToString{Value: value}}),
		arm(ir.UnknownString, ir.Return{Value: ir.UnionToString{Value: value}}),
		// Before any object's arm, through adamic_error_is, the runtime's one test of an Error (Lower's
		// ruling): no class a program declares extends Error yet (adamic.h), so no class's arm shadows it.
		arm(ir.UnknownError, ir.Return{Value: ir.ErrorToString{Value: object}}),
		arm(ir.UnknownInstance, l.instanceToString(helper, object)...),
		arm(ir.UnknownObject, l.plainToString(helper, object)...),
		// Node prints a class's own source text, and a function's.
		arm(ir.UnknownStaticClass, l.toPrimitiveStop("a class object")),
		arm(ir.UnknownFunction, l.toPrimitiveStop("a function")),
		arm(ir.UnknownArray, l.arrayToString(helper, value)...),
		// unknownView refuses a function, a Map, a Set and a tuple where it flows into an unknown, so
		// these arms are unreachable today; they stop rather than guess the day one gets through.
		arm(ir.UnknownCollection, l.toPrimitiveStop("a Map or a Set")),
		arm(ir.UnknownTypedArray, l.toPrimitiveStop("a typed array")),
		l.toPrimitiveStop("a value of this kind"),
		// Never reached, past the stop: a function that gives a string ends in a return, as
		// newClassValue's does.
		ir.Return{Value: ir.Undefined{Of: ir.String}},
	}
}

func (l *lowering) toPrimitiveLocal(function int, name string, of ir.Type) int {
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: of, Function: function})
	return local
}

// instanceToString is an arm for every non-static class the program makes, the closed world, which
// over-approximates the classes that reach an unknown and so can't miss one. Each is told by exact
// identity, so the method in its slot is the one that runs, and is called directly, as a devirtualized
// call is.
func (l *lowering) instanceToString(helper int, object ir.Expression) []ir.Statement {
	body := []ir.Statement{}
	for index, class := range l.result.Classes {
		if class.Static {
			continue
		}
		body = append(body, ir.If{
			Condition: ir.InstanceOf{Value: object, Class: index + 1, Exact: true},
			Then:      l.classToString(helper, index+1, object),
		})
	}
	// Unreachable: every class an object can have is one of the arms above (TestToPrimitiveCoversEveryClass).
	return append(body, ir.Panic{Message: ir.StringConstant{Index: l.constant("compiler bug: a class instance without a ToPrimitive arm")}})
}

// classToString is ToPrimitive with the string hint for an object of one class: its toString, which
// its method table holds through the base chain, and then its valueOf. A class that declares neither
// inherits Object.prototype.toString, whose "[object Object]" is a primitive, so valueOf isn't asked.
func (l *lowering) classToString(helper, class int, object ir.Expression) []ir.Statement {
	stop := func(why string) []ir.Statement {
		return []ir.Statement{l.toPrimitiveStop("an object whose class " + why)}
	}
	for at := class; at != 0; at = l.result.Classes[at-1].Base {
		for _, accessor := range l.result.Classes[at-1].Accessors {
			if accessor.Name == "toString" || accessor.Name == "valueOf" {
				return stop("has a toString or valueOf accessor")
			}
		}
	}
	metadata := l.result.Classes[class-1]
	for _, fields := range [][]ir.Field{metadata.Fields, metadata.PublicFields} {
		for _, field := range fields {
			if !field.Private && (field.Name == "toString" || field.Name == "valueOf") {
				// An own field, a function or not, is what Node reads first.
				return stop("has a toString or valueOf field")
			}
		}
	}
	method := func(name string) int {
		for _, function := range metadata.Methods {
			if function >= 0 && l.result.Functions[function].MethodName == name {
				return function
			}
		}
		return -1
	}
	call := func(function int) ([]ir.Statement, bool) {
		if len(l.result.Functions[function].Parameters) != 1 || l.result.Functions[function].RestElement != 0 {
			return stop("has a toString or valueOf taking parameters"), false
		}
		return l.primitiveResult(helper, ir.Call{Function: function, Arguments: []ir.Expression{object}, Returns: l.result.Functions[function].Returns})
	}
	toString := method("toString")
	if toString < 0 {
		return []ir.Statement{ir.Return{Value: ir.StringConstant{Index: l.constant("[object Object]")}}}
	}
	body, goesOn := call(toString)
	if !goesOn {
		return body
	}
	if valueOf := method("valueOf"); valueOf >= 0 {
		next, goesOn := call(valueOf)
		body = append(body, next...)
		if !goesOn {
			return body
		}
	}
	// Object.prototype.valueOf gives the object itself, which isn't a primitive.
	return append(body, l.toPrimitiveTypeError())
}

// arrayToString is Array.prototype.join's "," for an array, each element through the helper again, as
// Node's join writes it with String, but null and undefined as "". An array whose join is already under
// way, one that holds itself however deep, joins as "", as V8's join stack makes it (Compiler's must on
// #09dvwny); the stack lets go of the array however its join ends, so a join a throw left joins again.
// The length is read once, as join reads it, and an element a toString took away since is undefined.
// An array of only numbers, booleans, null or undefined, with a number or a boolean among them or only
// undefined, stops, named: natively an array of numbers or booleans, or of number | undefined, holds
// them unboxed, saying nothing of which they are, so the test both backends make alike is by the
// elements (ir.ArrayOfScalars), until Memory's element kind comes. Its cost is an over-stop, not a
// hole: an array of a union type holding only numbers, or only undefined, stops too, though natively
// it's boxed and could join, since the JavaScript backend can't tell it from a number[] or a number |
// undefined one. Both backends stop on it. [null, undefined] joins, since a null is always boxed.
func (l *lowering) arrayToString(helper int, value ir.Expression) []ir.Statement {
	array := ir.Narrow{Value: value, To: ir.Array}
	joined := l.toPrimitiveLocal(helper, "joined", ir.String)
	length := l.toPrimitiveLocal(helper, "length", ir.Number)
	index := l.toPrimitiveLocal(helper, "index", ir.Number)
	element := l.toPrimitiveLocal(helper, "element", ir.Union)
	readJoined := ir.Read{Local: joined, Of: ir.String}
	readIndex := ir.Read{Local: index, Of: ir.Number}
	readElement := ir.Read{Local: element, Of: ir.Union}
	text := func(written string) ir.Expression { return ir.StringConstant{Index: l.constant(written)} }
	add := func(part ir.Expression) ir.Statement {
		return ir.Assign{Local: joined, Value: ir.Concat{Parts: []ir.Expression{readJoined, part}}}
	}
	// null and undefined, the two kinds numbered first, join as nothing.
	present := ir.Binary{Operator: ir.Greater, Left: ir.UnknownKind{Value: readElement}, Right: ir.NumberConstant{Value: ir.UnknownNull}}
	each := []ir.Statement{
		ir.If{Condition: ir.Binary{Operator: ir.Greater, Left: readIndex, Right: ir.NumberConstant{Value: 0}}, Then: []ir.Statement{add(text(","))}},
		ir.If{Condition: ir.Binary{Operator: ir.Less, Left: readIndex, Right: ir.Length{Array: array}}, Then: []ir.Statement{
			ir.Declare{Local: element, Value: ir.ArrayIndex{Array: array, Index: readIndex, Element: ir.Union}},
			ir.If{Condition: present, Then: []ir.Statement{add(ir.Call{Function: helper, Arguments: []ir.Expression{readElement}, Returns: ir.String})}},
		}},
	}
	loop := ir.Loop{
		Condition: ir.Binary{Operator: ir.Less, Left: readIndex, Right: ir.Read{Local: length, Of: ir.Number}},
		Body:      each,
		Update:    []ir.Statement{ir.Assign{Local: index, Value: ir.Binary{Operator: ir.Add, Left: readIndex, Right: ir.NumberConstant{Value: 1}}}},
	}
	return []ir.Statement{
		ir.If{Condition: ir.ArrayOfScalars{Array: array}, Then: []ir.Statement{l.toPrimitiveStop("an array of only numbers, booleans, null or undefined")}},
		ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: ir.JoinEnter{Array: array}}, Then: []ir.Statement{ir.Return{Value: text("")}}},
		ir.Declare{Local: joined, Value: text("")},
		ir.Declare{Local: length, Value: ir.Length{Array: array}},
		ir.Declare{Local: index, Value: ir.NumberConstant{Value: 0}},
		ir.Try{Body: []ir.Statement{loop}, Finally: []ir.Statement{ir.Evaluate{Value: ir.JoinLeave{}}}, CatchLocal: -1, HasFinally: true},
		ir.Return{Value: readJoined},
	}
}

// toPrimitiveTypeError is the throw when neither method gives a primitive, in exceptions.c's words
// (adamic_ordinary_to_primitive), a TypeError a catch takes.
func (l *lowering) toPrimitiveTypeError() ir.Statement {
	return ir.Throw{Value: ir.MakeError{Message: ir.StringConstant{Index: l.constant("Cannot convert object to primitive value")}, Name: ir.StringConstant{Index: l.constant("TypeError")}}}
}

// primitiveResult is what ToPrimitive does with what a method gave, classified by the method's own
// representation: a string, a number, a boolean, undefined (void, or a missing string or number) and
// null are primitives, each written as String writes it and returned; an object goes on to the next
// method; and a union is told apart when it's there. goesOn says the statements can fall through to
// the next method.
func (l *lowering) primitiveResult(helper int, call ir.Expression) ([]ir.Statement, bool) {
	text := func(written string) ir.Expression { return ir.StringConstant{Index: l.constant(written)} }
	switch of := call.Type(); of {
	case 0:
		// A method that returns nothing gives undefined, a primitive, so valueOf isn't asked.
		return []ir.Statement{ir.Evaluate{Value: call}, ir.Return{Value: text("undefined")}}, false
	case ir.String:
		return []ir.Statement{ir.Return{Value: ir.Coalesce{Value: call, Fallback: text("undefined"), Of: ir.String}}}, false
	case ir.Number:
		return []ir.Statement{ir.Return{Value: ir.NumberToString{Value: call}}}, false
	case ir.Boolean:
		return []ir.Statement{ir.Return{Value: ir.BooleanToString{Value: call}}}, false
	case ir.MaybeNumber, ir.MaybeBoolean:
		return []ir.Statement{ir.Return{Value: ir.MaybeToString{Value: call}}}, false
	case ir.Union:
		// A primitive in it is written by the helper's own primitive arms; an object goes on.
		result := l.toPrimitiveLocal(helper, "result", ir.Union)
		read := ir.Read{Local: result, Of: ir.Union}
		primitive := ir.Binary{Operator: ir.LessOrEqual, Left: ir.UnknownKind{Value: read}, Right: ir.NumberConstant{Value: ir.UnknownString}}
		return []ir.Statement{
			ir.Declare{Local: result, Value: call},
			ir.If{Condition: primitive, Then: []ir.Statement{ir.Return{Value: ir.Call{Function: helper, Arguments: []ir.Expression{read}, Returns: ir.String}}}},
		}, true
	default:
		if !of.IsReference() {
			return []ir.Statement{l.toPrimitiveStop("an object whose toString or valueOf gives this representation")}, false
		}
		// An object goes on. A missing one is undefined or null, which a reference of this kind doesn't
		// tell apart (unknownView's nullable reference), so it stops rather than guess which.
		result := l.toPrimitiveLocal(helper, "result", of)
		return []ir.Statement{
			ir.Declare{Local: result, Value: call},
			ir.If{Condition: ir.IsUndefined{Value: ir.Read{Local: result, Of: of}}, Then: []ir.Statement{l.toPrimitiveStop("an object whose toString or valueOf gives a missing object")}},
		}, true
	}
}

// plainToString is ToPrimitive with the string hint for a plain object, one an object literal made or
// a record (the JavaScript backend can't tell the two apart, so neither does this arm). Present means
// an own property, never an inherited one (adamic_has_property says yes to toString on every object):
// one holding a function is called, one holding anything else is skipped, as Node skips it, and with
// no own toString at all, Object.prototype.toString's "[object Object]" is the answer.
func (l *lowering) plainToString(helper int, object ir.Expression) []ir.Statement {
	plain := l.plainMethods()
	step := func(name string) []ir.Statement {
		functions, closed := plain.functions[name], plain.closed[name]
		if !closed || plain.records {
			// A function the program put there that this can't name, or a record's, which a call through
			// the object's own field (adamic_object_callee) can't find.
			return []ir.Statement{l.toPrimitiveStop("an object whose own " + name + " is a function this program can't name")}
		}
		returns := ir.Type(-1)
		for _, function := range functions {
			// A method in a literal takes its object as this, its first parameter; an arrow function
			// takes none.
			receiver := 0
			if l.result.Functions[function].Receiver {
				receiver = 1
			}
			if len(l.result.Functions[function].Parameters) != receiver || l.result.Functions[function].RestElement != 0 {
				return []ir.Statement{l.toPrimitiveStop("an object whose own " + name + " takes parameters")}
			}
			if returns != -1 && returns != l.result.Functions[function].Returns {
				return []ir.Statement{l.toPrimitiveStop("an object whose own " + name + " functions give different representations")}
			}
			returns = l.result.Functions[function].Returns
		}
		if returns == -1 {
			// No function the program writes is ever one, so no object's own one is.
			return []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant("compiler bug: an own " + name + " function outside ToPrimitive's closed world")}}}
		}
		// Through the object's own field, as object.name() through an interface is, which passes the
		// object as this to a method and nothing to an arrow function.
		call := ir.CallClosure{Closure: ir.Property{Object: object, Name: name, Of: ir.Closure, Method: true}, Returns: returns}
		body, _ := l.primitiveResult(helper, call)
		return body
	}
	key := func(name string) ir.Expression { return ir.StringConstant{Index: l.constant(name)} }
	return []ir.Statement{
		ir.If{Condition: ir.OwnFunction{Object: object, Name: "toString"}, Then: step("toString"), Else: []ir.Statement{
			ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: ir.HasOwn{Object: object, Key: key("toString")}}, Then: []ir.Statement{ir.Return{Value: key("[object Object]")}}},
		}},
		ir.If{Condition: ir.OwnFunction{Object: object, Name: "valueOf"}, Then: step("valueOf")},
		// Object.prototype.valueOf gives the object itself, which isn't a primitive.
		l.toPrimitiveTypeError(),
	}
}

// plainMethods is the closed world of the functions an object's own toString or valueOf can hold: every
// value the program writes into a property of either name, by an object literal or an assignment. A
// spread or Object.assign copies only what was written somewhere first. closed is false for a name
// where some value written is a function this can't name (a parameter, a variable, a union that may
// hold one), and records says some record may hold a function under any key at all.
type plainMethodSet struct {
	functions map[string][]int
	closed    map[string]bool
	records   bool
}

func (l *lowering) plainMethods() plainMethodSet {
	set := plainMethodSet{functions: map[string][]int{}, closed: map[string]bool{"toString": true, "valueOf": true}}
	written := func(name string, value ir.Expression) {
		if _, named := set.closed[name]; !named || value == nil {
			return
		}
		switch value.Type() {
		case ir.Closure:
			if made, known := value.(ir.MakeClosure); known {
				set.functions[name] = append(set.functions[name], made.Function)
				return
			}
			set.closed[name] = false
		case ir.Union:
			set.closed[name] = false
		}
	}
	holds := func(element ir.Type) {
		set.records = set.records || element == ir.Closure || element == ir.Union
	}
	visit := func(node any) bool {
		switch node := node.(type) {
		case ir.ObjectLiteral:
			for _, field := range node.Fields {
				written(field.Name, field.Value)
				if node.Record && field.Value != nil {
					holds(field.Value.Type())
				}
			}
		case ir.SetProperty:
			written(node.Name, node.Value)
		case ir.RecordLiteral:
			holds(node.Element)
		case ir.RecordCall:
			holds(node.Element)
		case ir.RecordCoalesce:
			holds(node.Element)
		}
		return true
	}
	walk(l.result.Main, visit)
	for _, function := range l.result.Functions {
		walk(function.Body, visit)
	}
	return set
}
