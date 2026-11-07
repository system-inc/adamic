package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Fold only effect-free strings; a runtime computed key is not a shape proof.
func constantStringKey(node *ast.Node) (string, bool) {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindStringLiteral {
		return node.Text(), true
	}
	if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindPlusToken {
		left, a := constantStringKey(node.AsBinaryExpression().Left)
		right, b := constantStringKey(node.AsBinaryExpression().Right)
		return left + right, a && b
	}
	return "", false
}

func (lowering *lowering) enumerationInModule(module *ast.SourceFile) bool {
	return lowering.enumerationMethodsInModule(module, false)
}

func (lowering *lowering) enumerationMethodsInModule(module *ast.SourceFile, valuesOnly bool) bool {
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(node.AsCallExpression().Expression)
			if callee.Kind == ast.KindPropertyAccessExpression &&
				lowering.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
				name := callee.Name().Text()
				found = found || name == "entries" || name == "values" || !valuesOnly && name == "keys"
			}
		}
		return node.ForEachChild(visit)
	}
	module.AsNode().ForEachChild(visit)
	return found
}

// The declaration alone does not promise a runtime dictionary. Ordinary closed
// literal origins stay fixed; indexed writes below explicitly select record storage.
func (lowering *lowering) enumerationIndexSignature(node *ast.Node) bool {
	if !lowering.enumerationInModule(ast.GetSourceFileOfNode(node).AsSourceFile()) {
		return false
	}
	signature := node.AsIndexSignatureDeclaration()
	return len(signature.Parameters.Nodes) == 1 &&
		signature.Parameters.Nodes[0].AsParameterDeclaration().Type != nil &&
		signature.Parameters.Nodes[0].AsParameterDeclaration().Type.Kind == ast.KindStringKeyword
}

// Reassigned origins are never guessed. Const aliases can still mutate their
// object: those writes select record storage and invalidate unchecked enumeration.
func (lowering *lowering) entriesLiteralOrigin(node *ast.Node, depth int) *ast.Node {
	if node == nil || depth > 32 {
		return nil
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindObjectLiteralExpression {
		return node
	}
	if node.Kind == ast.KindAsExpression && ast.IsConstTypeReference(node.AsAsExpression().Type) {
		return lowering.entriesLiteralOrigin(node.AsAsExpression().Expression, depth+1)
	}
	if !ast.IsIdentifier(node) {
		return nil
	}
	symbol := lowering.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return nil
	}
	return lowering.entriesLiteralOrigin(declaration.AsVariableDeclaration().Initializer, depth+1)
}

// entriesRecordLiteral says whether literal is a record: a const's initializer that the program writes through by a
// string key. Its answer is the program's and the checker's alone, so a lowering keeps it per literal, and every later
// ask reads that: object.go asks it for every object literal it lowers, and each walk read the whole program
// (#pv935yw). The walk's moduleOrder of the program's first file sets cyclicModules and provenModuleReads, the same
// state every ask set before, so an ask that reads the memo leaves standing what it would have set again.
func (lowering *lowering) entriesRecordLiteral(literal *ast.Node) bool {
	if answer, ok := lowering.entriesRecords[literal]; ok {
		return answer
	}
	answer := lowering.entriesRecordLiteralWalk(literal)
	if lowering.entriesRecords == nil {
		lowering.entriesRecords = map[*ast.Node]bool{}
	}
	lowering.entriesRecords[literal] = answer
	return answer
}

func (lowering *lowering) entriesRecordLiteralWalk(literal *ast.Node) bool {
	parent := literal.Parent
	for parent != nil && (parent.Kind == ast.KindAsExpression || parent.Kind == ast.KindParenthesizedExpression) {
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindVariableDeclaration || parent.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	contextual := lowering.slotType(literal)
	indexed := contextual != nil && len(lowering.checker.GetIndexInfosOfType(contextual)) != 0
	modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
	if moduleOrderError != nil {
		return false
	}
	written := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindBinaryExpression &&
			ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
			target := ast.SkipParentheses(node.AsBinaryExpression().Left)
			var receiver *ast.Node
			if target.Kind == ast.KindElementAccessExpression {
				receiver = target.AsElementAccessExpression().Expression
			}
			if target.Kind == ast.KindPropertyAccessExpression {
				receiver = target.AsPropertyAccessExpression().Expression
			}
			if receiver != nil &&
				lowering.entriesLiteralOrigin(receiver, 0) == literal &&
				(indexed || lowering.entriesStringIndexed(receiver)) {
				written = true
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return written
}

func (lowering *lowering) entriesRecordWrite(target, valueNode *ast.Node) ([]ir.Statement, bool, error) {
	var receiver, keyNode *ast.Node
	key, known := "", false
	if target.Kind == ast.KindElementAccessExpression {
		access := target.AsElementAccessExpression()
		receiver, keyNode = access.Expression, access.ArgumentExpression
		key, known = constantStringKey(keyNode)
	} else {
		receiver = target.AsPropertyAccessExpression().Expression
		key, known = target.Name().Text(), true
	}
	origin := lowering.entriesLiteralOrigin(receiver, 0)
	if origin == nil || !lowering.entriesRecordLiteral(origin) {
		return nil, false, nil
	}
	if !known ||
		key == "__proto__" ||
		strings.ContainsRune(key, 0) ||
		strings.HasPrefix(key, "#") ||
		key == iteratorSlot {
		return nil, true, lowering.notYet(target, "an indexed record write without a plain constant data key")
	}
	object, expressionError := lowering.expression(receiver)
	if expressionError != nil {
		return nil, true, expressionError
	}
	value, expressionError := lowering.expression(valueNode)
	if expressionError != nil {
		return nil, true, expressionError
	}
	if value.Type() != ir.Number &&
		value.Type() != ir.String &&
		value.Type() != ir.Boolean &&
		value.Type() != ir.Union {
		return nil, true, lowering.notYet(target, "a record write outside scalar storage")
	}
	return []ir.Statement{ir.SetProperty{
		Object: object,
		Name:   key,
		Value:  fit(value, ir.Union),
		Define: true,
		Record: true,
		Site:   lowering.writeSite(receiver),
	}}, true, nil
}

func (lowering *lowering) enumerationDescriptorRefusal(node *ast.Node) error {
	if node.Kind != ast.KindGetAccessor &&
		node.Kind != ast.KindSetAccessor &&
		node.Kind != ast.KindComputedPropertyName {
		return nil
	}
	if !lowering.enumerationMethodsInModule(ast.GetSourceFileOfNode(node).AsSourceFile(), true) {
		return nil
	}
	if node.Kind == ast.KindComputedPropertyName {
		if lowering.checker.GetTypeAtLocation(
			node.AsComputedPropertyName().Expression,
		).Flags()&checker.TypeFlagsESSymbolLike == 0 {
			return nil
		}
		return &Refused{
			Where: lowering.program.Where(node),
			What:  "Object enumeration with symbol keys",
			Fix:   "use own string-keyed data properties",
		}
	}
	return &Refused{
		Where: lowering.program.Where(node),
		What:  "Object enumeration with a getter or accessor",
		Fix:   "use own data properties",
	}
}

func init() {
	refusalRules.register(700, "enumerationDescriptorRefusal", (*lowering).enumerationDescriptorRefusal)
}

// entriesProgramHasRecords says whether any object literal in the program is a record. It's one answer per program,
// kept on the lowering after its first ask (#pv935yw).
func (lowering *lowering) entriesProgramHasRecords() bool {
	if !lowering.entriesProgramRecordsKnown {
		lowering.entriesProgramRecords, lowering.entriesProgramRecordsKnown = lowering.entriesProgramRecordsWalk(), true
	}
	return lowering.entriesProgramRecords
}

func (lowering *lowering) entriesProgramRecordsWalk() bool {
	modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
	if moduleOrderError != nil {
		return true
	}
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindObjectLiteralExpression && lowering.entriesRecordLiteral(node) {
			found = true
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return found
}

func (lowering *lowering) entriesStringIndexed(receiver *ast.Node) bool {
	for _, info := range lowering.checker.GetIndexInfosOfType(lowering.checker.GetTypeAtLocation(receiver)) {
		if info.KeyType().Flags()&checker.TypeFlagsStringLike != 0 {
			return true
		}
	}
	return false
}

// A parameter or replaced binding cannot establish which allocation must grow.
// Keep those writes outside this first record-origin implementation.
func (lowering *lowering) enumerationIndexedWriteRefusal(node *ast.Node) error {
	if node.Kind != ast.KindBinaryExpression ||
		!ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
		return nil
	}
	target := ast.SkipParentheses(node.AsBinaryExpression().Left)
	var receiver *ast.Node
	if target.Kind == ast.KindPropertyAccessExpression {
		receiver = target.AsPropertyAccessExpression().Expression
	}
	if target.Kind == ast.KindElementAccessExpression {
		receiver = target.AsElementAccessExpression().Expression
	}
	if receiver != nil && lowering.recordElement(lowering.checker.GetTypeAtLocation(receiver)) != nil {
		return nil // Dictionary writes carry their own named-slot and cycle proofs.
	}
	if receiver == nil ||
		!lowering.entriesStringIndexed(receiver) ||
		lowering.entriesLiteralOrigin(receiver, 0) != nil {
		return nil
	}
	return lowering.notYet(node, "an indexed mutation without a closed const literal origin")
}

func init() {
	refusalRules.register(600, "enumerationIndexedWriteRefusal", (*lowering).enumerationIndexedWriteRefusal)
}

// A record is a table behind one slot, never a fixed shape, so a plain union or maybe field read of
// one would find no field (the runtime stops it: object.c's adamic_refuse_record_fixed_read). In a
// program with records, such a read lowers only when no record can be its receiver: not one of the
// types records are made as (rule b), not a type one of them is assignable to (rule c, which covers
// a record passed through a parameter, a generic or a union member, and the reverse, which covers a
// record narrowed to a subtype or an intersection), and not a field a checked view registered, since
// a record reaches a fixed type through a view (rule a, checkRecordGuardedReads). Today (a) and (b)
// are implied by (c) and have no witness of their own; they are kept so a new entrance fails closed.
type recordGuardedRead struct {
	node           *ast.Node
	receiverTypeID int
	name           string
}

// entriesRecordStorageTypes are the types the program's record literals are made as: the literal's own,
// its contextual type and its variable's declared type.
func (lowering *lowering) entriesRecordStorageTypes() []*checker.Type {
	if lowering.entriesRecordTypesKnown {
		return lowering.entriesRecordTypes
	}
	lowering.entriesRecordTypesKnown = true
	modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
	if moduleOrderError != nil {
		return nil
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindObjectLiteralExpression && lowering.entriesRecordLiteral(node) {
			lowering.entriesRecordTypes = append(lowering.entriesRecordTypes, lowering.checker.GetTypeAtLocation(node))
			if contextual := lowering.slotType(node); contextual != nil {
				lowering.entriesRecordTypes = append(lowering.entriesRecordTypes, contextual)
			}
			parent := node.Parent
			for parent != nil &&
				(parent.Kind == ast.KindAsExpression || parent.Kind == ast.KindParenthesizedExpression) {
				parent = parent.Parent
			}
			if parent != nil && parent.Kind == ast.KindVariableDeclaration {
				lowering.entriesRecordTypes = append(
					lowering.entriesRecordTypes,
					lowering.checker.GetTypeAtLocation(parent.Name()),
				)
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return lowering.entriesRecordTypes
}

// recordMayReach says whether a record could be the receiver of a field read whose receiver has
// this type. Unknown fails closed: an unconstrained type parameter, any and unknown may be a record.
func (lowering *lowering) recordMayReach(receiver *checker.Type) bool {
	return lowering.typesMayReach(receiver, lowering.entriesRecordStorageTypes(), true, 0)
}

// typesMayReach says whether a value made as one of sources could be the receiver of a field read
// whose receiver has this type, records saying the sources are records' (an enumeration's own storage
// type is one too). Unknown fails closed, as recordMayReach says.
func (lowering *lowering) typesMayReach(receiver *checker.Type, sources []*checker.Type, records bool, depth int) bool {
	if receiver == nil || depth > 8 {
		return true
	}
	if receiver.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return true
	}
	if receiver.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range receiver.Types() {
			if member.Flags()&checker.TypeFlagsNullable == 0 &&
				lowering.typesMayReach(member, sources, records, depth+1) {
				return true
			}
		}
		return false
	}
	if receiver.Flags()&checker.TypeFlagsIntersection != 0 {
		for _, member := range receiver.Types() {
			if lowering.typesMayReach(member, sources, records, depth+1) {
				return true
			}
		}
	}
	if receiver.Flags()&checker.TypeFlagsInstantiable != 0 {
		if concrete := lowering.concrete(receiver); concrete != nil && concrete != receiver {
			return lowering.typesMayReach(concrete, sources, records, depth+1)
		}
		return lowering.typesMayReach(lowering.checker.GetBaseConstraintOfType(receiver), sources, records, depth+1)
	}
	if records && lowering.enumerationRecordStorageType(receiver) {
		return true
	}
	// Both ways: a record seen as a wider type, and a record narrowed to a narrower one (a predicate,
	// a discriminant), arrive alike.
	for _, source := range sources {
		if source == receiver ||
			lowering.checker.IsTypeAssignableTo(source, receiver) ||
			lowering.checker.IsTypeAssignableTo(receiver, source) {
			return true
		}
	}
	return false
}

// A bare-name CheckedFields entry stands for a record's field (readiness's markRecord) or a field a
// class merged with an interface promises without initializing (mergedFieldRead), and keeps every read
// and write of that name checked. plainReceiver records, for the receiver type a read or a write is
// checked under, that no value of it can be either (#9jnjwrw): no record and no merged class is
// assignable to it or it to them, by the same fail-closed walk as recordMayReach. The proof is the
// type's alone, never a claim that a value is a class instance: TypeScript is structural, so an object
// literal can stand where a class is declared, and a class type is plain only when no record and no
// merged class can be one. ir.Program.CheckedField then leaves such a receiver's reads unchecked by a
// bare name; a view member's own key still checks it.
func (lowering *lowering) plainReceiver(id int) {
	if lowering.plainAsked[id] {
		return
	}
	if lowering.plainAsked == nil {
		lowering.plainAsked = map[int]bool{}
	}
	lowering.plainAsked[id] = true
	if _, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0]); moduleOrderError != nil {
		return // Unknown fails closed: the record and merged class types are the modules'.
	}
	receiver := lowering.viewTypes[id]
	if receiver == nil ||
		lowering.recordMayReach(receiver) ||
		lowering.typesMayReach(receiver, lowering.mergedClassInstanceTypes(), false, 0) {
		return
	}
	if lowering.result.PlainReceivers == nil {
		lowering.result.PlainReceivers = ir.SortedMap[int, bool]{}
	}
	lowering.result.PlainReceivers[id] = true
}

// mergedClassInstanceTypes are the instance types of the program's classes merged with an interface.
func (lowering *lowering) mergedClassInstanceTypes() []*checker.Type {
	if lowering.mergedClassTypesKnown {
		return lowering.mergedClassTypes
	}
	lowering.mergedClassTypesKnown = true
	modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
	if moduleOrderError != nil {
		return nil // plainReceiver proves nothing without the modules.
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindClassDeclaration && lowering.mergedClass(node) == node {
			lowering.mergedClassTypes = append(
				lowering.mergedClassTypes,
				lowering.checker.GetTypeAtLocation(node.Name()),
			)
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return lowering.mergedClassTypes
}

// checkRecordGuardedReads runs once lowering has registered every checked view: a read admitted
// beside records stays NotYet when a view registered its field, since that's how a record reaches
// a fixed type the checker can't otherwise see.
func (lowering *lowering) checkRecordGuardedReads() error {
	for _, read := range lowering.recordGuardedReads {
		if lowering.result.CheckedField(read.receiverTypeID, read.name) {
			return lowering.notYet(read.node, "a union or optional field read in a program with record storage")
		}
	}
	return nil
}

// weakFieldRead is a union or optional field read, in a .ts file, whose receiver a record may be. tsc
// lets a record be assigned to a weak type (every field optional) whatever its value type, so a
// record's table can arrive where a fixed object's slot is read, and a slot read of it would be wrong.
// The receiver's storage decides as the read runs, in a helper that takes the receiver once
// (weak_field_read). A fixed object reads its slot as any field read does (fixedFieldRead). A record
// reads its key, an absent one undefined as Node reads it, and the value is held to the field's type
// by the check every record read takes (recordReadDeclared), so a string under a field of 1 | 2 stops
// there, named, and is never read as a number. Any other heap (a class instance, an Error, an array, a
// Map, a function) stops loudly by name and never reaches the slot read. A nullish receiver under ?.
// reads undefined, as the slot read does. Where the check can't hold the field's type (an object or
// an array member) or makes none (a field held as a Union), the read stays NotYet. The storage test and the value check
// are each a checked site in the report --explain-checks prints.
func (lowering *lowering) weakFieldRead(
	node *ast.Node,
	object ir.Expression,
	name string,
	of ir.Type,
	optional bool,
) (ir.Expression, error) {
	b := lowering.libraryArrayBuilder([]ir.Expression{object})
	// The helper's index is reserved now, so a helper the record's check makes takes the next one, as
	// expression.go's prefix_update reserves its own.
	lowering.result.Functions = append(lowering.result.Functions, ir.Function{Name: "weak_field_read", Returns: of})
	receiver := b.read(b.parameters[0])
	slot, fixedFieldReadError := lowering.fixedFieldRead(node, receiver, name, of, optional)
	if fixedFieldReadError != nil {
		return nil, fixedFieldReadError
	}
	// The fixed read runs only where the storage test found no record, so a record that writes a field
	// of this name mustn't make it a checked view read (readiness keeps view data for a record's field
	// names, CheckedField's bare name), which a fixed object's optional field can't be emitted as.
	switch read := slot.(type) {
	case ir.Property:
		read.NotRecord = true
		slot = read
	case ir.Unwrap:
		property, isProperty := read.Value.(ir.Property)
		if !isProperty {
			return nil, lowering.notYet(node, "a union or optional field read in a program with record storage")
		}
		property.NotRecord = true
		read.Value = property
		slot = read
	default:
		return nil, lowering.notYet(node, "a union or optional field read in a program with record storage")
	}
	stored := ir.RecordCall{
		Method:    "get",
		Arguments: []ir.Expression{receiver, ir.StringConstant{Index: lowering.constant(name)}},
		Element:   ir.Union,
		Returns:   ir.Union,
	}
	record, recordReadDeclaredError := lowering.recordReadDeclared(
		node,
		stored,
		lowering.checker.GetTypeAtLocation(node),
	)
	if recordReadDeclaredError != nil {
		return nil, recordReadDeclaredError
	}
	// recordReadDeclared gives the read back unchecked where it builds no check: a field held as a Union,
	// whose tag it trusts. Through a weak type a record's values may be of any type, so a number would be
	// read as object | string; that read waits for a value check of its own, as tsc's two
	// packageJsonContent.exports reads (exports?: object) do.
	if _, unchecked := record.(ir.RecordCall); unchecked || record.Type() != slot.Type() {
		return nil, lowering.notYet(node, "a union or optional field read in a program with record storage")
	}
	// What the receiver turned out to be, named, for the stop.
	kind := b.local("kind", ir.Number)
	found := ir.Expression(ir.StringConstant{Index: lowering.constant("another value")})
	for _, other := range []struct {
		kind int
		name string
	}{
		{ir.UnknownUndefined, "undefined"},
		{ir.UnknownNull, "null"},
		{ir.UnknownNumber, "a number"},
		{ir.UnknownBoolean, "a boolean"},
		{ir.UnknownString, "a string"},
		{ir.UnknownError, "an Error"},
		{ir.UnknownInstance, "a class instance"},
		{ir.UnknownStaticClass, "a class"},
		{ir.UnknownArray, "an array"},
		{ir.UnknownFunction, "a function"},
		{ir.UnknownCollection, "a Map or a Set"},
		{ir.UnknownTypedArray, "a typed array"},
	} {
		found = ir.Conditional{Condition: ir.Binary{
			Operator: ir.Equal,
			Left:     b.read(kind),
			Right:    ir.NumberConstant{Value: float64(other.kind)},
		}, WhenTrue: ir.StringConstant{Index: lowering.constant(other.name)}, WhenNot: found, Of: ir.String}
	}
	seen := lowering.checker.TypeToString(
		lowering.checker.GetNonNullableType(
			lowering.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression),
		),
	)
	message := ir.Concat{Parts: []ir.Expression{
		ir.StringConstant{Index: lowering.constant("field read failed: " +
			recordReadExpression(node) +
			" reads " +
			seen +
			" storage; expected a fixed object or a record, found ")},
		found,
	}}
	storage := []ir.Statement{
		ir.Declare{Local: kind, Value: ir.UnknownKind{Value: fit(receiver, ir.Union)}},
		ir.If{Condition: ir.Binary{
			Operator: ir.NotEqual,
			Left:     b.read(kind),
			Right:    ir.NumberConstant{Value: ir.UnknownObject},
		}, Then: []ir.Statement{ir.Panic{Message: message}}},
		ir.If{Condition: ir.RecordStorage{Value: receiver}, Then: []ir.Statement{ir.Return{Value: record}}},
	}
	if optional {
		// ?. asks nothing of an absent receiver: the slot read gives undefined for it.
		storage = []ir.Statement{
			ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: receiver}}, Then: storage},
		}
	}
	b.body = append(b.body, storage...)
	b.body = append(b.body, ir.Return{Value: slot})
	lowering.result.Functions[b.function] = ir.Function{
		Name:       "weak_field_read",
		Parameters: b.parameters,
		Returns:    slot.Type(),
		Body:       b.body,
	}
	lowering.result.PlaceholderChecks = append(lowering.result.PlaceholderChecks, ir.PlaceholderCheck{
		Origin: "record storage",
		Use:    "weak field read storage",
		Path:   sourceExpression(node),
		Where:  lowering.program.Where(node),
		Status: "checked",
	})
	lowering.result.PlaceholderChecks = append(lowering.result.PlaceholderChecks, ir.PlaceholderCheck{
		Origin: "record storage",
		Use:    "weak field read record value",
		Path:   sourceExpression(node),
		Where:  lowering.program.Where(node),
		Status: "checked",
	})
	return ir.Call{Function: b.function, Arguments: b.arguments, Returns: slot.Type()}, nil
}
