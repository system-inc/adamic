package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"strings"
)

// recordElement recognizes resolved mutable string dictionaries, including named keys.
func (lowering *lowering) recordElement(subjectType *checker.Type) *checker.Type {
	if subjectType == nil {
		return nil
	}
	subjectType = lowering.concrete(subjectType)
	if subjectType.Flags()&checker.TypeFlagsObject == 0 ||
		lowering.isArrayType(subjectType) ||
		checker.IsTupleType(subjectType) ||
		isClassInstance(subjectType) {
		return nil
	}
	if element := lowering.finitePartialRecordElement(subjectType); element != nil {
		return element
	}
	if base := lowering.pickedRecordBase(subjectType); base != nil {
		return lowering.recordElement(base)
	}
	if lowering.enumerationRecordStorageType(subjectType) {
		return nil
	}
	infos := lowering.checker.GetIndexInfosOfType(subjectType)
	for _, info := range infos {
		if declaration := info.Declaration(); declaration != nil {
			file := ast.GetSourceFileOfNode(declaration)
			if load.IsLibrary(file) && strings.Contains(string(file.AsSourceFile().FileName()), ".regexp.") {
				return nil
			}
		}
	}
	if len(infos) != 1 ||
		infos[0].KeyType().Flags() != checker.TypeFlagsString ||
		infos[0].IsReadonly() ||
		len(lowering.checker.GetSignaturesOfType(subjectType, checker.SignatureKindCall)) != 0 ||
		len(lowering.checker.GetSignaturesOfType(subjectType, checker.SignatureKindConstruct)) != 0 {
		return nil
	}
	return infos[0].ValueType()
}
func (lowering *lowering) recordLiteralElement(node *ast.Node) *checker.Type {
	if t := lowering.slotType(node); t != nil {
		if e := lowering.recordElement(lowering.checker.GetNonNullableType(t)); e != nil {
			return e
		}
		// A literal going into a union of records is built as the record its parameter reads.
		if e := lowering.recordUnionElement(lowering.checker.GetNonNullableType(t), ""); e != nil {
			return e
		}
	}
	return lowering.recordElement(lowering.checker.GetTypeAtLocation(node))
}
func (lowering *lowering) recordSlot(node *ast.Node) (ir.Type, error) {
	subjectType := lowering.recordElement(lowering.checker.GetTypeAtLocation(node))
	if subjectType == nil {
		return 0, lowering.notYet(node, "a record without a pure mutable string index signature")
	}
	of, known := lowering.recordStorageType(subjectType)
	if !known {
		return 0, lowering.notYet(node, "a record value of type "+lowering.checker.TypeToString(subjectType))
	}
	return of, nil
}
func (lowering *lowering) recordTarget(node *ast.Node) bool {
	n := ast.SkipParentheses(node)
	// The host owns environment entries; their reads use the host slice, not record storage.
	if _, environment := lowering.nodeEnvironmentAccess(n); environment {
		return false
	}
	if n.Kind == ast.KindElementAccessExpression {
		return lowering.recordElement(
			lowering.checker.GetTypeAtLocation(n.AsElementAccessExpression().Expression),
		) != nil
	}
	if n.Kind == ast.KindPropertyAccessExpression {
		return lowering.recordElement(
			lowering.checker.GetTypeAtLocation(n.AsPropertyAccessExpression().Expression),
		) != nil
	}
	return false
}

// recordUnionElement is the element of a union whose defined members are each record storage of one
// slot representation, with the named key held one way on every member. Such a value is the bare record
// pointer, so a read off it is a record read. Only reads and literals take it, never a write or a delete.
func (lowering *lowering) recordUnionElement(subjectType *checker.Type, name string) *checker.Type {
	if subjectType == nil || subjectType.Flags()&checker.TypeFlagsUnion == 0 {
		return nil
	}
	var first *checker.Type
	var shared, named ir.Type
	for index, member := range lowering.definedMembers(subjectType) {
		element := lowering.recordElement(member)
		if element == nil {
			return nil
		}
		of, known := lowering.recordStorageType(element)
		if !known {
			return nil
		}
		// A member without the key reads it through its index signature, which is another contract.
		contract := ir.Type(0)
		if name != "" {
			if property := lowering.checker.GetPropertyOfType(member, name); property != nil {
				contract, known = lowering.representation(lowering.checker.GetTypeOfSymbol(property))
				if !known {
					return nil
				}
			}
		}
		if index == 0 {
			first, shared, named = element, of, contract
		} else if of != shared || contract != named {
			return nil
		}
	}
	return first
}

// recordReadTarget is recordTarget, widened for a read to a union of like record storage.
func (lowering *lowering) recordReadTarget(node *ast.Node) bool {
	if lowering.recordTarget(node) {
		return true
	}
	n := ast.SkipParentheses(node)
	if n.Kind == ast.KindPropertyAccessExpression {
		return lowering.recordUnionElement(
			lowering.checker.GetTypeAtLocation(n.AsPropertyAccessExpression().Expression),
			n.Name().Text(),
		) != nil
	}
	return false
}

// indexSignatureStorageGap stops an index signature that is neither an enumeration's nor a record's
// storage.
func (lowering *lowering) indexSignatureStorageGap(node *ast.Node) error {
	if node.Kind == ast.KindIndexSignature && !lowering.enumerationIndexSignature(node) {
		if lowering.recordElement(lowering.checker.GetTypeAtLocation(node.Parent)) == nil {
			return lowering.notYet(
				node,
				"a non-mutable unrestricted string signature, or numeric, symbol, patterned or callable dictionary storage",
			)
		}
	}
	return nil
}

func init() {
	refusalRules.register(1200, "indexSignatureStorageGap", (*lowering).indexSignatureStorageGap)
}

// deleteRefusal refuses delete on anything but a record entry.
func (lowering *lowering) deleteRefusal(node *ast.Node) error {
	if node.Kind == ast.KindDeleteExpression &&
		!lowering.recordTarget(node.AsDeleteExpression().Expression) &&
		!lowering.nodeProcessEnvironmentDelete(node) {
		return &Refused{
			Where: lowering.program.Where(node),
			What:  "delete",
			Fix:   "fixed objects cannot lose fields; use a record or Map",
		}
	}
	return nil
}

func init() { refusalRules.register(1300, "deleteRefusal", (*lowering).deleteRefusal) }

func recordPrototypeName(name string) bool {
	switch name {
	case "constructor", "__defineGetter__", "__defineSetter__", "hasOwnProperty", "__lookupGetter__", "__lookupSetter__", "isPrototypeOf", "propertyIsEnumerable", "toString", "valueOf", "__proto__", "toLocaleString":
		return true
	}
	return false
}
func (lowering *lowering) recordKey(node *ast.Node, read bool) (ir.Expression, error) {
	n := ast.SkipParentheses(node)
	if read &&
		(n.Kind == ast.KindStringLiteral || n.Kind == ast.KindNoSubstitutionTemplateLiteral) &&
		recordPrototypeName(n.Text()) {
		return nil, &Refused{
			Where: lowering.program.Where(node),
			What:  "record key " + n.Text() + " names an Object.prototype member",
			Fix:   "use an own-property test and a supported own-key operation; inherited values are outside the record's value type",
		}
	}
	key, expressionError := lowering.expression(node)
	if expressionError != nil {
		return nil, expressionError
	}
	if key.Type() == ir.Number {
		return ir.NumberToString{Value: key}, nil
	}
	if key.Type() != ir.String {
		return nil, lowering.notYet(node, "a record key other than string or number")
	}
	return key, nil
}
func (lowering *lowering) recordAccess(node *ast.Node, read bool) (ir.Expression, ir.Expression, ir.Type, error) {
	n := ast.SkipParentheses(node)
	var receiver, keyNode *ast.Node
	var name string
	if n.Kind == ast.KindElementAccessExpression {
		expression := n.AsElementAccessExpression()
		if expression.QuestionDotToken != nil {
			return nil, nil, 0, lowering.notYet(node, "optional record indexing")
		}
		receiver, keyNode = expression.Expression, expression.ArgumentExpression
	} else {
		expression := n.AsPropertyAccessExpression()
		if expression.QuestionDotToken != nil {
			return nil, nil, 0, lowering.notYet(node, "optional record property access")
		}
		receiver = expression.Expression
		name = expression.Name().Text()
	}
	of, err := lowering.recordSlot(receiver)
	if element := lowering.recordUnionElement(
		lowering.checker.GetTypeAtLocation(receiver),
		name,
	); err != nil && element != nil && read {
		// Only recordReadTarget's reads reach here with a union; writes and deletes ask recordTarget.
		of, _ = lowering.recordStorageType(element)
		err = nil
	}
	if err != nil {
		return nil, nil, 0, err
	}
	expression, err := lowering.expression(receiver)
	if err != nil {
		return nil, nil, 0, err
	}
	var key ir.Expression
	if keyNode != nil {
		key, err = lowering.recordKey(keyNode, read)
	} else {
		if read && recordPrototypeName(name) {
			return nil, nil, 0, &Refused{
				Where: lowering.program.Where(node),
				What:  "record key " + name + " names an Object.prototype member",
				Fix:   "use own keys",
			}
		}
		key = ir.StringConstant{Index: lowering.constant(name)}
	}
	return expression, key, of, err
}
func (lowering *lowering) recordExpression(node *ast.Node) (ir.Expression, bool, error) {
	if lowering.recordTarget(node) ||
		lowering.recordReadTarget(node) && !lowering.recordReadDiscarded(node) && !lowering.recordReadOwnGuarded(node) {
		discarded := lowering.recordReadDiscarded(node)
		guarded := lowering.recordReadOwnGuarded(node)
		r, k, of, recordAccessError := lowering.recordAccess(node, !discarded && !guarded)
		if recordAccessError != nil {
			return nil, true, recordAccessError
		}
		var read ir.Expression = ir.RecordCall{
			Method:    "get",
			Arguments: []ir.Expression{r, k},
			Element:   of,
			Returns:   ir.Maybe(of),
		}
		if discarded {
			return lowering.recordOwnRead(r, k, of), true, nil
		}
		if guarded {
			read = lowering.recordOwnRead(r, k, of)
		}
		here := lowering.checker.GetTypeAtLocation(node)
		read, recordReadTypeError := lowering.recordReadType(node, read, here)
		if recordReadTypeError != nil {
			return nil, true, recordReadTypeError
		}
		if !lowering.includesUndefined(here) && !recordTagObservation(node) {
			if narrowed, known := lowering.representation(here); known && read.Type().IsMaybe() && !narrowed.IsMaybe() {
				read = ir.Unwrap{Value: read}
			}
			if read.Type().IsReference() && read.Type() != ir.Union {
				message := "undefined where the checker narrowed it away: a call since the narrowing put it back"
				if parent := node.Parent; parent != nil &&
					parent.Kind == ast.KindPropertyAccessExpression &&
					parent.AsPropertyAccessExpression().Expression == node {
					message = "TypeError: Cannot read properties of undefined (reading '" + parent.Name().Text() + "')"
				}
				read = ir.Defined{Value: read, Message: message}
			}
		}
		return read, true, nil
	}
	if node.Kind == ast.KindDeleteExpression && lowering.recordTarget(node.AsDeleteExpression().Expression) {
		r, k, of, recordAccessError := lowering.recordAccess(node.AsDeleteExpression().Expression, false)
		return ir.RecordCall{
			Method:    "delete",
			Arguments: []ir.Expression{r, k},
			Element:   of,
			Returns:   ir.Boolean,
		}, true, recordAccessError
	}
	if value, handled, recordScalarComparisonError := lowering.recordScalarComparison(node); handled {
		return value, true, recordScalarComparisonError
	}
	if node.Kind != ast.KindBinaryExpression {
		return nil, false, nil
	}
	binaryExpression := node.AsBinaryExpression()
	if binaryExpression.OperatorToken.Kind == ast.KindInKeyword &&
		lowering.recordElement(lowering.checker.GetTypeAtLocation(binaryExpression.Right)) != nil {
		k, recordKeyError := lowering.recordKey(binaryExpression.Left, true)
		if recordKeyError != nil {
			return nil, true, recordKeyError
		}
		r, expressionError := lowering.expression(binaryExpression.Right)
		return ir.RecordCall{
			Method:    "has",
			Arguments: []ir.Expression{k, r},
			Returns:   ir.Boolean,
		}, true, expressionError
	}
	if (binaryExpression.OperatorToken.Kind == ast.KindEqualsToken ||
		binaryExpression.OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken) &&
		lowering.recordTarget(binaryExpression.Left) {
		r, k, of, recordAccessError := lowering.recordAccess(
			binaryExpression.Left,
			binaryExpression.OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken,
		)
		if recordAccessError != nil {
			return nil, true, recordAccessError
		}
		v, expressionError := lowering.expression(binaryExpression.Right)
		if expressionError != nil {
			return nil, true, expressionError
		}
		if recordNamedWriteError := lowering.recordNamedWrite(
			binaryExpression.Left,
			binaryExpression.Right,
		); recordNamedWriteError != nil {
			return nil, true, recordNamedWriteError
		}
		returns := v.Type()
		v = fit(v, of)
		if v.Type() != of {
			return nil, true, lowering.notYet(node, "a record write with a different slot representation")
		}
		if binaryExpression.OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken {
			if of.IsMaybe() || of == ir.Union {
				return nil, true, lowering.notYet(node, "record ??= with a tagged optional value")
			}
			return ir.RecordCoalesce{
				Record:  r,
				Key:     k,
				Value:   v,
				Element: of,
				Site:    lowering.writeSite(recordReceiver(binaryExpression.Left)),
			}, true, nil
		}
		write := ir.Expression(ir.RecordCall{
			Method:    "set",
			Arguments: []ir.Expression{r, k, v},
			Element:   of,
			Returns:   of,
			Site:      lowering.writeSite(recordReceiver(binaryExpression.Left)),
		})
		if of == ir.Union && returns != of {
			write = ir.Narrow{Value: write, To: returns}
		} else if of.IsMaybe() && returns == of.Present() {
			write = ir.Unwrap{Value: write}
		}
		return write, true, nil
	}
	return nil, false, nil
}
func recordReceiver(node *ast.Node) *ast.Node {
	n := ast.SkipParentheses(node)
	if n.Kind == ast.KindElementAccessExpression {
		return n.AsElementAccessExpression().Expression
	}
	return n.AsPropertyAccessExpression().Expression
}
func (lowering *lowering) recordLiteral(node *ast.Node, subjectType *checker.Type) (ir.Expression, error) {
	of, known := lowering.recordStorageType(subjectType)
	if !known {
		return nil, lowering.notYet(node, "record values with an unsupported slot representation")
	}
	result := ir.RecordLiteral{Element: of, Site: lowering.writeSite(node)}
	for index, p := range node.AsObjectLiteralExpression().Properties.Nodes {
		if p.Kind == ast.KindSpreadAssignment {
			if index != 0 {
				return nil, &Refused{
					Where: lowering.program.Where(p),
					What:  "a spread after the first field",
					Fix:   "spread once, first (adamic/single-spread)",
				}
			}
			source := p.AsSpreadAssignment().Expression
			element := lowering.recordElement(lowering.checker.GetTypeAtLocation(source))
			if element == nil {
				return nil, lowering.notYet(p, "a spread into record storage from a fixed object")
			}
			if element != nil {
				if !lowering.sameRecordStorage(element, subjectType, map[[2]*checker.Type]bool{}) ||
					!lowering.sameKeeping(element, subjectType, map[[2]*checker.Type]bool{}) {
					return nil, lowering.notYet(p, "record spread values seen through different storage")
				}
				if widened := lowering.widened(element, subjectType, map[[2]*checker.Type]bool{}); widened != nil {
					return nil, &Refused{
						Where: lowering.program.Where(p),
						What:  "record spread shares values with an invariant-mutable slot that the target can widen",
						Fix:   "keep the value's mutable fields invariant or copy the values too (adamic/invariant-mutable)",
					}
				}
			}
			e, recordSlotError := lowering.recordSlot(source)
			if recordSlotError != nil {
				return nil, recordSlotError
			}
			if e != of {
				return nil, lowering.notYet(p, "record spread with a different value representation")
			}
			var expressionError error
			result.Spread, expressionError = lowering.expression(source)
			if expressionError != nil {
				return nil, expressionError
			}
			continue
		}
		if p.Kind != ast.KindPropertyAssignment && p.Kind != ast.KindShorthandPropertyAssignment {
			return nil, lowering.notYet(p, "a record literal member other than a data entry")
		}
		name := p.Name()
		var key ir.Expression
		var err error
		if name.Kind == ast.KindComputedPropertyName {
			if recordComputedEntryError := lowering.recordComputedEntry(node, p, of); recordComputedEntryError != nil {
				return nil, recordComputedEntryError
			}
			key, err = lowering.recordKey(name.AsComputedPropertyName().Expression, false)
		} else {
			if p.Kind == ast.KindPropertyAssignment && name.Text() == "__proto__" {
				return nil, lowering.notYet(p, "a prototype-setting record literal")
			}
			// A numeric name is the key its number's canonical string is, as JavaScript's ToPropertyKey
			// makes it: the scanner keeps a numeric literal's text as that string already (jsnum's
			// String, so 0x1F is "31", 1e3 is "1000" and 1.50 is "1.5"). Its name never goes through
			// l.expression, where the checker reads a property name as its property's type, not a number.
			key = ir.StringConstant{Index: lowering.constant(name.Text())}
		}
		if err != nil {
			return nil, err
		}
		var expression ir.Expression
		if p.Kind == ast.KindPropertyAssignment {
			expression, err = lowering.expression(p.AsPropertyAssignment().Initializer)
		} else {
			expression, err = lowering.shorthand(p)
		}
		if err != nil {
			return nil, err
		}
		expression = fit(expression, of)
		if expression.Type() != of {
			return nil, lowering.notYet(p, "a record literal value with a different representation")
		}
		result.Entries = append(result.Entries, ir.RecordEntry{Key: key, Value: expression})
	}
	return result, nil
}
func (lowering *lowering) recordObjectCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	return lowering.recordObjectCallArguments(node, name, node.AsCallExpression().Arguments.Nodes)
}

func (lowering *lowering) recordObjectCallArguments(
	node *ast.Node,
	name string,
	args []*ast.Node,
) (ir.Expression, bool, error) {
	if len(args) == 0 || lowering.recordElement(lowering.checker.GetTypeAtLocation(args[0])) == nil {
		return nil, false, nil
	}
	if name != "keys" && name != "values" && name != "entries" && name != "hasOwn" {
		if _, _, admitted := lowering.recordDefineShape(node); admitted {
			// As a statement, recordDefineStatement lowers it; its result, the object, isn't yet.
			return nil, true, lowering.notYet(node, "Object.defineProperty on a record whose result is used")
		}
		switch name {
		case "defineProperty", "defineProperties", "getOwnPropertyDescriptor", "getOwnPropertyDescriptors", "getPrototypeOf", "setPrototypeOf", "create":
			// Descriptors and prototypes on a record are refused as on any object (library_object.go, the
			// Expandos row): an expando or prototype surgery, never a shape we'll build. Only C4's one
			// define is admitted, above and in recordDefineStatement (Adamic's ruling on #3xq2xzc).
			return nil, false, nil
		}
		return nil, true, lowering.notYet(node, "Object."+name+" on a record")
	}
	// A Pick holds the whole record, so its values include fields it doesn't name, held otherwise.
	if (name == "values" || name == "entries") &&
		lowering.pickedRecordBase(lowering.checker.GetTypeAtLocation(args[0])) != nil {
		return nil, true, lowering.notYet(
			node,
			"Object."+name+" on a Pick of a record, which holds fields the Pick doesn't name",
		)
	}
	count := 1
	if name == "hasOwn" {
		count = 2
	}
	if len(args) != count {
		return nil, true, lowering.notYet(node, "record Object operation argument count")
	}
	of, recordSlotError := lowering.recordSlot(args[0])
	if recordSlotError != nil {
		return nil, true, recordSlotError
	}
	expression, expressionError := lowering.expression(args[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	result := ir.RecordCall{Method: name, Arguments: []ir.Expression{expression}, Element: of, Returns: ir.Array}
	if name == "hasOwn" {
		expression, recordKeyError := lowering.recordKey(args[1], false)
		if recordKeyError != nil {
			return nil, true, recordKeyError
		}
		result.Arguments = append(result.Arguments, expression)
		result.Returns = ir.Boolean
	}
	return result, true, nil
}

// Aliases cannot change the native storage kind, even when tsc accepts the structural view.
func (lowering *lowering) recordStorageView(node *ast.Node) error {
	source, target := lowering.recordViewSite(node)
	if source == nil {
		return nil
	}
	// A .ts view may drop a named contract its reads hold (droppedContractRead); .a refuses it at the
	// write (#bq3fvdt), in recordDeclaredKeyRefusal before this rule.
	lowering.droppedContractsRead = strings.HasSuffix(lowering.program.FileName(ast.GetSourceFileOfNode(node)), ".ts")
	defer func() { lowering.droppedContractsRead = false }()
	own := lowering.checker.GetTypeAtLocation(source)
	if stored, literal := lowering.branchesStoredAs(source, target); literal && stored {
		return nil
	} else if literal || !lowering.sameRecordStorage(own, target, map[[2]*checker.Type]bool{}) {
		return lowering.notYet(node, "a "+
			lowering.checker.TypeToString(own)+
			" seen as "+
			lowering.checker.TypeToString(target)+
			" (fixed objects and records have different storage; copy explicitly)")
	}
	return nil
}

// recordViewSite is the value a node puts into a typed slot and the slot's type, as recordStorageView
// judges them, or nil: an initializer, an as expression, or a view site, but an object literal, which
// is made as its slot, and an argument the library's own record operations take.
func (lowering *lowering) recordViewSite(node *ast.Node) (*ast.Node, *checker.Type) {
	var source *ast.Node
	var target *checker.Type
	if node.Kind == ast.KindVariableDeclaration {
		source = node.AsVariableDeclaration().Initializer
		if source != nil {
			target = lowering.checker.GetTypeAtLocation(node.Name())
		}
	} else if node.Kind == ast.KindAsExpression {
		source = node.AsAsExpression().Expression
		target = lowering.checker.GetTypeAtLocation(node)
	} else if lowering.isExpression(node) && viewSite(node) {
		source = node
		target = lowering.slotType(node)
	}
	if source == nil || target == nil || ast.SkipParentheses(source).Kind == ast.KindObjectLiteralExpression {
		return nil, nil
	}
	if lowering.recordLibraryArgument(source) {
		return nil, nil // Intrinsic adapters validate the actual storage, not their erased checker parameter.
	}
	if parent := source.Parent; parent != nil && parent.Kind == ast.KindCallExpression {
		if lowering.detachedOwnCallReceiver(parent) != nil {
			return nil, nil
		}
		callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
		if lowering.isLibraryGlobal(callee, "String") {
			return nil, nil
		}
		if callee.Kind == ast.KindPropertyAccessExpression && (lowering.isLibraryGlobal(
			callee.AsPropertyAccessExpression().Expression,
			"Object",
		) || lowering.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "JSON")) {
			return nil, nil
		}
	}
	return source, target
}

func init() { refusalRules.register(2000, "recordStorageView", (*lowering).recordStorageView) }

// branchesStoredAs judges a source that can evaluate to an object literal one value at a time: each arm
// of a conditional, and an arrow's expression body against the target's return. A literal is stored as
// what objectLiteral builds it as (recordLiteralElement: its slot's record, else its own type), never
// as its own type alone, so a literal built fixed meets the target as fixed. literal reports whether
// any arm was an object literal; when none was, the source's own type is judged as before.
func (lowering *lowering) branchesStoredAs(source *ast.Node, target *checker.Type) (stored, literal bool) {
	node := ast.SkipParentheses(source)
	switch node.Kind {
	case ast.KindObjectLiteralExpression:
		built := lowering.checker.GetTypeAtLocation(node)
		if slot := lowering.slotType(node); slot != nil &&
			lowering.recordLiteralElement(node) != nil &&
			lowering.recordElement(lowering.checker.GetNonNullableType(slot)) != nil {
			built = lowering.checker.GetNonNullableType(slot)
		}
		return lowering.sameRecordStorage(built, target, map[[2]*checker.Type]bool{}), true
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		whenTrue, trueLiteral := lowering.branchesStoredAs(conditional.WhenTrue, target)
		whenFalse, falseLiteral := lowering.branchesStoredAs(conditional.WhenFalse, target)
		return whenTrue && whenFalse, trueLiteral || falseLiteral
	case ast.KindArrowFunction:
		body := node.Body()
		if body == nil || body.Kind == ast.KindBlock {
			break
		}
		own, viewed := lowering.checker.GetSignaturesOfType(
			lowering.checker.GetTypeAtLocation(node),
			checker.SignatureKindCall,
		), lowering.checker.GetSignaturesOfType(lowering.checker.GetNonNullableType(target), checker.SignatureKindCall)
		if len(own) != 1 || len(viewed) != 1 {
			break
		}
		returned, bodyLiteral := lowering.branchesStoredAs(body, lowering.checker.GetReturnTypeOfSignature(viewed[0]))
		if !bodyLiteral {
			break
		}
		ownParameters, viewedParameters := own[0].Parameters(), viewed[0].Parameters()
		parameters := lowering.parameterStorage(func() bool {
			for index := 0; index < len(ownParameters) && index < len(viewedParameters); index++ {
				// A caller passes the viewed parameter's value into the arrow's own: that's the flow.
				if !lowering.sameRecordStorage(
					lowering.checker.GetTypeOfSymbol(viewedParameters[index]),
					lowering.checker.GetTypeOfSymbol(ownParameters[index]),
					map[[2]*checker.Type]bool{},
				) {
					return false
				}
			}
			return true
		})
		return returned && parameters, true
	}
	return lowering.sameRecordStorage(
		lowering.checker.GetTypeAtLocation(node),
		target,
		map[[2]*checker.Type]bool{},
	), false
}

// Structural views must preserve dictionary storage at every depth, including callbacks.
func (lowering *lowering) sameRecordStorage(from, to *checker.Type, visited map[[2]*checker.Type]bool) bool {
	if from == nil || to == nil || from == to || visited[[2]*checker.Type{from, to}] {
		return true
	}
	if from.Flags()&checker.TypeFlagsUnion != 0 {
		visited[[2]*checker.Type{from, to}] = true
		for _, member := range lowering.definedMembers(from) {
			if member.Flags()&checker.TypeFlagsNull == 0 && !lowering.sameRecordStorage(member, to, visited) {
				return false
			}
		}
		return true
	}
	if to.Flags()&checker.TypeFlagsUnion != 0 {
		visited[[2]*checker.Type{from, to}] = true
		members := lowering.definedMembers(to)
		// A value enters a union that names its own type as that member, whatever other members it's
		// assignable to: CompilerOptions in CompilerOptions | PrinterOptions stays a record, and a read on
		// the union dispatches on the value it holds.
		for _, member := range members {
			if member == from {
				return true
			}
		}
		for _, member := range members {
			if lowering.checker.IsTypeAssignableTo(from, member) && !lowering.sameRecordStorage(from, member, visited) {
				return false
			}
		}
		return true
	}
	from, to = lowering.present(from), lowering.present(to)
	// A Pick of a record is stored as the whole record, so it's compared as that record.
	if base := lowering.pickedRecordBase(from); base != nil {
		from = base
	}
	if base := lowering.pickedRecordBase(to); base != nil {
		to = base
	}
	if from == nil || to == nil || from == to || visited[[2]*checker.Type{from, to}] {
		return true
	}
	visited[[2]*checker.Type{from, to}] = true
	inside, viewed := lowering.recordElement(from), lowering.recordElement(to)
	if inside != nil || viewed != nil {
		return inside != nil &&
			viewed != nil &&
			lowering.sameRecordNamedContracts(from, to, inside, viewed, visited) &&
			lowering.sameRecordStorage(inside, viewed, visited)
	}
	a, b := lowering.checker.GetSignaturesOfType(
		from,
		checker.SignatureKindCall,
	), lowering.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	if len(a) > 0 && len(b) > 0 {
		ap, bp := a[0].Parameters(), b[0].Parameters()
		parameters := lowering.parameterStorage(func() bool {
			for index := 0; index < len(ap) && index < len(bp); index++ {
				if !lowering.sameRecordStorage(
					lowering.checker.GetTypeOfSymbol(ap[index]),
					lowering.checker.GetTypeOfSymbol(bp[index]),
					visited,
				) {
					return false
				}
			}
			return true
		})
		return parameters && lowering.sameRecordStorage(
			lowering.checker.GetReturnTypeOfSignature(a[0]),
			lowering.checker.GetReturnTypeOfSignature(b[0]),
			visited,
		)
	}
	if lowering.isArrayType(from) ||
		checker.IsTupleType(from) ||
		lowering.isLibraryType(from, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
		a, b := lowering.typeArguments(from), lowering.typeArguments(to)
		for index := 0; index < len(a) && index < len(b); index++ {
			if !lowering.sameRecordStorage(a[index], b[index], visited) {
				return false
			}
		}
		return true
	}
	// Two instances of one generic type store alike when their arguments do. Walking their members
	// instead expands Promise's then and catch into fresh instantiations at every level, which no
	// visited pair ever repeats: 1.3 million calls, 294,120 deep, for Promise<() => void>. The
	// arguments decide outright, failing closed when they differ, since a pair this walk marked
	// visited would read as alike to any member walk after it.
	if from.ObjectFlags()&to.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
		from.Target() != nil &&
		from.Target() == to.Target() {
		a, b := lowering.checker.GetTypeArguments(from), lowering.checker.GetTypeArguments(to)
		if len(a) != len(b) {
			return false
		}
		for index := range a {
			if !lowering.sameRecordStorage(a[index], b[index], visited) {
				return false
			}
		}
		return true
	}
	for _, property := range lowering.checker.GetPropertiesOfType(to) {
		if own := lowering.checker.GetPropertyOfType(from, property.Name); own != nil {
			if !lowering.sameRecordStorage(
				lowering.checker.GetTypeOfSymbol(own),
				lowering.checker.GetTypeOfSymbol(property),
				visited,
			) {
				return false
			}
		}
	}
	return true
}

// Catch literal reads before method/builtin dispatch can bypass ordinary value lowering.
func (lowering *lowering) recordLiteralRead(node *ast.Node) error {
	if !lowering.recordReadTarget(node) || lowering.recordReadDiscarded(node) || lowering.recordReadOwnGuarded(node) {
		return nil
	}
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	if parent := at.Parent; parent != nil {
		if parent.Kind == ast.KindDeleteExpression {
			return nil
		}
		if parent.Kind == ast.KindBinaryExpression {
			expression := parent.AsBinaryExpression()
			if expression.Left == at && expression.OperatorToken.Kind == ast.KindEqualsToken {
				return nil
			}
		}
	}
	if parent := at.Parent; parent != nil && parent.Kind == ast.KindBinaryExpression {
		// The scalar comparison owns the read and proves its inherited value unobservable.
		expression := parent.AsBinaryExpression()
		other := expression.Right
		if expression.Right == at {
			other = expression.Left
		}
		subjectType := lowering.checker.GetTypeAtLocation(other)
		if (expression.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken ||
			expression.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken) &&
			subjectType.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike) != 0 &&
			!lowering.includesUndefined(subjectType) {
			return nil
		}
	}
	var name string
	if node.Kind == ast.KindPropertyAccessExpression {
		name = node.Name().Text()
	} else {
		key := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if key.Kind == ast.KindStringLiteral || key.Kind == ast.KindNoSubstitutionTemplateLiteral {
			name = key.Text()
		}
	}
	if recordPrototypeName(name) {
		return &Refused{
			Where: lowering.program.Where(node),
			What:  "record key " + name + " names an Object.prototype member",
			Fix:   "use own keys",
		}
	}
	return nil
}

func init() { refusalRules.register(1100, "recordLiteralRead", (*lowering).recordLiteralRead) }
