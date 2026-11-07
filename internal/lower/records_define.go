package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// recordDefineShape is the key and value of Object.defineProperty(receiver, "key", { enumerable: false,
// writable: false, value }), the one descriptor a record entry takes (tsc's commandLineParser hides
// options.configFile so). The key is a string literal, and the descriptor an object literal of exactly
// those three data properties, in any order, enumerable and writable the literal false. Anything else,
// and a call that isn't this one, is not admitted.
func (l *lowering) recordDefineShape(node *ast.Node) (string, *ast.Node, bool) {
	if node.Kind != ast.KindCallExpression {
		return "", nil, false
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if call.QuestionDotToken != nil ||
		callee.Kind != ast.KindPropertyAccessExpression ||
		callee.AsPropertyAccessExpression().QuestionDotToken != nil ||
		callee.Name().Text() != "defineProperty" ||
		!l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
		return "", nil, false
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) != 3 {
		return "", nil, false
	}
	key := ast.SkipParentheses(call.Arguments.Nodes[1])
	descriptor := ast.SkipParentheses(call.Arguments.Nodes[2])
	if key.Kind != ast.KindStringLiteral || descriptor.Kind != ast.KindObjectLiteralExpression {
		return "", nil, false
	}
	fields := map[string]*ast.Node{}
	for _, property := range descriptor.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment {
			return "", nil, false
		}
		name := property.Name()
		if name.Kind != ast.KindIdentifier && name.Kind != ast.KindStringLiteral {
			return "", nil, false
		}
		if _, twice := fields[name.Text()]; twice {
			return "", nil, false
		}
		fields[name.Text()] = property.AsPropertyAssignment().Initializer
	}
	literalFalse := func(value *ast.Node) bool {
		return value != nil && ast.SkipParentheses(value).Kind == ast.KindFalseKeyword
	}
	if len(fields) != 3 ||
		!literalFalse(fields["enumerable"]) ||
		!literalFalse(fields["writable"]) ||
		fields["value"] == nil {
		return "", nil, false
	}
	return key.Text(), fields["value"], true
}

// recordDefineStatement lowers an admitted define, written as a statement, on proven record storage:
// the key is hidden (adamic_record_define_hidden), and a define it refuses throws Node's TypeError.
// Its key joins the program's hidden keys (ir.Program's HiddenRecordKeys), which decide the record
// writes and deletes that test for one: lowering is one pass over every module, and the exception
// analysis and the backends read the set only once it's done. A define on a fixed
// object isn't handled here, and Object.defineProperty's own refusal stops it. The call's result, the
// object, is never read as a statement, so a define whose result is used stays not yet
// (recordObjectCallArguments).
func (l *lowering) recordDefineStatement(node *ast.Node) ([]ir.Statement, bool, error) {
	key, valueNode, admitted := l.recordDefineShape(node)
	if !admitted {
		return nil, false, nil
	}
	receiver := node.AsCallExpression().Arguments.Nodes[0]
	proven := l.checker.GetTypeAtLocation(receiver)
	if l.recordElement(proven) == nil {
		return nil, false, nil
	}
	if l.includesUndefined(proven) || l.includesNull(proven) {
		return nil, true, l.notYet(receiver, "Object.defineProperty on a record that may be null or undefined")
	}
	of, recordSlotError := l.recordSlot(receiver)
	if recordSlotError != nil {
		return nil, true, recordSlotError
	}
	if recordDefineValueFitsError := l.recordDefineValueFits(
		proven,
		valueNode,
		key,
	); recordDefineValueFitsError != nil {
		return nil, true, recordDefineValueFitsError
	}
	record, expressionError := l.expression(receiver)
	if expressionError != nil {
		return nil, true, expressionError
	}
	if record.Type() != ir.Record {
		return nil, true, l.notYet(receiver, "Object.defineProperty on dictionary storage that isn't a record's")
	}
	value, expressionError := l.expression(valueNode)
	if expressionError != nil {
		return nil, true, expressionError
	}
	value = fit(value, of)
	if value.Type() != of {
		return nil, true, l.notYet(
			valueNode,
			"Object.defineProperty with a value of a different representation than the record's slots",
		)
	}
	if l.result.HiddenRecordKeys == nil {
		l.result.HiddenRecordKeys = ir.SortedMap[string, bool]{}
	}
	l.result.HiddenRecordKeys[key] = true
	define := ir.RecordCall{
		Method:    "defineHidden",
		Arguments: []ir.Expression{record, ir.StringConstant{Index: l.constant(key)}, value},
		Element:   of,
		Returns:   ir.Boolean,
		Site:      l.writeSite(receiver),
	}
	redefined := ir.MakeError{
		Message: ir.StringConstant{Index: l.constant("Cannot redefine property: " + key)},
		Name:    ir.StringConstant{Index: l.constant("TypeError")},
	}
	return []ir.Statement{
		ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: define}, Then: []ir.Statement{ir.Throw{Value: redefined}}},
	}, true, nil
}

// recordDefineValueFits judges the descriptor's value, which lib types any, so the checker never held
// it to the record's slot. It's judged as a write of it under key: against the named property key
// names, when the record declares one (namedWriteStop), and otherwise against the index signature's
// value type by the same rules. The slot stays set while the value is lowered, for lowering's own
// checks at it.
func (l *lowering) recordDefineValueFits(holder *checker.Type, value *ast.Node, key string) error {
	if property := l.checker.GetPropertyOfType(holder, key); property != nil {
		if namedWriteStopError := l.namedWriteStop(
			value,
			holder,
			value,
			map[string]bool{key: true},
		); namedWriteStopError != nil {
			return namedWriteStopError
		}
		l.keepSlot(value, l.checker.GetNonMissingTypeOfSymbol(property))
		return nil
	}
	var element *checker.Type
	for _, info := range l.checker.GetIndexInfosOfType(holder) {
		if info.KeyType().Flags() == checker.TypeFlagsString {
			element = info.ValueType()
		}
	}
	if element == nil {
		return l.notYet(
			value,
			"Object.defineProperty of a key the record neither declares nor takes through a string index signature",
		)
	}
	from := l.checker.GetTypeAtLocation(value)
	if !l.checker.IsTypeAssignableTo(from, element) ||
		!l.sameRecordStorage(from, element, map[[2]*checker.Type]bool{}) {
		return l.notYet(
			value,
			"Object.defineProperty with a value the record's index signature doesn't take as it stores its values",
		)
	}
	if refuseInSlotError := l.refuseInSlot(value, element); refuseInSlotError != nil {
		return refuseInSlotError
	}
	held, known := l.representation(from)
	if valueSlotStopsError := l.valueSlotStops(
		value,
		element,
		!known || held.IsReference(),
	); valueSlotStopsError != nil {
		return valueSlotStopsError
	}
	l.keepSlot(value, element)
	return nil
}
