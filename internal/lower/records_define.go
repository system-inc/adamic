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
func (lowering *lowering) recordDefineShape(node *ast.Node) (string, *ast.Node, bool) {
	if node.Kind != ast.KindCallExpression {
		return "", nil, false
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if call.QuestionDotToken != nil ||
		callee.Kind != ast.KindPropertyAccessExpression ||
		callee.AsPropertyAccessExpression().QuestionDotToken != nil ||
		callee.Name().Text() != "defineProperty" ||
		!lowering.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
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
func (lowering *lowering) recordDefineStatement(node *ast.Node) ([]ir.Statement, bool, error) {
	key, valueNode, admitted := lowering.recordDefineShape(node)
	if !admitted {
		return nil, false, nil
	}
	receiver := node.AsCallExpression().Arguments.Nodes[0]
	proven := lowering.checker.GetTypeAtLocation(receiver)
	if lowering.recordElement(proven) == nil {
		return nil, false, nil
	}
	if lowering.includesUndefined(proven) || lowering.includesNull(proven) {
		return nil, true, lowering.notYet(receiver, "Object.defineProperty on a record that may be null or undefined")
	}
	of, recordSlotError := lowering.recordSlot(receiver)
	if recordSlotError != nil {
		return nil, true, recordSlotError
	}
	if recordDefineValueFitsError := lowering.recordDefineValueFits(
		proven,
		valueNode,
		key,
	); recordDefineValueFitsError != nil {
		return nil, true, recordDefineValueFitsError
	}
	record, expressionError := lowering.expression(receiver)
	if expressionError != nil {
		return nil, true, expressionError
	}
	if record.Type() != ir.Record {
		return nil, true, lowering.notYet(receiver, "Object.defineProperty on dictionary storage that isn't a record's")
	}
	value, expressionError := lowering.expression(valueNode)
	if expressionError != nil {
		return nil, true, expressionError
	}
	value = fit(value, of)
	if value.Type() != of {
		return nil, true, lowering.notYet(
			valueNode,
			"Object.defineProperty with a value of a different representation than the record's slots",
		)
	}
	if lowering.result.HiddenRecordKeys == nil {
		lowering.result.HiddenRecordKeys = ir.SortedMap[string, bool]{}
	}
	lowering.result.HiddenRecordKeys[key] = true
	define := ir.RecordCall{
		Method:    "defineHidden",
		Arguments: []ir.Expression{record, ir.StringConstant{Index: lowering.constant(key)}, value},
		Element:   of,
		Returns:   ir.Boolean,
		Site:      lowering.writeSite(receiver),
	}
	redefined := ir.MakeError{
		Message: ir.StringConstant{Index: lowering.constant("Cannot redefine property: " + key)},
		Name:    ir.StringConstant{Index: lowering.constant("TypeError")},
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
func (lowering *lowering) recordDefineValueFits(holder *checker.Type, value *ast.Node, key string) error {
	if property := lowering.checker.GetPropertyOfType(holder, key); property != nil {
		if namedWriteStopError := lowering.namedWriteStop(
			value,
			holder,
			value,
			map[string]bool{key: true},
		); namedWriteStopError != nil {
			return namedWriteStopError
		}
		lowering.keepSlot(value, lowering.checker.GetNonMissingTypeOfSymbol(property))
		return nil
	}
	var element *checker.Type
	for _, info := range lowering.checker.GetIndexInfosOfType(holder) {
		if info.KeyType().Flags() == checker.TypeFlagsString {
			element = info.ValueType()
		}
	}
	if element == nil {
		return lowering.notYet(
			value,
			"Object.defineProperty of a key the record neither declares nor takes through a string index signature",
		)
	}
	from := lowering.checker.GetTypeAtLocation(value)
	if !lowering.checker.IsTypeAssignableTo(from, element) ||
		!lowering.sameRecordStorage(from, element, map[[2]*checker.Type]bool{}) {
		return lowering.notYet(
			value,
			"Object.defineProperty with a value the record's index signature doesn't take as it stores its values",
		)
	}
	if refuseInSlotError := lowering.refuseInSlot(value, element); refuseInSlotError != nil {
		return refuseInSlotError
	}
	held, known := lowering.representation(from)
	if valueSlotStopsError := lowering.valueSlotStops(
		value,
		element,
		!known || held.IsReference(),
	); valueSlotStopsError != nil {
		return valueSlotStopsError
	}
	lowering.keepSlot(value, element)
	return nil
}
