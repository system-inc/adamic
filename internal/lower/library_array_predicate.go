package lower

import (
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Port of V8's ArrayFilter/Every/SomeLoopContinuation (Node 24.19.0).
// The receiver is an ordinary dense array: HasProperty is its current length
// check. Holes, prototypes, species and explicit thisArg remain unsupported.
func (l *lowering) libraryArrayPredicate(node *ast.Node, array ir.Expression, element ir.Type, name string) (ir.Expression, bool, error) {
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) != 1 {
		return nil, true, l.notYet(node, name+" with other than one callback")
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(written[0]), checker.SignatureKindCall)
	if len(signatures) != 1 || len(signatures[0].Parameters()) > 3 {
		return nil, true, l.notYet(node, name+" with an overloaded callback or more than three parameters")
	}
	returned := l.checker.GetReturnTypeOfSignature(signatures[0])
	returns, known := l.representation(returned)
	if known && returns == ir.Boolean {
		return l.arrayVisit(node, array, element, name)
	}
	isVoid := returned.Flags()&checker.TypeFlagsVoid != 0
	if isVoid {
		return nil, true, l.notYet(node, name+" with a void callback result (void can hide a returned value)")
	}
	if !known || returns == ir.Union || returns == ir.Weak {
		return nil, true, l.notYet(node, name+" with an unrepresented callback truth value")
	}
	if returns == ir.Object {
		// Structural object views such as {} also admit primitives. A pointer
		// presence test would incorrectly treat a boxed zero as truthy.
		for _, primitive := range []*checker.Type{l.checker.GetNumberType(), l.checker.GetStringType(), l.checker.GetBooleanType()} {
			if l.checker.IsTypeAssignableTo(primitive, returned) {
				return nil, true, l.notYet(node, name+" with an object callback result that can hide primitive truth values")
			}
		}
	}
	callback, err := l.expression(written[0])
	if err != nil {
		return nil, true, err
	}
	if callback.Type() != ir.Closure {
		return nil, true, l.notYet(node, name+" with a callback that isn't a function")
	}
	b := l.libraryArrayBuilder([]ir.Expression{array, callback})
	source := b.read(b.parameters[0])
	length := b.declare("length", ir.Length{Array: source})
	result := -1
	if name == "filter" {
		result = b.declare("filtered", ir.ArrayLiteral{Element: element})
	}
	index := b.declare("index", ir.NumberConstant{})
	input := ir.Expression(ir.ArrayIndex{Array: source, Index: b.read(index), Element: element})
	if input.Type().IsMaybe() && input.Type() != element {
		input = ir.Unwrap{Value: input}
	}
	item := b.local("element", element)
	// Keep kValue before the callback; filter writes that value even if the
	// callback replaces or removes the receiver's element.
	body := []ir.Statement{ir.Declare{Local: item, Value: input}}
	arguments := []ir.Expression{b.read(item), b.read(index), source}
	for i, parameter := range signatures[0].Parameters() {
		of, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
		if !known || slotless(of) {
			return nil, true, l.notYet(node, name+" with an unrepresented callback parameter")
		}
		arguments[i] = fit(arguments[i], of)
		if arguments[i].Type() != of {
			return nil, true, l.notYet(node, name+" with an incompatible callback parameter")
		}
	}
	call := ir.CallClosure{Closure: b.read(b.parameters[1]), Arguments: arguments, Returns: returns}
	selected := ir.Expression(ir.BooleanConstant{Value: false})
	answer := b.local("answer", returns)
	body = append(body, ir.Declare{Local: answer, Value: call})
	if returned.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {
		selected = libraryArrayToBoolean(b.read(answer))
	}
	if name == "filter" {
		body = append(body, ir.If{Condition: selected, Then: []ir.Statement{ir.Evaluate{Value: ir.ArrayPush{Array: b.read(result), Value: b.read(item), Element: element, Site: l.writeSite(node)}}}})
	} else {
		answer := name == "some"
		if !answer {
			selected = ir.Unary{Operator: ir.Not, Operand: selected}
		}
		body = append(body, ir.If{Condition: selected, Then: []ir.Statement{ir.Return{Value: ir.BooleanConstant{Value: answer}}}})
	}
	b.body = append(b.body, ir.Loop{Condition: ir.Binary{Operator: ir.Less, Left: b.read(index), Right: b.read(length)}, Body: []ir.Statement{ir.If{Condition: ir.Binary{Operator: ir.Less, Left: b.read(index), Right: ir.Length{Array: source}}, Then: body}}, Update: []ir.Statement{ir.Assign{Local: index, Value: ir.Binary{Operator: ir.Add, Left: b.read(index), Right: ir.NumberConstant{Value: 1}}}}})
	final := ir.Expression(ir.BooleanConstant{Value: name == "every"})
	if name == "filter" {
		final = b.read(result)
	}
	return b.finish("array_predicate_"+name, final), true, nil
}

// ToBoolean has no user coercion hooks. It never calls an object's valueOf.
func libraryArrayToBoolean(value ir.Expression) ir.Expression {
	if value.Type().IsMaybe() {
		return ir.Conditional{Condition: ir.IsUndefined{Value: value}, WhenTrue: ir.BooleanConstant{Value: false}, WhenNot: libraryArrayToBoolean(ir.Unwrap{Value: value})}
	}
	switch value.Type() {
	case ir.Boolean:
		return value
	case ir.Number:
		return ir.Binary{Operator: ir.And, Left: ir.Binary{Operator: ir.NotEqual, Left: value, Right: ir.NumberConstant{}}, Right: ir.Unary{Operator: ir.Not, Operand: ir.NumberCall{Function: "isNaN", Arguments: []ir.Expression{value}}}}
	case ir.String:
		return ir.Conditional{Condition: ir.IsUndefined{Value: value}, WhenTrue: ir.BooleanConstant{Value: false}, WhenNot: ir.Binary{Operator: ir.Greater, Left: ir.StringLength{Value: value}, Right: ir.NumberConstant{}}}
	default:
		return ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: value}}
	}
}

func (l *lowering) isArrayPredicateCall(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	return call.QuestionDotToken == nil && len(call.Arguments.Nodes) == 1 && callee.Kind == ast.KindPropertyAccessExpression && callee.AsPropertyAccessExpression().QuestionDotToken == nil && callee.Name().Text() == "isArray" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Array") && l.libraryMember(callee)
}

func (l *lowering) arrayIsArray(node *ast.Node) (ir.Expression, bool, error) {
	if !l.isArrayPredicateCall(node) {
		return nil, false, nil
	}
	argument := node.AsCallExpression().Arguments.Nodes[0]
	proven := l.checker.GetTypeAtLocation(argument)
	of, represented := l.representation(proven)
	// Keep the library's effect-preserving static facts, including empty literals
	// and intrinsic identities. Mixed slots use the compiler's runtime brand test.
	static := represented && of != ir.Union && proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsUnknown|checker.TypeFlagsAny|checker.TypeFlagsTypeParameter) == 0 && !isClassInstance(proven) && !checker.IsTupleType(proven)
	plain := ast.SkipParentheses(argument)
	if plain.Kind == ast.KindArrayLiteralExpression && len(plain.AsArrayLiteralExpression().Elements.Nodes) == 0 {
		static = true
	}
	for _, global := range []string{"Object", "Array", "String", "Number", "JSON", "Math", "Map", "Set"} {
		if l.isLibraryGlobal(plain, global) {
			static = true
		}
	}
	if plain.Kind == ast.KindPropertyAccessExpression && plain.Name().Text() == "prototype" && l.isLibraryGlobal(plain.AsPropertyAccessExpression().Expression, "Array") {
		static = true
	}
	if static {
		return l.libraryArrayIsArray(node)
	}
	if !l.arrayPredicateDomain(l.checker.GetTypeAtLocation(argument)) && !l.exactObject(argument, 0) {
		return nil, true, l.notYet(argument, "Array.isArray on a tuple or an erased object/any/unknown view")
	}
	value, err := l.expression(argument)
	if err != nil {
		return nil, true, err
	}
	return ir.ArrayIsArray{Value: fit(value, ir.Union)}, true, nil
}

// The library's any[] predicate describes the runtime test, not an element
// contract. Only array members already promised by the declared input survive.
func (l *lowering) arrayPredicateMembers(proven *checker.Type, array bool) *checker.Type {
	members := []*checker.Type{}
	var collect func(*checker.Type)
	collect = func(member *checker.Type) {
		member = l.concrete(member)
		if member.Flags()&checker.TypeFlagsUnion != 0 {
			for _, part := range member.Types() {
				collect(part)
			}
			return
		}
		if member.Flags()&checker.TypeFlagsUnknown != 0 {
			if array {
				members = append(members, predicateReadonlyArray(l.checker, member, true))
			} else {
				members = append(members, member)
			}
			return
		}
		if member.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter) != 0 {
			return
		}
		if l.checker.IsArrayType(member) == array {
			members = append(members, member)
		}
	}
	collect(proven)
	return l.checker.GetUnionType(members)
}

// Tuples currently have an object representation. Erased object views can hide
// one too, so the native array brand cannot decide their JavaScript arrayness.
func (l *lowering) arrayPredicateDomain(proven *checker.Type) bool {
	proven = l.concrete(proven)
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if !l.arrayPredicateDomain(member) {
				return false
			}
		}
		return true
	}
	if proven.Flags()&checker.TypeFlagsUnknown != 0 {
		return true
	}
	if proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsTypeParameter|checker.TypeFlagsIntersection) != 0 || checker.IsTupleType(proven) {
		return false
	}
	if l.checker.IsArrayType(proven) {
		return true
	}
	return proven.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(proven) || l.isLibraryType(proven, "Uint8Array", "Int32Array", "Float64Array")
}

// TypeScript's built-in signature is value is any[]. Recover the declared
// array member at a direct, dominating positive test instead of adopting any.
func (l *lowering) arrayPredicateType(node *ast.Node) *checker.Type {
	node = ast.SkipParentheses(node)
	actual := l.checker.GetTypeAtLocation(node)
	if node.Kind != ast.KindIdentifier {
		return actual
	}
	symbol := l.symbol(node)
	if symbol == nil || node.FlowNodeData() == nil {
		return actual
	}
	declared := l.checker.GetTypeOfSymbol(symbol)
	if !l.arrayPredicateDomain(declared) && declared.Flags()&checker.TypeFlagsUnknown == 0 {
		return actual
	}
	flow := node.FlowNodeData().FlowNode
	for steps := 0; flow != nil && steps < 256; steps++ {
		if flow.Flags&(ast.FlowFlagsAssignment|ast.FlowFlagsLabel|ast.FlowFlagsCall) != 0 {
			return actual
		}
		if flow.Flags&ast.FlowFlagsCondition != 0 && flow.Node != nil && l.arrayPredicateFlowCall(flow.Node) {
			call := ast.SkipParentheses(flow.Node).AsCallExpression()
			if l.symbol(ast.SkipParentheses(call.Arguments.Nodes[0])) == symbol {
				return l.arrayPredicateMembers(declared, flow.Flags&ast.FlowFlagsTrueCondition != 0)
			}
		}
		flow = flow.Antecedent
	}
	return actual
}

// Indexed reads and their ?? result retain the recovered element contract,
// including the missing element required by noUncheckedIndexedAccess.
func (l *lowering) arrayPredicateObservedType(node *ast.Node) *checker.Type {
	node = ast.SkipParentheses(node)
	actual := l.checker.GetTypeAtLocation(node)
	if node.Kind == ast.KindIdentifier {
		return l.arrayPredicateType(node)
	}
	if actual.Flags()&checker.TypeFlagsAny == 0 {
		return actual
	}
	switch node.Kind {
	case ast.KindElementAccessExpression:
		array := l.arrayPredicateType(node.AsElementAccessExpression().Expression)
		if l.checker.IsArrayType(array) {
			element := l.checker.GetElementTypeOfArrayType(array)
			if element != nil && element.Flags()&checker.TypeFlagsAny == 0 {
				return l.checker.GetUnionType([]*checker.Type{element, l.checker.GetUndefinedType()})
			}
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindQuestionQuestionToken {
			left := l.arrayPredicateObservedType(binary.Left)
			if left.Flags()&checker.TypeFlagsAny == 0 {
				return l.checker.GetUnionType([]*checker.Type{l.checker.GetNonNullableType(left), l.checker.GetTypeAtLocation(binary.Right)})
			}
		}
	}
	return actual
}

// The pinned checker owns construction of the readonly array type, just as it
// owns predicate flow facts. This introduces no any element contract.
//
//go:linkname predicateReadonlyArray github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).createArrayTypeEx
func predicateReadonlyArray(receiver *checker.Checker, element *checker.Type, readonly bool) *checker.Type

// A proven wrapper has the same arrayness test as the builtin. Its annotation
// cannot supply element facts missing from the caller's declared union.
func (l *lowering) arrayPredicateFlowCall(node *ast.Node) bool {
	if l.isArrayPredicateCall(node) {
		return true
	}
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	if call.QuestionDotToken != nil || len(call.Arguments.Nodes) != 1 {
		return false
	}
	symbol := l.symbol(ast.SkipParentheses(call.Expression))
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindFunctionDeclaration || len(declaration.Parameters()) != 1 || declaration.Body() == nil || declaration.Type() == nil || declaration.Type().Kind != ast.KindTypePredicate {
			continue
		}
		target := l.checker.GetTypeFromTypeNode(declaration.Type().AsTypePredicateNode().Type)
		if !checker.Checker_isTypeIdenticalTo(l.checker, target, predicateReadonlyArray(l.checker, l.checker.GetUnknownType(), true)) {
			continue
		}
		body := declaration.Body()
		if body.Kind != ast.KindBlock || len(body.AsBlock().Statements.Nodes) != 1 {
			continue
		}
		statement := body.AsBlock().Statements.Nodes[0]
		if statement.Kind != ast.KindReturnStatement || statement.AsReturnStatement().Expression == nil {
			continue
		}
		returned := ast.SkipParentheses(statement.AsReturnStatement().Expression)
		if l.isArrayPredicateCall(returned) && l.symbol(ast.SkipParentheses(returned.AsCallExpression().Arguments.Nodes[0])) == l.symbol(declaration.Parameters()[0].Name()) {
			return true
		}
	}
	return false
}
