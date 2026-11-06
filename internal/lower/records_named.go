package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/ir"
)

// Dictionary entries have one storage kind. Heterogeneous values use the existing
// counted union boxes. Two-word optional booleans retain their existing boundary.
// A nullable value is kept as representation boxes it: a Union slot keeps null
// (adamic_null) apart from undefined (a null pointer), and the entry keeps presence
// apart from both. Reads off it unbox through recordReadType's checked tag.
func (l *lowering) recordStorageType(t *checker.Type) (ir.Type, bool) {
	of, known := l.kept(t)
	return of, known && of != 0 && of != ir.Weak && of != ir.MaybeBoolean
}

// The checker resolves literal keys to their named declaration and dynamic keys
// to the index signature. Never interpret a boxed index value as an unboxed field.
func recordTagObservation(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return comparedWithUndefined(node) || parent != nil && parent.Kind == ast.KindTypeOfExpression
}

func (l *lowering) recordReadType(node *ast.Node, read ir.Expression, here *checker.Type) (ir.Expression, error) {
	if recordTagObservation(node) {
		return read, nil // These operations observe the runtime tag rather than unbox it.
	}
	return l.recordReadDeclared(node, read, here)
}

// recordReadDeclared is recordReadType's check of a value read out of a slot boxed as a union, held
// to here, the type the read is declared or narrowed to, whatever observes it. A weak field read of
// a record (weakFieldRead) is held by it too: its result is the field's representation, which a
// typeof or an === undefined then compares, so there's no tag left to observe.
func (l *lowering) recordReadDeclared(node *ast.Node, read ir.Expression, here *checker.Type) (ir.Expression, error) {
	to, known := l.representation(here)
	if !known || read.Type() != ir.Union || to == ir.Union {
		return read, nil
	}
	name := ""
	switch to.Present() {
	case ir.Number:
		name = "number"
	case ir.Boolean:
		name = "boolean"
	case ir.String:
		name = "string"
	case ir.Closure:
		name = "function"
	case ir.Object, ir.Record:
		return l.recordReadHeapMember(node, read, here, to)
	case ir.Array:
		// Arrays of different element types share one heap kind, and walking the elements proves nothing
		// once another alias pushes into it, so an array read waits on an element-checked view, ruled on
		// #3xq2xzc and built in Baobab (#y852c3c): tsc's typeRoots, rootDirs, moduleSuffixes and types.
		return nil, l.notYet(node, "a dictionary member narrowed to an array without an element-checked view")
	default:
		return nil, l.notYet(node, "a dictionary member narrowed to an object kind without a checked runtime tag")
	}
	b := l.libraryArrayBuilder([]ir.Expression{read})
	value := b.read(b.parameters[0])
	matches := ir.Expression(ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: value}, Right: ir.StringConstant{Index: l.constant(name)}})
	if l.includesUndefined(here) {
		matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsUndefined{Value: value}}
	}
	// A write through the index signature, by a computed key or through a wider alias, can leave any
	// member of the slot's union under a declared key. The stop names the read and what it found.
	expression := recordReadExpression(node)
	declared := l.checker.TypeToString(l.checker.GetNonNullableType(here))
	found := ir.Conditional{Condition: ir.IsNull{Value: value}, WhenTrue: ir.StringConstant{Index: l.constant("null")}, WhenNot: ir.TypeOf{Value: value}, Of: ir.String}
	message := ir.Concat{Parts: []ir.Expression{
		ir.StringConstant{Index: l.constant("field read failed: " + expression + " is not a " + declared + "; expected " + declared + ", found ")},
		found,
		ir.StringConstant{Index: l.constant(", which a computed-key or index-signature write may have stored")},
	}}
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: message}}})
	// typeof alone lets a 3 under 1 | 2, or Kind.C under Kind.A | Kind.B. Where the declared type is
	// its literals alone (viewLiterals, the list a checked view holds a field to), a present value must
	// be one of them; beside a member of its own representation (1 | number) the typeof is enough, and
	// viewLiterals gives none. Lower's ruling on #1w8gvx9. A whole numeric enum is no closed list:
	// numeric enums are open (docs/rulings.md, resolved October 19), so any number is one of its values,
	// and only a strict subset of its members (ModuleKind.ESNext | ModuleKind.CommonJS) is held.
	if allowed := l.viewLiterals(l.checker.GetNonNullableType(here)); len(allowed) != 0 && !l.openNumericEnumType(here) {
		present := b.local("present", to.Present())
		var member ir.Expression
		for _, literal := range allowed {
			equal := ir.Binary{Operator: ir.Equal, Left: b.read(present), Right: literal}
			if member == nil {
				member = equal
			} else {
				member = ir.Binary{Operator: ir.Or, Left: member, Right: equal}
			}
		}
		shown := b.read(present)
		switch to.Present() {
		case ir.Number:
			shown = ir.NumberToString{Value: shown}
		case ir.Boolean:
			shown = ir.BooleanToString{Value: shown}
		}
		missed := ir.Concat{Parts: []ir.Expression{
			ir.StringConstant{Index: l.constant("field read failed: " + expression + " is not a " + declared + "; expected " + declared + ", found " + name + " ")},
			shown,
			ir.StringConstant{Index: l.constant(", which a computed-key or index-signature write may have stored")},
		}}
		membership := []ir.Statement{
			ir.Declare{Local: present, Value: ir.Narrow{Value: value, To: to.Present()}},
			ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: member}, Then: []ir.Statement{ir.Panic{Message: missed}}},
		}
		if l.includesUndefined(here) {
			membership = []ir.Statement{ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: value}}, Then: membership}}
		}
		b.body = append(b.body, membership...)
	}
	return b.finish("record_declared_member", ir.Narrow{Value: value, To: to}), nil
}

// recordReadExpression is the read's source text, as a stop names it.
func recordReadExpression(node *ast.Node) string {
	file := ast.GetSourceFileOfNode(node)
	return file.Text()[scanner.GetTokenPosOfNode(node, file, false):node.End()]
}

// recordReadHeapMember is recordReadType's read of a fixed object or a dictionary out of a slot boxed as
// a union. typeof says object for both, and for an array and a Map, so no tag names the member. But every
// value in the slot is one of the slot union's members (the checker's assignability at each write, with
// recordStorageView refusing a fixed object seen as a record and back), so a heap test proves the declared
// member when no other member of the slot's union has its runtime class: a fixed object is a non-null
// object heap that isn't a record, a dictionary one that is (ir.Narrow's checked arms). tsc's
// CompilerOptions reads configFile?: TsConfigSourceFile and paths?: MapLike<string[]> this way, beside
// arrays, scalars and null. Two members of one class, an array's four, a member whose runtime class
// isn't known, or a declared type that admits null (which the checked arm stops on) keep the NotYet.
func (l *lowering) recordReadHeapMember(node *ast.Node, read ir.Expression, here *checker.Type, to ir.Type) (ir.Expression, error) {
	declared := l.concrete(l.checker.GetNonNullableType(here))
	if l.includesNull(here) || !l.recordSlotMemberUnique(recordReceiver(node), declared, to) {
		return nil, l.notYet(node, "a dictionary member narrowed to an object kind without a checked runtime tag")
	}
	b := l.libraryArrayBuilder([]ir.Expression{read})
	value := b.read(b.parameters[0])
	if l.includesUndefined(here) {
		b.body = append(b.body, ir.If{Condition: ir.IsUndefined{Value: value}, Then: []ir.Statement{ir.Return{Value: ir.Undefined{Of: to}}}})
	}
	named := l.checker.TypeToString(declared)
	// A write through the index signature, by a computed key or through a wider alias, can leave any
	// member of the slot's union under a declared key. The stop names the read and what it found.
	message := "field read failed: " + recordReadExpression(node) + " is not a " + named + "; expected " + named + ", found another member of the index signature's union, which a computed-key or index-signature write may have stored"
	return b.finish("record_declared_member", ir.Narrow{Value: value, To: to, Checked: true, Message: message}), nil
}

// recordSlotMemberUnique is whether declared is the only member of receiver's index signature value
// union whose runtime class is to's, nullish members aside. Every member's representation must be one
// whose heap kind is its own: a Union (unknown, object, {}) could be anything, a Closure from the
// library's Function could be a class constructor, an object kind with an index signature may be
// enumeration storage, and the rest (a Promise, a typed array, a Weak) aren't vouched for here.
func (l *lowering) recordSlotMemberUnique(receiver *ast.Node, declared *checker.Type, to ir.Type) bool {
	infos := l.checker.GetIndexInfosOfType(l.concrete(l.checker.GetTypeAtLocation(receiver)))
	if len(infos) != 1 || infos[0].KeyType().Flags() != checker.TypeFlagsString {
		return false
	}
	if to == ir.Object && len(l.checker.GetIndexInfosOfType(declared)) != 0 {
		return false
	}
	var unique *checker.Type
	for _, member := range l.definedMembers(l.concrete(infos[0].ValueType())) {
		if member.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
			continue
		}
		of, known := l.representation(member)
		if !known || !recordSlotHeapClasses[of] {
			return false
		}
		if of == ir.Object && len(l.checker.GetIndexInfosOfType(member)) != 0 || of == ir.Closure && len(l.checker.GetSignaturesOfType(member, checker.SignatureKindCall)) == 0 {
			return false
		}
		if of != to {
			continue
		}
		if unique != nil {
			return false
		}
		unique = member
	}
	return unique != nil && (unique == declared || l.checker.IsTypeAssignableTo(unique, declared) && l.checker.IsTypeAssignableTo(declared, unique) && l.sameRecordStorage(unique, declared, map[[2]*checker.Type]bool{}) && l.sameRecordStorage(declared, unique, map[[2]*checker.Type]bool{}))
}

// recordSlotHeapClasses is each representation recordSlotMemberUnique tells apart at runtime: a scalar by
// its box, a string, an array, a Map and a closure by their heap kinds, and a fixed object from a
// dictionary by its shape.
var recordSlotHeapClasses = map[ir.Type]bool{ir.Number: true, ir.Boolean: true, ir.String: true, ir.Array: true, ir.Map: true, ir.Closure: true, ir.Object: true, ir.Record: true}

// An unrestricted key can hit a narrower named slot. Preserve that slot's
// contract instead of trusting the checker's permissive index assignment.
func (l *lowering) recordNamedWrite(access, value *ast.Node) error {
	access = ast.SkipParentheses(access)
	if access.Kind != ast.KindElementAccessExpression {
		return nil
	}
	a := access.AsElementAccessExpression()
	key := ast.SkipParentheses(a.ArgumentExpression)
	if key.Kind == ast.KindStringLiteral || key.Kind == ast.KindNoSubstitutionTemplateLiteral || key.Kind == ast.KindNumericLiteral {
		return nil // The checker already applies the named property's write type.
	}
	return l.namedWriteStop(access, l.checker.GetTypeAtLocation(a.Expression), value, nil)
}

// recordComputedEntry is recordNamedWrite for a record literal's entry under a computed name, which is
// written as the literal is made: { names: new Point(), [key]: new Other() } put an Other under names
// when key was "names", and native ran Point's describe on it, as w8's assignment did. The entry is
// judged against the named properties of the literal's own type and of the type it's written into.
// A name strings joined by + make (constantStringKey) is judged as that name alone, so tsc's
// ["" + "constructor"] beside the keywords it declares (01_scanner_keywords.a) names none of them.
// A method or an accessor under such a name stores a function under it unjudged, so it's not yet
// unless its folded name names no declared property.
//
// In a slot boxed as a union (of), a declared property whose read is proven to stop on a lie there is
// left to its read (readGuardedInUnionSlot). Every other one is judged, an unknown representation
// included: a function's tag says only "function", so a () => void under names: () => unknown was
// read as unknown.
func (l *lowering) recordComputedEntry(literal, entry *ast.Node, of ir.Type) error {
	key := ast.SkipParentheses(entry.Name().AsComputedPropertyName().Expression)
	if key.Kind == ast.KindStringLiteral || key.Kind == ast.KindNoSubstitutionTemplateLiteral || key.Kind == ast.KindNumericLiteral {
		return nil
	}
	var folded map[string]bool
	if name, known := constantStringKey(key); known {
		folded = map[string]bool{name: true}
	}
	receivers := []*checker.Type{l.checker.GetTypeAtLocation(literal)}
	if contextual := l.slotType(literal); contextual != nil {
		receivers = append(receivers, l.checker.GetNonNullableType(contextual))
	}
	for _, receiver := range receivers {
		names := map[string]bool{}
		for _, property := range l.checker.GetPropertiesOfType(receiver) {
			if folded != nil && !folded[property.Name] {
				continue
			}
			if of == ir.Union {
				if stored, known := l.representation(l.checker.GetNonMissingTypeOfSymbol(property)); known && readGuardedInUnionSlot[stored.Present()] {
					continue
				}
			}
			names[property.Name] = true
		}
		if len(names) == 0 {
			continue
		}
		if entry.Kind != ast.KindPropertyAssignment {
			return l.notYet(entry.Name(), "a computed method or accessor in a record literal whose name a named property may take")
		}
		if err := l.namedWriteStop(entry.Name(), receiver, entry.AsPropertyAssignment().Initializer, names); err != nil {
			return err
		}
	}
	return nil
}

// readGuardedInUnionSlot is each representation whose declared read from a slot boxed as a union is
// proven to stop on a lie, each by a pin in TestNamedRecordWritesAreJudgedAsTheirDirectTwins. A number,
// string or boolean is read through recordReadType's exact tag check, which stops loudly
// (records_union_lying.ts is Adamic's .ts witness). A plain object or a dictionary that's the one
// member of its heap class in the slot's union is read through a checked heap narrow
// (recordReadHeapMember), which stops on the other class; beside another of its class it has no
// checked tag, and neither has an array or a Map, so their read is not yet: Language's b8, a8 and p8,
// and r8, the MapLike that records_union_options.ts, tsc's commandLineParser shape, declares. A function's tag says only
// "function", so it's judged (b10), as is any representation not here.
var readGuardedInUnionSlot = map[ir.Type]bool{ir.Number: true, ir.String: true, ir.Boolean: true, ir.Object: true, ir.Array: true, ir.Map: true, ir.Record: true}

// namedWriteStop stops a value written under a name the key can't promise into receiver, through its
// index signature, at the first named property of receiver (of those names holds, when it's given) it
// doesn't fit, or returns nil. The value lands under that name, so it's judged as written into that
// property: the whole refusal registry runs over it with the property's type as its slot (refuseInSlot,
// #b6m00y7), and it stops as the direct write to the property would. A second list of slot rules here
// failed open on the rules it didn't name: nominal ancestry let an Other into names: Point, whose method
// native then ran on it (Language's w8), and void widening let a () => void into names: () => unknown
// (w10). Lowering's own slot checks at the value run against the property too (valueSlotStops), since
// the value is lowered against the index type. What stays here is what no rule at the direct write
// judges, since a direct write can't do it:
// a readonly property written through the index, a value tsc's own relation doesn't take (the registry
// presumes the site passed it), and a value the record's one storage can't hold as that property does.
func (l *lowering) namedWriteStop(where *ast.Node, receiver *checker.Type, value *ast.Node, names map[string]bool) error {
	from := l.checker.GetTypeAtLocation(value)
	for _, property := range l.checker.GetPropertiesOfType(receiver) {
		if names != nil && !names[property.Name] {
			continue
		}
		to := l.checker.GetNonMissingTypeOfSymbol(property)
		if l.checker.IsReadonlySymbol(property) || !l.checker.IsTypeAssignableTo(from, to) || !l.sameRecordStorage(from, to, map[[2]*checker.Type]bool{}) {
			return l.notYet(where, "an unrestricted dictionary write that can violate named property "+property.Name+"; use a literal key or a value accepted by every named property")
		}
		if err := l.refuseInSlot(value, to); err != nil {
			return err
		}
		// Lowering judges the value against the index type it's stored as, so its own slot checks (a
		// host object's view, a weak or strong keep, a tuple seen as an array) are run here against the
		// declared property too (Lower's probe: a host object under names: View, its index Clock). A value
		// whose representation isn't known is judged as a reference, so this never leans on its own
		// lowering stopping it later (Lower's should on fffa97b5).
		held, known := l.representation(from)
		if err := l.valueSlotStops(value, to, !known || held.IsReference()); err != nil {
			return err
		}
	}
	return nil
}

// Erasing or widening a named contract would allow another alias to write through
// the index signature. Storage identity alone is not a proof of that relation.
func (l *lowering) sameRecordNamedContracts(from, to, inside, viewed *checker.Type, visited map[[2]*checker.Type]bool) bool {
	for index, pair := range [][3]*checker.Type{{from, to, viewed}, {to, from, inside}} {
		if len(l.checker.GetIndexInfosOfType(pair[0])) == 0 {
			continue // Finite partial records retain their existing element invariance rule.
		}
		for _, property := range l.checker.GetPropertiesOfType(pair[0]) {
			other := l.checker.GetPropertyOfType(pair[1], property.Name)
			t := l.checker.GetNonMissingTypeOfSymbol(property)
			if other == nil {
				// A .ts view may drop the source's own contract (droppedContractRead), never add one.
				if (l.checker.IsReadonlySymbol(property) || !l.checker.IsTypeAssignableTo(pair[2], t)) && !(index == 0 && l.droppedContractRead(t, inside, viewed)) {
					return false
				}
				continue
			}
			u := l.checker.GetNonMissingTypeOfSymbol(other)
			if l.checker.IsReadonlySymbol(property) != l.checker.IsReadonlySymbol(other) || !l.checker.IsTypeAssignableTo(t, u) || !l.checker.IsTypeAssignableTo(u, t) || !l.sameRecordStorage(t, u, visited) || l.widened(t, u, map[[2]*checker.Type]bool{}) != nil || l.widened(u, t, map[[2]*checker.Type]bool{}) != nil {
				return false
			}
		}
	}
	return true
}

// droppedContractRead is whether a .ts view (droppedContractsRead, set by recordStorageView) may drop
// a named property of type t, declared by the record seen, whose index signature holds held, through a
// record whose index signature holds seen: tsc's CompilerOptions and WatchOptions seen as OptionsBase
// (commandLineParser.ts's serializeOptionBaseObject and handleOptionConfigDirTemplateSubstitution), and
// CompilerOptions seen as MapLike (builder.ts's getOwnKeys). No copy is made, so the object keeps its
// identity. What the view writes under the name through its index signature is any value of seen, and
// isn't stopped at the write: the reads hold the contract, as recordComputedEntry leaves a computed
// entry to them, on Language's three conditions (#6c4swvt, 2026-11-06 03:08Z):
//   - (a) Both index value types keep one representation, so a read through either reads the slot as it
//     is stored, and keep everything inside them alike (sameKeeping).
//   - (b) The property is one of its own index signature's values, so its value is stored at the index's
//     representation and a read through the index reads its bytes.
//   - (c) Every named read of the record's value is checked at run time by its declared type, whatever
//     the receiver it's read through: the slot is boxed as a union, so each read goes through
//     recordReadType, and the property is a scalar that check proves by its exact tag or value
//     (droppedContractScalars). A function's tag says only "function", and a union is read unchecked.
//     An object kind (a MapLike, an array, a Map, an object) is never dropped, Adamic's condition on
//     Language's 08:38Z must: an honest MapLike written through the view under paths read as a
//     non-member of the slot's union natively, where Node and the JavaScript backend read it.
//
// The flag is cleared while a function's parameters are walked (parameterStorage): a parameter is
// handed the other side's value, so the view there runs against the flow and would add the contract
// it seems to drop, Language's 08:38Z should.
//
// A readonly property is dropped on the same conditions: the view may write under its name, and the
// read holds what it finds there as it does any other write through an index signature. The other way,
// a record seen as one declaring a property its index signature can't hold (Record<string,
// CompilerOptionsValue> seen as CompilerOptions, or the as CompilerOptions back-cast), keeps the stop.
func (l *lowering) droppedContractRead(t, held, seen *checker.Type) bool {
	if !l.droppedContractsRead {
		return false
	}
	of, known := l.recordStorageType(held)
	viewedOf, viewedKnown := l.recordStorageType(seen)
	if !known || !viewedKnown || of != viewedOf || !l.sameKeeping(held, seen, map[[2]*checker.Type]bool{}) {
		return false
	}
	if !l.checker.IsTypeAssignableTo(t, held) {
		return false
	}
	stored, known := l.representation(t)
	return of == ir.Union && known && droppedContractScalars[stored.Present()]
}

// droppedContractScalars is each representation droppedContractRead drops a property of: a scalar whose
// read from a slot boxed as a union recordReadDeclared proves by its exact tag, and by its value where
// the declared type is literals (a number, string or boolean, their literal unions and numeric enums).
var droppedContractScalars = map[ir.Type]bool{ir.Number: true, ir.String: true, ir.Boolean: true}

// parameterStorage is check run with droppedContractsRead cleared, then restored: a function's
// parameter is handed the other side's value, so no named contract is dropped through it, in either
// direction.
func (l *lowering) parameterStorage(check func() bool) bool {
	dropping := l.droppedContractsRead
	l.droppedContractsRead = false
	defer func() { l.droppedContractsRead = dropping }()
	return check()
}
