package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Recognize declaration ancestry, never a coincidental type name or structural
// array-like object. Element and extra storage are then resolved by the checker.
func (l *lowering) nodeArrayBase(proven *checker.Type, seen map[*checker.Type]bool) *checker.Type {
	if proven == nil || seen[proven] || proven.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	seen[proven] = true
	if l.isLibraryType(proven, "Array", "ReadonlyArray") {
		return proven
	}
	target := proven
	if proven.ObjectFlags()&checker.ObjectFlagsReference != 0 && proven.Target() != nil {
		target = proven.Target()
	}
	if target.ObjectFlags()&checker.ObjectFlagsInterface == 0 {
		return nil
	}
	for _, base := range l.checker.GetBaseTypes(target) {
		if found := l.nodeArrayBase(base, seen); found != nil {
			return found
		}
	}
	return nil
}

func (l *lowering) nodeArrayLayoutOf(proven *checker.Type) (ArrayLayout, *checker.Type, bool) {
	if proven == nil || l.isArrayType(proven) || checker.IsTupleType(proven) {
		return ArrayLayout{}, nil, false
	}
	base := l.nodeArrayBase(proven, map[*checker.Type]bool{})
	if base == nil {
		return ArrayLayout{}, nil, false
	}
	element := l.checker.GetIndexTypeOfType(proven, l.checker.GetNumberType())
	if element == nil {
		return ArrayLayout{}, nil, false
	}
	if l.nodeArrayBase(element, map[*checker.Type]bool{}) != nil && !l.isArrayType(element) {
		return ArrayLayout{}, nil, false
	}
	held, known := l.representation(element)
	if !known || slotless(held) {
		return ArrayLayout{}, nil, false
	}
	extras := []FieldContract{}
	for _, property := range l.checker.GetPropertiesOfType(proven) {
		if l.checker.GetPropertyOfType(base, property.Name) != nil {
			continue
		}
		declared := l.checker.GetTypeOfSymbol(property)
		of, known := l.representation(declared)
		if !known {
			return ArrayLayout{}, nil, false
		}
		extras = append(extras, FieldContract{Name: property.Name, DeclaredType: l.checker.TypeToString(declared), Of: of, Optional: property.Flags&ast.SymbolFlagsOptional != 0})
	}
	layout, err := NodeArrayLayout(held, extras)
	return layout, element, err == nil
}

func (l *lowering) nodeArrayCast(node *ast.Node) bool {
	as := node.AsAsExpression()
	if ast.SkipParentheses(as.Expression).Kind != ast.KindArrayLiteralExpression {
		return false
	}
	layout, element, known := l.nodeArrayLayoutOf(l.concrete(l.checker.GetTypeAtLocation(node)))
	if !known || len(layout.Fields) == 0 {
		return false
	}
	actual := l.elementTypeOfArray(l.checker.GetTypeAtLocation(as.Expression))
	// Mutable array elements keep their representation and their full element contract.
	return actual != nil && l.castAssignable(actual, element) && l.castAssignable(element, l.checker.GetWidenedType(actual)) && l.sameKeeping(actual, element, map[[2]*checker.Type]bool{})
}

// graphNodeArraySite is one NodeArray literal: where, and its NodeArray type.
type graphNodeArraySite struct {
	node  *ast.Node
	array *checker.Type
}

func (l *lowering) constructNodeArray(node *ast.Node) (ir.Expression, error) {
	arrayType := l.concrete(l.checker.GetTypeAtLocation(node))
	layout, element, _ := l.nodeArrayLayoutOf(arrayType)
	l.graphNodeArrays = append(l.graphNodeArrays, graphNodeArraySite{node: node, array: arrayType})
	value, err := l.arrayLiteral(ast.SkipParentheses(node.AsAsExpression().Expression))
	if err != nil {
		return nil, err
	}
	literal, known := value.(ir.ArrayLiteral)
	if !known {
		return nil, l.notYet(node, "NodeArray construction from a nonliteral array operation")
	}
	held, _ := l.representation(element)
	if literal.Element != held {
		return nil, l.notYet(node, "NodeArray element storage conversion")
	}
	for _, field := range layout.Fields {
		zero := ir.Expression(zeroValue(field.Of))
		if field.Of.IsReference() {
			zero = ir.Undefined{Of: field.Of}
		}
		literal.Metadata = append(literal.Metadata, ir.Field{Name: field.Name, Value: zero, Absent: true})
	}
	return literal, nil
}

func (l *lowering) nodeArrayProperty(node *ast.Node, object ir.Expression, name string) (ir.Expression, bool, error) {
	receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression))
	// A field of NodeArray<E> & readonly [T1..Tn] is its NodeArray part's, in the one storage (nodeArrayTuple).
	if parts, isTuple := l.nodeArrayTuple(l.concrete(receiver)); isTuple {
		receiver = parts.nodeArray
	}
	layout, _, known := l.nodeArrayLayoutOf(receiver)
	if !known {
		return nil, false, nil
	}
	field, err := layout.Read(name)
	if err != nil {
		return nil, true, l.notYet(node, "a NodeArray field outside its fixed metadata layout")
	}
	declared := l.checker.GetTypeOfSymbol(l.checker.GetSymbolAtLocation(node.Name()))
	value := ir.DynamicProperty{Object: fit(object, ir.Union), Name: name}
	checked, err := l.unsetSpecialization(node, value, declared, "field '"+name+"', factory 'NodeArray allocation'")
	_ = field // Layout membership is proved independently of observed presence.
	return checked, true, err
}

// nodeArrayTupleParts is NodeArray<E> & readonly [T1..Tn] taken apart: the NodeArray part, its element E, and
// each position's Ti.
type nodeArrayTupleParts struct {
	nodeArray, element *checker.Type
	positions          []*checker.Type
}

// nodeArrayTuple is Adamic's ruling 3 on the scout pass (#bbssdcy, Nov 24, families 1 D and 13 B): tsc's
// predicates narrow a NodeArray to NodeArray<E> & readonly [T1..Tn], each Ti extending E, and that intersection
// is held as its NodeArray part's own layout (nodeArrayLayoutOf), one value in one storage, never a dictionary
// (objectIntersection) and never a layout of its own. representationReason holds it as an array, its length is
// the array's, a field read is the NodeArray part's (nodeArrayProperty), and nodeArrayTupleRead reads its
// positions. nodeArrayLayoutOf itself still answers no for it, so every road that asks for an array's element
// type (iteration, the array methods, spread, destructuring) and every write through it keeps the stop it had
// before. Anything else the rule can't decide is no intersection of this kind, and stops where it did: a tuple
// part that isn't readonly or has an optional, rest or variadic element, a third part that isn't implied by the
// NodeArray part (nodeArrayImpliedMember), a Ti that doesn't extend E or is held otherwise, and an E or a Ti that
// admits undefined or null, whose absence a position read couldn't tell from a missing element.
func (l *lowering) nodeArrayTuple(proven *checker.Type) (nodeArrayTupleParts, bool) {
	if proven == nil || proven.Flags()&checker.TypeFlagsIntersection == 0 {
		return nodeArrayTupleParts{}, false
	}
	parts := nodeArrayTupleParts{}
	var tuple *checker.Type
	var element *checker.Type
	others := []*checker.Type{}
	for _, part := range proven.Types() {
		switch {
		case checker.IsTupleType(part) && tuple == nil:
			tuple = part
		case checker.IsTupleType(part):
			return nodeArrayTupleParts{}, false
		case parts.nodeArray == nil:
			if _, partElement, known := l.nodeArrayLayoutOf(part); known {
				parts.nodeArray, element = part, partElement
				continue
			}
			others = append(others, part)
		default:
			others = append(others, part)
		}
	}
	if tuple == nil || parts.nodeArray == nil || !tuple.TargetTupleType().IsReadonly() || checker.TupleType_combinedFlags(tuple.TargetTupleType())&checker.ElementFlagsNonRequired != 0 {
		return nodeArrayTupleParts{}, false
	}
	for _, member := range others {
		if !l.nodeArrayImpliedMember(parts.nodeArray, member) {
			return nodeArrayTupleParts{}, false
		}
	}
	if l.includesUndefined(element) || l.includesNull(element) {
		return nodeArrayTupleParts{}, false
	}
	held, _ := l.representation(element)
	count := len(tuple.TargetTupleType().ElementFlags())
	arguments := l.typeArguments(tuple)
	if count == 0 || len(arguments) < count {
		return nodeArrayTupleParts{}, false
	}
	for _, position := range arguments[:count] {
		position = l.concrete(position)
		if !l.checker.IsTypeAssignableTo(position, element) || l.includesUndefined(position) || l.includesNull(position) {
			return nodeArrayTupleParts{}, false
		}
		if of, known := l.representation(position); !known || of != held {
			return nodeArrayTupleParts{}, false
		}
		parts.positions = append(parts.positions, position)
	}
	parts.element = element
	return parts, true
}

// nodeArrayImpliedMember is Adamic's answer on #tjfrgyq (Nov 25 00:20Z): a part of NodeArray<E> & readonly
// [T1..Tn] beside those two adds no constraint and no slot when every property of it is the NodeArray part's own,
// readonly on both sides and at an identical type, as tsc's Readonly<TextRange> is beside NodeArray's
// ReadonlyTextRange pos and end (types.ts:3056, BindableObjectDefinePropertyCall's arguments). A member that
// declares any property mutable, optional where the NodeArray part's isn't, at another type or absent from the
// NodeArray part isn't implied, and neither is one with a call, construct or index signature, or a tuple.
func (l *lowering) nodeArrayImpliedMember(nodeArray, member *checker.Type) bool {
	if member.Flags()&checker.TypeFlagsObject == 0 || checker.IsTupleType(member) {
		return false
	}
	if len(l.checker.GetSignaturesOfType(member, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(member, checker.SignatureKindConstruct)) != 0 || len(l.checker.GetIndexInfosOfType(member)) != 0 {
		return false
	}
	for _, property := range l.checker.GetPropertiesOfType(member) {
		own := l.checker.GetPropertyOfType(nodeArray, property.Name)
		if own == nil || property.Flags&ast.SymbolFlagsOptional != own.Flags&ast.SymbolFlagsOptional || !l.checker.IsReadonlySymbol(property) || !l.checker.IsReadonlySymbol(own) {
			return false
		}
		if !checker.Checker_isTypeIdenticalTo(l.checker, l.checker.GetTypeOfSymbol(property), l.checker.GetTypeOfSymbol(own)) {
			return false
		}
	}
	return true
}

// nodeArrayTupleRead is x[i] on NodeArray<E> & readonly [T1..Tn] (nodeArrayTuple), handled when x is one. Under
// ruling 3 as Adamic amended it by @system_adamic_language's must (#bbssdcy, Nov 24 16:4xZ), a length test proves
// a position is present, never that its element is still Ti: readonly binds the view, not the array, and a
// mutable alias of the backing array can swap element 0 for another E without changing the length. Language's
// t01 (internal/oracle/testdata/node_array_tuple/t01.ts) is that program, and t01-alias.ts beside it is the same
// program on a construction the subset admits: a four-field MutableNodeArray literal its caller keeps, handed
// out as NodeArray. So, at a literal position i below n:
//
//   - where Ti is E, the read is the NodeArray's own element read when a length test covering i (or the
//     predicate that made the intersection, which tests the length) dominates it with no call and no write
//     between, and a checked read otherwise;
//   - where Ti is narrower than E, the read is a checked view read of Ti, which stops at exit 70 with a located
//     message when the element isn't a Ti, unless the proof of Ti itself, that predicate or a kind test of that
//     position, dominates it with no call and no write between. A length test alone never skips this check.
//
// Any other read, a computed index, a position the tuple part doesn't hold or an optional one, stops.
func (l *lowering) nodeArrayTupleRead(node *ast.Node, object ir.Expression, optional bool) (ir.Expression, bool, error) {
	access := node.AsElementAccessExpression()
	receiver := l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(access.Expression)))
	parts, isTuple := l.nodeArrayTuple(receiver)
	if !isTuple || object.Type() != ir.Array {
		return nil, false, nil
	}
	key := ast.SkipParentheses(access.ArgumentExpression)
	if optional || key.Kind != ast.KindNumericLiteral {
		return nil, true, l.notYet(node, "a read of "+l.checker.TypeToString(receiver)+" other than at a literal position")
	}
	position, err := strconv.Atoi(key.Text())
	if err != nil || position < 0 || position >= len(parts.positions) {
		return nil, true, l.notYet(node, "a read of "+l.checker.TypeToString(receiver)+" at a position its tuple part doesn't hold")
	}
	held, _ := l.representation(parts.element)
	read := ir.ArrayIndex{Array: object, Index: ir.NumberConstant{Value: float64(position)}, Element: held}
	at := parts.positions[position]
	if l.nodeArrayPositionProven(node, access.Expression, receiver, position, at, checker.Checker_isTypeIdenticalTo(l.checker, at, parts.element)) {
		return read, true, nil
	}
	id, err := l.viewContract(node, at)
	if err != nil {
		return nil, true, err
	}
	view := sourceExpression(node) + " at " + l.program.Where(node)
	expected := l.checker.TypeToString(at)
	checked := ir.ViewElement{Array: object, Index: read.Index, Contract: id, Where: l.program.Where(node), View: view, ViewType: expected, Of: held, Element: held, Position: true}
	// The check hands out undefined where the alias shortened the array, which is no Ti either. Where the read is a
	// property read's receiver, Node throws there, so the stop says Node's TypeError, as defined's does.
	absent := "element read failed: " + view + " is not a " + expected + "; expected " + expected + ", found undefined"
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if parent := outer.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == outer && !l.questionDot(parent) {
		absent = "TypeError: Cannot read properties of undefined (reading '" + parent.Name().Text() + "')"
	}
	return ir.Defined{Value: checked, Message: absent}, true, nil
}

// nodeArrayPositionProven is whether a proof dominates read, receiver[position] on intersection, with no call and
// no write between (codeBetween) and no assignment of receiver: a test read takes the true side of, an if's
// condition, a && chain's left or a conditional's condition, in read's own function and inside any loop that
// holds read, since a loop's next pass runs what follows the read before the read again. A conjunct of the test
// proves the position when it's a call of a predicate claiming intersection for receiver itself, or, where the
// position's type is E (same), a length test covering it, or, where it's narrower, a test of that position's
// kind against its literal. receiver must be a name; any other receiver is never proven.
func (l *lowering) nodeArrayPositionProven(read, receiver *ast.Node, intersection *checker.Type, position int, at *checker.Type, same bool) bool {
	receiver = ast.SkipParentheses(receiver)
	if receiver.Kind != ast.KindIdentifier {
		return false
	}
	symbol := l.valueSymbol(receiver)
	if symbol == nil {
		return false
	}
	names := func(node *ast.Node) bool {
		node = ast.SkipParentheses(node)
		return node.Kind == ast.KindIdentifier && l.valueSymbol(node) == symbol
	}
	proves := func(conjunct *ast.Node) bool {
		conjunct = ast.SkipParentheses(conjunct)
		if conjunct.Kind == ast.KindCallExpression {
			call := conjunct.AsCallExpression()
			signature := l.checker.GetResolvedSignature(conjunct)
			if signature == nil {
				return false
			}
			claim := predicateOfSignature(l.checker, signature)
			if claim == nil || claim.Kind() != checker.TypePredicateKindIdentifier || claim.Type() == nil || claim.ParameterIndex() < 0 || int(claim.ParameterIndex()) >= len(call.Arguments.Nodes) {
				return false
			}
			return names(call.Arguments.Nodes[claim.ParameterIndex()]) && checker.Checker_isTypeIdenticalTo(l.checker, l.concrete(claim.Type()), intersection)
		}
		if conjunct.Kind != ast.KindBinaryExpression {
			return false
		}
		binary := conjunct.AsBinaryExpression()
		operator := binary.OperatorToken.Kind
		left, right := ast.SkipParentheses(binary.Left), ast.SkipParentheses(binary.Right)
		if left.Kind != ast.KindPropertyAccessExpression {
			return false
		}
		field := left.AsPropertyAccessExpression()
		if l.questionDot(left) || l.inOptionalChain(left) {
			return false
		}
		if same {
			if field.Name().Text() != "length" || !names(field.Expression) || right.Kind != ast.KindNumericLiteral {
				return false
			}
			bound, err := strconv.ParseFloat(right.Text(), 64)
			if err != nil {
				return false
			}
			switch operator {
			case ast.KindEqualsEqualsEqualsToken, ast.KindEqualsEqualsToken, ast.KindGreaterThanEqualsToken:
				return bound > float64(position)
			case ast.KindGreaterThanToken:
				return bound >= float64(position)
			}
			return false
		}
		tag := l.fieldLiteral(at, "kind")
		if tag == nil || field.Name().Text() != "kind" || (operator != ast.KindEqualsEqualsEqualsToken && operator != ast.KindEqualsEqualsToken) {
			return false
		}
		element := ast.SkipParentheses(field.Expression)
		if element.Kind != ast.KindElementAccessExpression || l.questionDot(element) || !names(element.AsElementAccessExpression().Expression) {
			return false
		}
		index := ast.SkipParentheses(element.AsElementAccessExpression().ArgumentExpression)
		if index.Kind != ast.KindNumericLiteral || index.Text() != strconv.Itoa(position) {
			return false
		}
		compared := l.checker.GetTypeAtLocation(right)
		return compared.Flags()&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBooleanLiteral) != 0 && l.checker.IsTypeAssignableTo(compared, tag) && l.checker.IsTypeAssignableTo(tag, compared)
	}
	untouched := func(from int) bool {
		if l.codeBetween(from, read, nil) {
			return false
		}
		assigned := false
		l.visitBetween(from, read, func(node *ast.Node) {
			if node.Kind == ast.KindBinaryExpression && ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) && names(node.AsBinaryExpression().Left) {
				assigned = true
			}
		})
		return !assigned
	}
	var conjuncts func(test *ast.Node) []*ast.Node
	conjuncts = func(test *ast.Node) []*ast.Node {
		test = ast.SkipParentheses(test)
		if test.Kind == ast.KindBinaryExpression && test.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
			return append(conjuncts(test.AsBinaryExpression().Left), conjuncts(test.AsBinaryExpression().Right)...)
		}
		return []*ast.Node{test}
	}
	provenBy := func(test *ast.Node) bool {
		for _, conjunct := range conjuncts(test) {
			if proves(conjunct) && untouched(conjunct.End()) {
				return true
			}
		}
		return false
	}
	child := read
	for parent := read.Parent; parent != nil; child, parent = parent, parent.Parent {
		if ast.IsFunctionLike(parent) || ast.IsClassLike(parent) || parent.Kind == ast.KindSourceFile {
			return false
		}
		switch parent.Kind {
		case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement, ast.KindWhileStatement, ast.KindDoStatement, ast.KindLabeledStatement:
			return false
		case ast.KindIfStatement:
			if child == parent.AsIfStatement().ThenStatement && provenBy(parent.AsIfStatement().Expression) {
				return true
			}
		case ast.KindBinaryExpression:
			if binary := parent.AsBinaryExpression(); binary.OperatorToken.Kind == ast.KindAmpersandAmpersandToken && child == binary.Right && provenBy(binary.Left) {
				return true
			}
		case ast.KindConditionalExpression:
			if conditional := parent.AsConditionalExpression(); child == conditional.WhenTrue && provenBy(conditional.Condition) {
				return true
			}
		}
	}
	return false
}

// nodeArrayTupleMembership is predicateMembership's test that value, held as an array, is NodeArray<E> & readonly
// [T1..Tn] (nodeArrayTuple): its length is n, and each position whose Ti is narrower than the source's element is
// a Ti by that type's own membership, read once the length has said the position is there. It's the predicate's
// proof of every Ti, which nodeArrayTupleRead takes in place of a position's check only where nothing runs between
// it and the read. A source not held as an array, or a position whose own membership needs statements ahead of
// the length test, stops.
func (l *lowering) nodeArrayTupleMembership(node *ast.Node, value ir.Expression, source *checker.Type, parts nodeArrayTupleParts, local func(string, ir.Type) int, depth int) ([]ir.Statement, ir.Expression, error) {
	_, element, known := l.nodeArrayLayoutOf(source)
	if !known && l.isArrayType(source) {
		element, known = l.elementTypeOfArray(source), true
	}
	if value.Type() != ir.Array || !known || element == nil {
		return nil, nil, l.notYet(node, "a predicate narrowing "+l.checker.TypeToString(source)+" to "+l.checker.TypeToString(parts.nodeArray)+" and a readonly tuple, from a value not held as an array")
	}
	held, known := l.representation(element)
	if expected, _ := l.representation(parts.element); !known || held != expected {
		return nil, nil, l.notYet(node, "a predicate narrowing "+l.checker.TypeToString(source)+" to a NodeArray and a readonly tuple whose element is held otherwise")
	}
	var test ir.Expression = ir.Binary{Operator: ir.And, Left: censusCondition(value), Right: ir.Binary{Operator: ir.Equal, Left: ir.Length{Array: value}, Right: ir.NumberConstant{Value: float64(len(parts.positions))}}}
	for position, at := range parts.positions {
		if checker.Checker_isTypeIdenticalTo(l.checker, l.concrete(element), at) {
			continue
		}
		statements, member, err := l.predicateMembership(node, ir.ArrayIndex{Array: value, Index: ir.NumberConstant{Value: float64(position)}, Element: held}, element, at, local, depth+1)
		if err != nil {
			return nil, nil, err
		}
		if len(statements) != 0 {
			return nil, nil, l.notYet(node, "a predicate narrowing to a NodeArray and a readonly tuple whose position "+strconv.Itoa(position)+" needs statements to test")
		}
		test = ir.Binary{Operator: ir.And, Left: test, Right: member}
	}
	return nil, test, nil
}
