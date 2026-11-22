package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"strings"
)

// recordElement recognizes resolved mutable string dictionaries, including named keys, and a readonly one the
// program holds as a record (readonlyRecordType).
func (l *lowering) recordElement(t *checker.Type) *checker.Type {
	if t == nil {
		return nil
	}
	t = l.concrete(t)
	if t.Flags()&checker.TypeFlagsObject == 0 || l.isArrayType(t) || checker.IsTupleType(t) || isClassInstance(t) {
		return nil
	}
	if element := l.finitePartialRecordElement(t); element != nil {
		return element
	}
	if base := l.pickedRecordBase(t); base != nil {
		return l.recordElement(base)
	}
	if l.enumerationRecordStorageType(t) {
		return nil
	}
	infos := l.checker.GetIndexInfosOfType(t)
	for _, info := range infos {
		if declaration := info.Declaration(); declaration != nil {
			file := ast.GetSourceFileOfNode(declaration)
			if load.IsLibrary(file) && strings.Contains(string(file.AsSourceFile().FileName()), ".regexp.") {
				return nil
			}
		}
	}
	if len(infos) != 1 || infos[0].KeyType().Flags() != checker.TypeFlagsString || infos[0].IsReadonly() && !l.readonlyRecordType(t) || len(l.checker.GetSignaturesOfType(t, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(t, checker.SignatureKindConstruct)) != 0 {
		return nil
	}
	return infos[0].ValueType()
}
func (l *lowering) recordLiteralElement(node *ast.Node) *checker.Type {
	if t := l.slotType(node); t != nil {
		if e := l.recordElement(l.checker.GetNonNullableType(t)); e != nil {
			return e
		}
		// A literal going into a union of records is built as the record its parameter reads.
		if e := l.recordUnionElement(l.checker.GetNonNullableType(t), ""); e != nil {
			return e
		}
		if e := l.recordUnionDictionaryMember(node, l.checker.GetNonNullableType(t)); e != nil {
			return e
		}
	}
	return l.recordElement(l.checker.GetTypeAtLocation(node))
}

// recordUnionDictionaryMember is the element of the one dictionary member of a union slot a literal fits, when
// it fits no fixed-object member, so the literal is built as that dictionary: tsc's { k: ["a"] } into
// CompilerOptionsValue is a MapLike<string[]>. Built fixed, a read that narrows the slot's value to its
// dictionary member (recordReadHeapMember) found a fixed object and stopped on an honest value, Language's g5
// p06 on #6c4swvt. A literal that fits a fixed-object member too, or two dictionaries, is built as before, as is
// one in a library's union (Object.entries' { [s: string]: T } | ArrayLike<T>), whose adapter reads the
// value's own storage.
func (l *lowering) recordUnionDictionaryMember(literal *ast.Node, slot *checker.Type) *checker.Type {
	if slot.Flags()&checker.TypeFlagsUnion == 0 {
		return nil
	}
	own := l.checker.GetTypeAtLocation(literal)
	var found *checker.Type
	for _, member := range l.definedMembers(slot) {
		if member.Flags()&checker.TypeFlagsObject == 0 || !l.checker.IsTypeAssignableTo(own, member) {
			continue
		}
		element := l.recordElement(member)
		if element == nil || found != nil {
			return nil
		}
		// Only a dictionary the program declares: a library's (Object.entries' { [s: string]: T }, Record) is a
		// parameter its own adapter reads by the value's actual storage.
		symbol := l.unbranded(member).Symbol()
		if symbol == nil || len(symbol.Declarations) == 0 || load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0])) || load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
			return nil
		}
		found = element
	}
	return found
}
func (l *lowering) recordSlot(node *ast.Node) (ir.Type, error) {
	t := l.recordElement(l.checker.GetTypeAtLocation(node))
	if t == nil {
		return 0, l.notYet(node, "a record without a pure mutable string index signature")
	}
	of, known := l.recordStorageType(t)
	if !known {
		return 0, l.notYet(node, "a record value of type "+l.checker.TypeToString(t))
	}
	return of, nil
}
func (l *lowering) recordTarget(node *ast.Node) bool {
	n := ast.SkipParentheses(node)
	// The host owns environment entries; their reads use the host slice, not record storage.
	if _, environment := l.nodeEnvironmentAccess(n); environment {
		return false
	}
	if n.Kind == ast.KindElementAccessExpression {
		return l.recordElement(l.checker.GetTypeAtLocation(n.AsElementAccessExpression().Expression)) != nil
	}
	if n.Kind == ast.KindPropertyAccessExpression {
		return l.recordElement(l.checker.GetTypeAtLocation(n.AsPropertyAccessExpression().Expression)) != nil
	}
	return false
}

// recordUnionElement is the element of a union whose defined members are each record storage of one
// slot representation, with the named key held one way on every member. Such a value is the bare record
// pointer, so a read off it is a record read. Only reads and literals take it, never a write or a delete.
func (l *lowering) recordUnionElement(t *checker.Type, name string) *checker.Type {
	if t == nil || t.Flags()&checker.TypeFlagsUnion == 0 {
		return nil
	}
	var first *checker.Type
	var shared, named ir.Type
	for index, member := range l.definedMembers(t) {
		element := l.recordElement(member)
		if element == nil {
			return nil
		}
		of, known := l.recordStorageType(element)
		if !known {
			return nil
		}
		// A member without the key reads it through its index signature, which is another contract.
		contract := ir.Type(0)
		if name != "" {
			if property := l.checker.GetPropertyOfType(member, name); property != nil {
				contract, known = l.representation(l.checker.GetTypeOfSymbol(property))
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
func (l *lowering) recordReadTarget(node *ast.Node) bool {
	if l.recordTarget(node) {
		return true
	}
	n := ast.SkipParentheses(node)
	if n.Kind == ast.KindPropertyAccessExpression {
		return l.recordUnionElement(l.checker.GetTypeAtLocation(n.AsPropertyAccessExpression().Expression), n.Name().Text()) != nil
	}
	return false
}

// indexSignatureStorageGap stops an index signature that is neither an enumeration's nor a record's
// storage.
func (l *lowering) indexSignatureStorageGap(node *ast.Node) error {
	if node.Kind == ast.KindIndexSignature && !l.enumerationIndexSignature(node) {
		if l.recordElement(l.checker.GetTypeAtLocation(node.Parent)) == nil {
			return l.notYet(node, "a non-mutable unrestricted string signature, or numeric, symbol, patterned or callable dictionary storage")
		}
	}
	return nil
}

func init() {
	refusalRules.register(1200, "indexSignatureStorageGap", (*lowering).indexSignatureStorageGap)
}

// deleteRefusal refuses delete on anything but a record entry.
func (l *lowering) deleteRefusal(node *ast.Node) error {
	if node.Kind == ast.KindDeleteExpression && !l.recordTarget(node.AsDeleteExpression().Expression) && !l.nodeProcessEnvironmentDelete(node) {
		return &Refused{Where: l.program.Where(node), What: "delete", Fix: "fixed objects cannot lose fields; use a record or Map"}
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
func (l *lowering) recordKey(node *ast.Node, read bool) (ir.Expression, error) {
	n := ast.SkipParentheses(node)
	if read && (n.Kind == ast.KindStringLiteral || n.Kind == ast.KindNoSubstitutionTemplateLiteral) && recordPrototypeName(n.Text()) {
		return nil, &Refused{Where: l.program.Where(node), What: "record key " + n.Text() + " names an Object.prototype member", Fix: "use an own-property test and a supported own-key operation; inherited values are outside the record's value type"}
	}
	key, err := l.expression(node)
	if err != nil {
		return nil, err
	}
	if key.Type() == ir.Number {
		return ir.NumberToString{Value: key}, nil
	}
	if key.Type() != ir.String {
		return nil, l.notYet(node, "a record key other than string or number")
	}
	return key, nil
}
func (l *lowering) recordAccess(node *ast.Node, read bool) (ir.Expression, ir.Expression, ir.Type, error) {
	n := ast.SkipParentheses(node)
	var receiver, keyNode *ast.Node
	var name string
	if n.Kind == ast.KindElementAccessExpression {
		a := n.AsElementAccessExpression()
		if a.QuestionDotToken != nil {
			return nil, nil, 0, l.notYet(node, "optional record indexing")
		}
		receiver, keyNode = a.Expression, a.ArgumentExpression
	} else {
		a := n.AsPropertyAccessExpression()
		if a.QuestionDotToken != nil {
			return nil, nil, 0, l.notYet(node, "optional record property access")
		}
		receiver = a.Expression
		name = a.Name().Text()
	}
	of, err := l.recordSlot(receiver)
	if element := l.recordUnionElement(l.checker.GetTypeAtLocation(receiver), name); err != nil && element != nil && read {
		// Only recordReadTarget's reads reach here with a union; writes and deletes ask recordTarget.
		of, _ = l.recordStorageType(element)
		err = nil
	}
	if err != nil {
		return nil, nil, 0, err
	}
	r, err := l.expression(receiver)
	if err != nil {
		return nil, nil, 0, err
	}
	var key ir.Expression
	if keyNode != nil {
		key, err = l.recordKey(keyNode, read)
	} else {
		if read && recordPrototypeName(name) {
			return nil, nil, 0, &Refused{Where: l.program.Where(node), What: "record key " + name + " names an Object.prototype member", Fix: "use own keys"}
		}
		key = ir.StringConstant{Index: l.constant(name)}
	}
	return r, key, of, err
}
func (l *lowering) recordExpression(node *ast.Node) (ir.Expression, bool, error) {
	if l.recordTarget(node) || l.recordReadTarget(node) && !l.recordReadDiscarded(node) && !l.recordReadOwnGuarded(node) {
		discarded := l.recordReadDiscarded(node)
		guarded := l.recordReadOwnGuarded(node)
		r, k, of, err := l.recordAccess(node, !discarded && !guarded)
		if err != nil {
			return nil, true, err
		}
		var read ir.Expression = ir.RecordCall{Method: "get", Arguments: []ir.Expression{r, k}, Element: of, Returns: ir.Maybe(of)}
		if discarded {
			return l.recordOwnRead(r, k, of), true, nil
		}
		if guarded {
			read = l.recordOwnRead(r, k, of)
		}
		here := l.checker.GetTypeAtLocation(node)
		read, err = l.recordReadType(node, read, here)
		if err != nil {
			return nil, true, err
		}
		if !l.includesUndefined(here) && !recordTagObservation(node) {
			// Under a ?. a reference's Defined would throw where Node reads undefined, and a catch
			// would take that throw, a branch Node never takes (#70kdt66); a scalar's Unwrap would
			// stop where Node reads undefined. A call TypeScript's narrowing doesn't see can delete
			// the key, so the read can be missing here. Until the ?. parent's own test reads it, it
			// stops. A union carries its own undefined to that test, with neither check.
			if read.Type() != ir.Union && optionalChainOperand(node) {
				return nil, true, l.notYet(node, "a record's value narrowed present and read under ?. (a call since the narrowing can delete its key, and the ?. would see it missing only after a TypeError)")
			}
			if narrowed, known := l.representation(here); known && read.Type().IsMaybe() && !narrowed.IsMaybe() {
				read = ir.Unwrap{Value: read}
			}
			if read.Type().IsReference() && read.Type() != ir.Union {
				message := "undefined where the checker narrowed it away: a call since the narrowing put it back"
				if parent := node.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == node {
					message = "TypeError: Cannot read properties of undefined (reading '" + parent.Name().Text() + "')"
				}
				read = ir.Defined{Value: read, Message: message}
			}
		}
		return read, true, nil
	}
	if node.Kind == ast.KindDeleteExpression && l.recordTarget(node.AsDeleteExpression().Expression) {
		r, k, of, err := l.recordAccess(node.AsDeleteExpression().Expression, false)
		return ir.RecordCall{Method: "delete", Arguments: []ir.Expression{r, k}, Element: of, Returns: ir.Boolean}, true, err
	}
	if value, handled, err := l.recordScalarComparison(node); handled {
		return value, true, err
	}
	if node.Kind != ast.KindBinaryExpression {
		return nil, false, nil
	}
	b := node.AsBinaryExpression()
	if b.OperatorToken.Kind == ast.KindInKeyword && l.recordElement(l.checker.GetTypeAtLocation(b.Right)) != nil {
		k, err := l.recordKey(b.Left, true)
		if err != nil {
			return nil, true, err
		}
		r, err := l.expression(b.Right)
		return ir.RecordCall{Method: "has", Arguments: []ir.Expression{k, r}, Returns: ir.Boolean}, true, err
	}
	if (b.OperatorToken.Kind == ast.KindEqualsToken || b.OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken) && l.recordTarget(b.Left) {
		r, k, of, err := l.recordAccess(b.Left, b.OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken)
		if err != nil {
			return nil, true, err
		}
		v, err := l.expression(b.Right)
		if err != nil {
			return nil, true, err
		}
		if err := l.recordNamedWrite(b.Left, b.Right); err != nil {
			return nil, true, err
		}
		returns := v.Type()
		v = fit(v, of)
		if v.Type() != of {
			return nil, true, l.notYet(node, "a record write with a different slot representation")
		}
		if b.OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken {
			if of.IsMaybe() || of == ir.Union {
				return nil, true, l.notYet(node, "record ??= with a tagged optional value")
			}
			return ir.RecordCoalesce{Record: r, Key: k, Value: v, Element: of, Site: l.writeSite(recordReceiver(b.Left))}, true, nil
		}
		write := ir.Expression(ir.RecordCall{Method: "set", Arguments: []ir.Expression{r, k, v}, Element: of, Returns: of, Site: l.writeSite(recordReceiver(b.Left))})
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
func (l *lowering) recordLiteral(node *ast.Node, t *checker.Type) (ir.Expression, error) {
	of, known := l.recordStorageType(t)
	if !known {
		return nil, l.notYet(node, "record values with an unsupported slot representation")
	}
	record := l.recordLiteralType(node)
	result := ir.RecordLiteral{Element: of, Site: l.writeSite(node)}
	// The record type it's built as is its graph identity, as an object literal's slot type is
	// (graphAllocation): a record of a type graphTypes selects is a graph member wherever it flows, and
	// graphFlows adds the literal's own site where a graph value is demanded through it (#1vb0x5s).
	if record != nil {
		result.GraphTypes = []int{l.cycleTypeID(l.concrete(record))}
	}
	// The record types noteRecordSlot asks a seed's slot to reach (reachedRecords). The empty fallback a
	// spread of R | undefined takes below is a second allocation, but it holds nothing and is only spread
	// into this literal, so it can't close a cycle by itself (Lower's read on #t5q4prw).
	l.graphRecords = append(l.graphRecords, graphRecordSite{node: node, record: record, empty: len(node.AsObjectLiteralExpression().Properties.Nodes) == 0})
	for i, p := range node.AsObjectLiteralExpression().Properties.Nodes {
		if p.Kind == ast.KindSpreadAssignment {
			if i != 0 {
				return nil, &Refused{Where: l.program.Where(p), What: "a spread after the first field", Fix: "spread once, first (adamic/single-spread)"}
			}
			source := p.AsSpreadAssignment().Expression
			// A source that may be undefined or null (tracing.ts's args: Args | undefined) is judged as
			// the record it holds when present, since JavaScript's { ...undefined } and { ...null } copy
			// nothing. Only this site strips them: recordElement's other callers read, write, delete or
			// walk the record itself, where undefined throws or isn't a record.
			own := l.checker.GetTypeAtLocation(source)
			present := own
			if own.Flags()&checker.TypeFlagsUnion != 0 {
				present = l.checker.GetNonNullableType(own)
			}
			mayBeMissing := present != own
			element := l.recordElement(present)
			if element == nil {
				return nil, l.notYet(p, "a spread into record storage from a fixed object")
			}
			if walk := newRecordStorageWalk(); !l.sameRecordStorage(element, t, walk) || !l.sameKeeping(element, t, map[[2]*checker.Type]bool{}) {
				if walk.bounded {
					return nil, l.recordStorageBounded(p, element, t)
				}
				return nil, l.notYet(p, "record spread values seen through different storage")
			}
			if widened := l.widened(element, t, map[[2]*checker.Type]bool{}); widened != nil {
				return nil, &Refused{Where: l.program.Where(p), What: "record spread shares values with an invariant-mutable slot that the target can widen", Fix: "keep the value's mutable fields invariant or copy the values too (adamic/invariant-mutable)"}
			}
			var e ir.Type
			var err error
			if mayBeMissing {
				// recordSlot asks the source's own type, which the missing arm keeps from being a record.
				var known bool
				if e, known = l.recordStorageType(element); !known {
					return nil, l.notYet(source, "a record value of type "+l.checker.TypeToString(element))
				}
			} else if e, err = l.recordSlot(source); err != nil {
				return nil, err
			}
			if e != of {
				return nil, l.notYet(p, "record spread with a different value representation")
			}
			result.Spread, err = l.expression(source)
			if err != nil {
				return nil, err
			}
			if mayBeMissing {
				// A missing record is a null pointer natively (or, where null may be held, the boxed
				// union's null), which the spread's walk can't read, so the source is tested as it runs:
				// present, it's spread as the record pointer the union holds bare; missing, an empty record
				// is, and nothing is copied. The source is evaluated once, as Coalesce's Value.
				if result.Spread.Type() != ir.Record && result.Spread.Type() != ir.Union {
					return nil, l.notYet(p, "a spread into record storage from a record that may be missing, held as other than a record or a union")
				}
				result.Spread = ir.Coalesce{Value: result.Spread, Fallback: ir.RecordLiteral{Element: of}, Of: ir.Record}
			}
			continue
		}
		if p.Kind != ast.KindPropertyAssignment && p.Kind != ast.KindShorthandPropertyAssignment {
			return nil, l.notYet(p, "a record literal member other than a data entry")
		}
		name := p.Name()
		var key ir.Expression
		var err error
		if name.Kind == ast.KindComputedPropertyName {
			if err := l.recordComputedEntry(node, p, of); err != nil {
				return nil, err
			}
			key, err = l.recordKey(name.AsComputedPropertyName().Expression, false)
		} else {
			if p.Kind == ast.KindPropertyAssignment && name.Text() == "__proto__" {
				return nil, l.notYet(p, "a prototype-setting record literal")
			}
			// A numeric name is the key its number's canonical string is, as JavaScript's ToPropertyKey
			// makes it: the scanner keeps a numeric literal's text as that string already (jsnum's
			// String, so 0x1F is "31", 1e3 is "1000" and 1.50 is "1.5"). Its name never goes through
			// l.expression, where the checker reads a property name as its property's type, not a number.
			key = ir.StringConstant{Index: l.constant(name.Text())}
		}
		if err != nil {
			return nil, err
		}
		var v ir.Expression
		if p.Kind == ast.KindPropertyAssignment {
			v, err = l.expression(p.AsPropertyAssignment().Initializer)
		} else {
			v, err = l.shorthand(p)
		}
		if err != nil {
			return nil, err
		}
		v = fit(v, of)
		if v.Type() != of {
			return nil, l.notYet(p, "a record literal value with a different representation")
		}
		result.Entries = append(result.Entries, ir.RecordEntry{Key: key, Value: v})
	}
	return result, nil
}
func (l *lowering) recordObjectCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	return l.recordObjectCallArguments(node, name, node.AsCallExpression().Arguments.Nodes)
}

func (l *lowering) recordObjectCallArguments(node *ast.Node, name string, args []*ast.Node) (ir.Expression, bool, error) {
	if len(args) == 0 || l.recordElement(l.checker.GetTypeAtLocation(args[0])) == nil {
		return nil, false, nil
	}
	if name != "keys" && name != "values" && name != "entries" && name != "hasOwn" {
		if _, _, admitted := l.recordDefineShape(node); admitted {
			// As a statement, recordDefineStatement lowers it; its result, the object, isn't yet.
			return nil, true, l.notYet(node, "Object.defineProperty on a record whose result is used")
		}
		switch name {
		case "defineProperty", "defineProperties", "getOwnPropertyDescriptor", "getOwnPropertyDescriptors", "getPrototypeOf", "setPrototypeOf", "create":
			// Descriptors and prototypes on a record are refused as on any object (library_object.go, the
			// Expandos row): an expando or prototype surgery, never a shape we'll build. Only C4's one
			// define is admitted, above and in recordDefineStatement (Adamic's ruling on #3xq2xzc).
			return nil, false, nil
		}
		return nil, true, l.notYet(node, "Object."+name+" on a record")
	}
	// A Pick holds the whole record, so its values include fields it doesn't name, held otherwise.
	if (name == "values" || name == "entries") && l.pickedRecordBase(l.checker.GetTypeAtLocation(args[0])) != nil {
		return nil, true, l.notYet(node, "Object."+name+" on a Pick of a record, which holds fields the Pick doesn't name")
	}
	count := 1
	if name == "hasOwn" {
		count = 2
	}
	if len(args) != count {
		return nil, true, l.notYet(node, "record Object operation argument count")
	}
	of, err := l.recordSlot(args[0])
	if err != nil {
		return nil, true, err
	}
	r, err := l.expression(args[0])
	if err != nil {
		return nil, true, err
	}
	result := ir.RecordCall{Method: name, Arguments: []ir.Expression{r}, Element: of, Returns: ir.Array}
	if name == "entries" {
		result.ElementNullish = l.slotNullish(l.recordElement(l.checker.GetTypeAtLocation(args[0])), of)
		l.unplannedAllocation(node, l.recordElement(l.checker.GetTypeAtLocation(args[0])), "a record's entries")
	}
	if name == "hasOwn" {
		k, err := l.recordKey(args[1], false)
		if err != nil {
			return nil, true, err
		}
		result.Arguments = append(result.Arguments, k)
		result.Returns = ir.Boolean
	}
	return result, true, nil
}

// Aliases cannot change the native storage kind, even when tsc accepts the structural view.
func (l *lowering) recordStorageView(node *ast.Node) error {
	source, target := l.recordViewSite(node)
	if source == nil {
		return nil
	}
	// A .ts view may drop a named contract its reads hold (droppedContractRead); .a refuses it at the
	// write (#bq3fvdt), in recordDeclaredKeyRefusal before this rule.
	l.droppedContractsRead = strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts")
	defer func() { l.droppedContractsRead = false }()
	own := l.checker.GetTypeAtLocation(source)
	stored, literal, bounded := l.branchesStoredAs(source, target)
	if literal && stored {
		return nil
	}
	if walk := newRecordStorageWalk(); literal || !l.sameRecordStorage(own, target, walk) {
		if literal && bounded || walk.bounded {
			return l.recordStorageBounded(node, own, target)
		}
		return l.notYet(node, "a "+l.checker.TypeToString(own)+" seen as "+l.checker.TypeToString(target)+" (fixed objects and records have different storage; copy explicitly"+l.readonlyRecordReason(target, source)+")")
	}
	return nil
}

// recordViewSite is the value a node puts into a typed slot and the slot's type, as recordStorageView
// judges them, or nil: an initializer, an as expression, a view site, or a shorthand property, but an
// object literal, which is made as its slot, and an argument the library's own record operations take.
func (l *lowering) recordViewSite(node *ast.Node) (*ast.Node, *checker.Type) {
	var source *ast.Node
	var target *checker.Type
	if node.Kind == ast.KindVariableDeclaration {
		source = node.AsVariableDeclaration().Initializer
		if source != nil {
			target = l.checker.GetTypeAtLocation(node.Name())
		}
	} else if node.Kind == ast.KindAsExpression {
		source = node.AsAsExpression().Expression
		target = l.checker.GetTypeAtLocation(node)
	} else if node.Kind == ast.KindShorthandPropertyAssignment {
		// { bag } puts the variable bag into the literal's property bag, the same site as { bag: bag }'s
		// initializer, which viewSite takes. The shorthand isn't an expression and viewSite never sees it,
		// so a fixed object went into a record slot here unjudged, and a write through the record ran
		// adamic_record_define on it (#mez92e7, Language's f02: an ASan segmentation fault in map.c where Node prints 1 3).
		source = node.Name()
		target = l.slotType(node.Name())
	} else if l.isExpression(node) && viewSite(node) {
		source = node
		target = l.slotType(node)
	}
	if source == nil || target == nil || ast.SkipParentheses(source).Kind == ast.KindObjectLiteralExpression {
		return nil, nil
	}
	if l.recordLibraryArgument(source) {
		return nil, nil // Intrinsic adapters validate the actual storage, not their erased checker parameter.
	}
	if parent := source.Parent; parent != nil && parent.Kind == ast.KindCallExpression {
		if l.detachedOwnCallReceiver(parent) != nil {
			return nil, nil
		}
		callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
		if l.isLibraryGlobal(callee, "String") {
			return nil, nil
		}
		if callee.Kind == ast.KindPropertyAccessExpression && (l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") || l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "JSON")) {
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
// any arm was an object literal; when none was, the source's own type is judged as before. bounded
// reports whether any arm's walk stopped at typeNesting's bound, so a false stored may mean only that
// the walk couldn't see to the end (#fzx0j5z).
func (l *lowering) branchesStoredAs(source *ast.Node, target *checker.Type) (stored, literal, bounded bool) {
	node := ast.SkipParentheses(source)
	switch node.Kind {
	case ast.KindObjectLiteralExpression:
		built := l.checker.GetTypeAtLocation(node)
		if slot := l.slotType(node); slot != nil && l.recordLiteralElement(node) != nil && l.recordElement(l.checker.GetNonNullableType(slot)) != nil {
			built = l.checker.GetNonNullableType(slot)
		}
		walk := newRecordStorageWalk()
		return l.sameRecordStorage(built, target, walk), true, walk.bounded
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		whenTrue, trueLiteral, trueBounded := l.branchesStoredAs(conditional.WhenTrue, target)
		whenFalse, falseLiteral, falseBounded := l.branchesStoredAs(conditional.WhenFalse, target)
		return whenTrue && whenFalse, trueLiteral || falseLiteral, trueBounded || falseBounded
	case ast.KindArrowFunction:
		body := node.Body()
		if body == nil || body.Kind == ast.KindBlock {
			break
		}
		own, viewed := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindCall), l.checker.GetSignaturesOfType(l.checker.GetNonNullableType(target), checker.SignatureKindCall)
		if len(own) != 1 || len(viewed) != 1 {
			break
		}
		returned, bodyLiteral, returnedBounded := l.branchesStoredAs(body, l.checker.GetReturnTypeOfSignature(viewed[0]))
		if !bodyLiteral {
			break
		}
		ownParameters, viewedParameters := own[0].Parameters(), viewed[0].Parameters()
		parameters := l.parameterStorage(func() bool {
			for i := 0; i < len(ownParameters) && i < len(viewedParameters); i++ {
				// A caller passes the viewed parameter's value into the arrow's own: that's the flow.
				if walk := newRecordStorageWalk(); !l.sameRecordStorage(l.checker.GetTypeOfSymbol(viewedParameters[i]), l.checker.GetTypeOfSymbol(ownParameters[i]), walk) {
					bounded = walk.bounded
					return false
				}
			}
			return true
		})
		return returned && parameters, true, returnedBounded || bounded
	}
	walk := newRecordStorageWalk()
	return l.sameRecordStorage(l.checker.GetTypeAtLocation(node), target, walk), false, walk.bounded
}

// recordStorageWalk is one sameRecordStorage comparison: the pairs it has seen, and how deep a
// generic that instantiates itself deeper at each level is open on each side's path (typeNesting).
// bounded is set when typeNesting refused a side, and the walk answers false from there to its top, so
// a false with bounded set means the walk couldn't see to the end, not that storage differs.
type recordStorageWalk struct {
	visited  map[[2]*checker.Type]bool
	from, to typeNesting
	bounded  bool
}

// newRecordStorageWalk starts a comparison. Every caller starts its own, so no answer leans on
// another comparison's visited pairs.
func newRecordStorageWalk() *recordStorageWalk {
	return &recordStorageWalk{visited: map[[2]*checker.Type]bool{}}
}

// recordStorageBounded is the stop for a value of type from seen as to whose storage comparison
// stopped at typeNesting's bound: a valid program whose types are only too deep for the walk, not yet
// rather than different storage (#fzx0j5z).
func (l *lowering) recordStorageBounded(node *ast.Node, from, to *checker.Type) error {
	return l.notYet(node, "comparing record storage through a type that instantiates itself deeper at each level ("+l.checker.TypeToString(from)+" seen as "+l.checker.TypeToString(to)+")")
}

// Structural views must preserve dictionary storage at every depth, including callbacks.
func (l *lowering) sameRecordStorage(from, to *checker.Type, walk *recordStorageWalk) bool {
	if from == nil || to == nil || from == to || walk.visited[[2]*checker.Type{from, to}] {
		return true
	}
	if from.Flags()&checker.TypeFlagsUnion != 0 {
		walk.visited[[2]*checker.Type{from, to}] = true
		for _, member := range l.definedMembers(from) {
			if member.Flags()&checker.TypeFlagsNull == 0 && !l.sameRecordStorage(member, to, walk) {
				return false
			}
		}
		return true
	}
	if to.Flags()&checker.TypeFlagsUnion != 0 {
		walk.visited[[2]*checker.Type{from, to}] = true
		members := l.definedMembers(to)
		// A value enters a union that names its own type as that member, whatever other members it's
		// assignable to: CompilerOptions in CompilerOptions | PrinterOptions stays a record, and a read on
		// the union dispatches on the value it holds.
		for _, member := range members {
			if member == from {
				return true
			}
		}
		for _, member := range members {
			if l.checker.IsTypeAssignableTo(from, member) && !l.sameRecordStorage(from, member, walk) {
				return false
			}
		}
		return true
	}
	target := to
	from, to = l.present(from), l.present(to)
	if to == nil && l.fixedObjectIntersection(target) {
		// An intersection of plain object shapes is held as a fixed object (objectIntersection), so it's
		// judged as one: present has no single object type for it, and answering same storage let a
		// finite Partial<Record<K, V>> declaring each of its fields cross, and a fixed-object read of it
		// panic natively as a compiler bug where Node runs (#mft3zc6). Any other intersection keeps the
		// nil answer, as does every other type present has none for.
		to = target
	}
	// A Pick of a record is stored as the whole record, so it's compared as that record.
	if base := l.pickedRecordBase(from); base != nil {
		from = base
	}
	if base := l.pickedRecordBase(to); base != nil {
		to = base
	}
	if from == nil || to == nil || from == to || walk.visited[[2]*checker.Type{from, to}] {
		return true
	}
	walk.visited[[2]*checker.Type{from, to}] = true
	// A generic that instantiates itself deeper at each level (Own<T> { next: Own<Own<T>> } seen as
	// Deep<T> { next: Deep<Deep<T>> }) is a fresh pair at every step, so visited never repeats and the
	// walk ran without end (#fzx0j5z). typeNesting (type_nesting.go) is tsc's isDeeplyNestedType: past
	// its bound on either side, the walk stops and answers false with bounded set, failing closed as not
	// yet at the caller. The pair is entered here, before any arm walks inside it, since a record's
	// element and a signature's parameters recur through their own arms without reaching the members.
	fromOK, leaveFrom := walk.from.enter(from)
	defer leaveFrom()
	toOK, leaveTo := walk.to.enter(to)
	defer leaveTo()
	if !fromOK || !toOK {
		walk.bounded = true
		return false
	}
	inside, viewed := l.recordElement(from), l.recordElement(to)
	if inside != nil || viewed != nil {
		return inside != nil && viewed != nil && l.sameRecordNamedContracts(from, to, inside, viewed, walk) && l.sameRecordStorage(inside, viewed, walk)
	}
	a, b := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall), l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	if len(a) > 0 && len(b) > 0 {
		ap, bp := a[0].Parameters(), b[0].Parameters()
		parameters := l.parameterStorage(func() bool {
			for i := 0; i < len(ap) && i < len(bp); i++ {
				if !l.sameRecordStorage(l.checker.GetTypeOfSymbol(ap[i]), l.checker.GetTypeOfSymbol(bp[i]), walk) {
					return false
				}
			}
			return true
		})
		return parameters && l.sameRecordStorage(l.checker.GetReturnTypeOfSignature(a[0]), l.checker.GetReturnTypeOfSignature(b[0]), walk)
	}
	if l.isArrayType(from) || checker.IsTupleType(from) || l.isLibraryType(from, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
		a, b := l.typeArguments(from), l.typeArguments(to)
		for i := 0; i < len(a) && i < len(b); i++ {
			if !l.sameRecordStorage(a[i], b[i], walk) {
				return false
			}
		}
		return true
	}
	// A promise holds the value it settles to and nothing else of the program's, so lib's Promise and
	// PromiseLike store alike when their one type argument does. Walking their then members instead
	// meets a fresh instantiation at every level, past typeNesting's bound: Promise<number> seen as
	// PromiseLike<number> stopped at the bound on oracle's async adoption and callback pins (J-vsqgd66).
	if l.isLibraryType(from, "Promise", "PromiseLike") && l.isLibraryType(to, "Promise", "PromiseLike") {
		a, b := l.typeArguments(from), l.typeArguments(to)
		return len(a) == 1 && len(b) == 1 && l.sameRecordStorage(a[0], b[0], walk)
	}
	// Two instances of one generic type store alike when their arguments do. Walking their members
	// instead expands Promise's then and catch into fresh instantiations at every level, which no
	// visited pair ever repeats: 1.3 million calls, 294,120 deep, for Promise<() => void>. The
	// arguments decide outright, failing closed when they differ, since a pair this walk marked
	// visited would read as alike to any member walk after it.
	if from.ObjectFlags()&to.ObjectFlags()&checker.ObjectFlagsReference != 0 && from.Target() != nil && from.Target() == to.Target() {
		a, b := l.checker.GetTypeArguments(from), l.checker.GetTypeArguments(to)
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if !l.sameRecordStorage(a[i], b[i], walk) {
				return false
			}
		}
		return true
	}
	for _, property := range l.checker.GetPropertiesOfType(to) {
		if own := l.checker.GetPropertyOfType(from, property.Name); own != nil {
			if !l.sameRecordStorage(l.checker.GetTypeOfSymbol(own), l.checker.GetTypeOfSymbol(property), walk) {
				return false
			}
		}
	}
	return true
}

// fixedObjectIntersection reports whether proven is an intersection held as a fixed object: every
// member a plain object shape and every field one it holds, as objectIntersection admits it as
// ir.Object. A Weak's narrowing (Target & WeakBrand) and a primitive brand are held as their target or
// base instead, so neither is one. It's stricter than view_lazy.go's plainObjectIntersection, which
// asks only that every member be a plain object type and the whole not a Weak's present half: that
// admits an intersection whose field no object holds (a Weak or a function field), which has no
// representation and so is no fixed object to compare storage with.
func (l *lowering) fixedObjectIntersection(proven *checker.Type) bool {
	if proven == nil || proven.Flags()&checker.TypeFlagsIntersection == 0 || l.weakTarget(proven) != nil || l.phantomBase(proven) != nil {
		return false
	}
	held, known, _ := l.objectIntersection(proven)
	return known && held == ir.Object
}

// Catch literal reads before method/builtin dispatch can bypass ordinary value lowering.
func (l *lowering) recordLiteralRead(node *ast.Node) error {
	// recordReadTarget skips parentheses, so a parenthesized read answers too. The rule walk
	// visits the access inside the parentheses on its own, and judges it there (#grc95d6).
	if node.Kind != ast.KindPropertyAccessExpression && node.Kind != ast.KindElementAccessExpression {
		return nil
	}
	if !l.recordReadTarget(node) || l.recordReadDiscarded(node) || l.recordReadOwnGuarded(node) {
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
			b := parent.AsBinaryExpression()
			if b.Left == at && b.OperatorToken.Kind == ast.KindEqualsToken {
				return nil
			}
		}
	}
	if parent := at.Parent; parent != nil && parent.Kind == ast.KindBinaryExpression {
		// The scalar comparison owns the read and proves its inherited value unobservable.
		b := parent.AsBinaryExpression()
		other := b.Right
		if b.Right == at {
			other = b.Left
		}
		t := l.checker.GetTypeAtLocation(other)
		if (b.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken || b.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken) && t.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike) != 0 && !l.includesUndefined(t) {
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
		return &Refused{Where: l.program.Where(node), What: "record key " + name + " names an Object.prototype member", Fix: "use own keys"}
	}
	return nil
}

func init() { refusalRules.register(1100, "recordLiteralRead", (*lowering).recordLiteralRead) }

// optionalChainOperand reports whether node, through any parentheses, is what a ?. reads: the object of
// r["k"]?.name or r["k"]?.[i], or the callee of r["k"]?.(...).
func optionalChainOperand(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		access := parent.AsPropertyAccessExpression()
		return access.Expression == node && access.QuestionDotToken != nil
	case ast.KindElementAccessExpression:
		access := parent.AsElementAccessExpression()
		return access.Expression == node && access.QuestionDotToken != nil
	case ast.KindCallExpression:
		call := parent.AsCallExpression()
		return call.Expression == node && call.QuestionDotToken != nil
	}
	return false
}
