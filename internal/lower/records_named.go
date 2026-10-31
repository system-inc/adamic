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
	return l.namedWriteStop(access, l.checker.GetTypeAtLocation(a.Expression), l.checker.GetTypeAtLocation(value), nil)
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
// In a slot boxed as a union (of), a declared read already stops on a lie, so only a property read
// unchecked there is judged: recordReadType checks a declared scalar's tag and stops loudly
// (records_union_lying.ts is Adamic's .ts witness), and a declared object read from it is not yet;
// but a declared function is checked only by its tag, so a () => void under names: () => unknown
// would be read as unknown, and it's judged here.
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
	if contextual := l.checker.GetContextualType(literal, checker.ContextFlagsNone); contextual != nil {
		receivers = append(receivers, l.checker.GetNonNullableType(contextual))
	}
	for _, receiver := range receivers {
		names := map[string]bool{}
		for _, property := range l.checker.GetPropertiesOfType(receiver) {
			if folded != nil && !folded[property.Name] {
				continue
			}
			if of == ir.Union {
				if stored, known := l.representation(l.checker.GetNonMissingTypeOfSymbol(property)); known && stored.Present() != ir.Closure {
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
		if err := l.namedWriteStop(entry.Name(), receiver, l.checker.GetTypeAtLocation(entry.AsPropertyAssignment().Initializer), names); err != nil {
			return err
		}
	}
	return nil
}

// namedWriteStop is the not yet for a value of type from written under a name the key can't promise
// into receiver, through its index signature, when it doesn't fit one of its named properties
// (namedWriteMisfit), or nil.
func (l *lowering) namedWriteStop(where *ast.Node, receiver, from *checker.Type, names map[string]bool) error {
	property, identity := l.namedWriteMisfit(receiver, from, names)
	switch {
	case property == nil:
		return nil
	case identity:
		return l.notYet(where, "an unrestricted dictionary write of a value whose type isn't named property "+property.Name+"'s type itself; use a literal key or a value of exactly that type")
	}
	return l.notYet(where, "an unrestricted dictionary write that can violate named property "+property.Name+"; use a literal key or a value accepted by every named property")
}

// namedWriteMisfit is the first named property of receiver, of those names holds when it's given, that
// a value of type from, written under its name through the index signature, doesn't fit, and whether
// it's the identity check it fails; or nil.
//
// The list names some of Adamic's slot rules and fails open on those it doesn't: nominal ancestry let
// an Other into names: Point, whose method native then ran on it (Language's w8), and void widening
// let a () => void into names: () => unknown (w10). Until the registry judges the write (l.slotType,
// #b6m00y7), the value's widened type, its literals at their base type, must be each named property's
// type itself, as a delayed Array write's must be (delayedArrayWritesFit).
func (l *lowering) namedWriteMisfit(receiver, from *checker.Type, names map[string]bool) (*ast.Symbol, bool) {
	widened := checker.Checker_getBaseTypeOfLiteralType(l.checker, l.checker.GetWidenedType(from))
	for _, property := range l.checker.GetPropertiesOfType(receiver) {
		if names != nil && !names[property.Name] {
			continue
		}
		to := l.checker.GetNonMissingTypeOfSymbol(property)
		if l.checker.IsReadonlySymbol(property) || !l.checker.IsTypeAssignableTo(from, to) || !l.enumAssignable(from, to) || l.widened(from, to, map[[2]*checker.Type]bool{}) != nil || !l.sameRecordStorage(from, to, map[[2]*checker.Type]bool{}) {
			return property, false
		}
		if widened != to || !l.sameKeeping(widened, to, map[[2]*checker.Type]bool{}) {
			return property, true
		}
	}
	return nil, false
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
