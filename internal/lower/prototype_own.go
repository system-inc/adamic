package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// nonObjectOwnProperty spells the descriptors of the values whose shape is fixed by the language.
// A generated function keeps the receiver and key in parameters, so each is evaluated once, in
// JavaScript's order, even when the answer is constant. No native representation is read as an object.
func (l *lowering) nonObjectOwnProperty(node, receiver *ast.Node, name string, of ir.Type) (ir.Expression, bool, error) {
	functionPrototype := -1
	if of == ir.Closure {
		functionPrototype = l.functionPrototype(receiver, 0)
	}
	switch of {
	case ir.Number, ir.Boolean, ir.String, ir.Array, ir.Map:
	case ir.Closure:
		if name == "hasOwnProperty" && functionPrototype < 0 {
			return nil, true, l.notYet(node, "hasOwnProperty on a function (its prototype property depends on whether it was declared or made as an arrow)")
		}
	default:
		return nil, true, l.notYet(node, name+" on a "+typeName(of)+" (mixed value descriptors are not represented)")
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 {
		return nil, true, l.notYet(node, name+" with other than one key")
	}
	if keyType, _ := l.representation(l.checker.GetTypeAtLocation(arguments[0])); keyType != ir.String {
		return nil, true, l.notYet(node, name+" with a key that isn't a string")
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	key, err := l.expression(arguments[0])
	if err != nil {
		return nil, true, err
	}
	function := len(l.result.Functions)
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals,
		ir.Local{Name: "receiver", Type: of, Function: function},
		ir.Local{Name: "key", Type: ir.String, Function: function},
	)
	body := []ir.Statement{}
	result := ir.Expression(ir.BooleanConstant{Value: false})
	if of == ir.Closure && name == "hasOwnProperty" {
		readKey := ir.Read{Local: held + 1, Of: ir.String}
		result = ir.Binary{Operator: ir.Or,
			Left:  ir.Binary{Operator: ir.Equal, Left: readKey, Right: ir.StringConstant{Index: l.constant("length")}},
			Right: ir.Binary{Operator: ir.Equal, Left: readKey, Right: ir.StringConstant{Index: l.constant("name")}},
		}
		if functionPrototype == 1 {
			result = ir.Binary{Operator: ir.Or, Left: result, Right: ir.Binary{Operator: ir.Equal, Left: readKey, Right: ir.StringConstant{Index: l.constant("prototype")}}}
		}
	}
	if of == ir.Array || of == ir.String {
		index := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "index", Type: ir.Number, Function: function})
		readKey := ir.Read{Local: held + 1, Of: ir.String}
		parsed := ir.Read{Local: index, Of: ir.Number}
		readReceiver := ir.Read{Local: held, Of: of}
		length := ir.Expression(ir.StringLength{Value: readReceiver})
		if of == ir.Array {
			length = ir.Length{Array: readReceiver}
		}
		body = append(body, ir.Declare{Local: index, Value: ir.NumberCall{Function: "parseInt", Arguments: []ir.Expression{readKey, ir.NumberConstant{Value: 10}}}})
		// Only a canonical non-negative index is own. parseInt alone would admit '01', '-0', '1x'
		// and fractional text; the round trip rejects those. A NaN fails the comparisons first.
		result = ir.Binary{Operator: ir.And,
			Left: ir.Binary{Operator: ir.And,
				Left:  ir.Binary{Operator: ir.GreaterOrEqual, Left: parsed, Right: ir.NumberConstant{Value: 0}},
				Right: ir.Binary{Operator: ir.Less, Left: parsed, Right: length},
			},
			Right: ir.Binary{Operator: ir.Equal, Left: ir.NumberToString{Value: parsed}, Right: readKey},
		}
		// Both arrays and boxed strings have an own length, which is not enumerable.
		result = ir.Conditional{
			Condition: ir.Binary{Operator: ir.Equal, Left: readKey, Right: ir.StringConstant{Index: l.constant("length")}},
			WhenTrue:  ir.BooleanConstant{Value: name == "hasOwnProperty"}, WhenNot: result,
		}
	}
	body = append(body, ir.Return{Value: result})
	l.result.Functions = append(l.result.Functions, ir.Function{Name: name + "_" + typeName(of), Parameters: []int{held, held + 1}, Returns: ir.Boolean, Body: body})
	return ir.Call{Function: function, Arguments: []ir.Expression{value, key}, Returns: ir.Boolean}, true, nil
}

// A module function has an own, non-enumerable prototype; an arrow has none. A parameter or field
// can hold either, so it needs a runtime origin tag before hasOwnProperty can answer soundly.
func (l *lowering) functionPrototype(node *ast.Node, depth int) int {
	if depth > 16 {
		return -1
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindArrowFunction {
		return 0
	}
	if !ast.IsIdentifier(node) {
		return -1
	}
	symbol := l.symbol(node)
	if _, found := l.functions[symbol]; found {
		return 1
	}
	if _, found := l.generics[symbol]; found {
		return 1
	}
	if symbol == nil || len(symbol.Declarations) != 1 {
		return -1
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return -1
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	if initializer == nil {
		return -1
	}
	return l.functionPrototype(initializer, depth+1)
}
