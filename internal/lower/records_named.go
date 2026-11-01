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
	file := ast.GetSourceFileOfNode(node)
	expression := file.Text()[scanner.GetTokenPosOfNode(node, file, false):node.End()]
	declared := l.checker.TypeToString(l.checker.GetNonNullableType(here))
	found := ir.Conditional{Condition: ir.IsNull{Value: value}, WhenTrue: ir.StringConstant{Index: l.constant("null")}, WhenNot: ir.TypeOf{Value: value}, Of: ir.String}
	message := ir.Concat{Parts: []ir.Expression{
		ir.StringConstant{Index: l.constant("field read failed: " + expression + " is not a " + declared + "; expected " + declared + ", found ")},
		found,
		ir.StringConstant{Index: l.constant(", which a computed-key or index-signature write may have stored")},
	}}
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: message}}})
	return b.finish("record_declared_member", ir.Narrow{Value: value, To: to}), nil
}

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
// (records_union_lying.ts is Adamic's .ts witness). A plain object, an array, a Map and a dictionary
// have no checked tag, so their read is not yet: Language's b8, a8 and p8, and r8, the MapLike that
// records_union_options.ts, tsc's commandLineParser shape, declares. A function's tag says only
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
	for _, pair := range [][3]*checker.Type{{from, to, viewed}, {to, from, inside}} {
		if len(l.checker.GetIndexInfosOfType(pair[0])) == 0 {
			continue // Finite partial records retain their existing element invariance rule.
		}
		for _, property := range l.checker.GetPropertiesOfType(pair[0]) {
			other := l.checker.GetPropertyOfType(pair[1], property.Name)
			t := l.checker.GetNonMissingTypeOfSymbol(property)
			if other == nil {
				if l.checker.IsReadonlySymbol(property) || !l.checker.IsTypeAssignableTo(pair[2], t) {
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
