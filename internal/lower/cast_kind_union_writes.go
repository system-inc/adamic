package lower

import (
	"reflect"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A named write through a kind-checked cast's union, `(node as PropertyAccessExpression |
// ElementAccessExpression).flags = 5`, is #b5w3ycg slice 4's head 2. Adamic ruled it on Nov 17 as Rule C
// through a kind-checked view (rulings.md:73, its four conditions): the tag can't move, since a write to it is
// refused; a lie stops at the write, since main's checked write checks presence there, exit 70 on both
// backends; Rule C's record is main's road; and, the fourth, every declared type carrying any of the cast's
// kind literals declares the field with one written type, and each such layout holds the slot. Compiler
// confirmed on Nov 18 that the fourth is static, judged here at compile time over the program's declared
// types. A run-time stop isn't enough for it: two interfaces sharing a kind and declaring the field with
// different types (`{ text: string }` and `{ text: number }`, both kind StringLiteral) make a valid program
// Node runs, and a write of a string through the union would land in the number's layout, so that program is
// NotYet, never a stop at run time (the head 2 probe's w3). A declared type carrying a cast's kind and
// declaring no such field holds no slot for it, which waits under #x0nd1mh. One declaring it optional is judged
// as every optional field's write is, by whether each object reaching the write holds its slot (slot_writes.go,
// #w8nanzw), its type compared with the others' as any carrier's is. One declaring
// the field readonly holds the slot but promised it never changes, a promise Adamic's immutable default leans on
// (sharing, hoisted reads), so the write is NotYet there too, whatever the members declare. The written
// types are compared across every carrier, not only those sharing one literal, so main's checked-write road
// never meets members whose slots differ.
//
// Adamic amended the fourth condition on Nov 19, from Lower's rows: a value can be built at a type no
// declaration names (an annotated literal type, an inline intersection, a return's or a parameter's literal
// type, a kind written as const, or a literal whose kind widens to the open SyntaxKind, `as Node` or not), so a
// list of declared types is a claim about types while the hazard lives in storage. The write is admitted only
// when every allocation that can reach its receiver (ReachingAllocations) is an object the program builds whose
// own slot holds the field (reachHoldsSlot), and whose type holds it writable and at a type the members' written
// type relates to both ways (slotWriteSite, every slot write's); an unknown reach is NotYet. The declared index stays a fast
// refusal and never admits on its own: a write it doesn't refuse is recorded (kindUnionWrites' recordSlotWrite)
// and judged once lowering is done, on the allocation graph (checkSlotWrites), as checkBoundArguments judges a
// bind. A frozen carrier (Object.freeze) never reaches
// the write, since invariance refuses its Readonly<...> type seen as the writable union first; reach itself doesn't
// ask about freezing (Language's s03).
//
// This head admits a named write alone: an assignment to the field, plain or compound, and an increment or a
// decrement. An element write (`u[k] = v`) and a field written as a destructuring's or a loop's target stay
// NotYet (kindUnionWrites), each a road main's checked write hasn't been measured on through a union; a delete
// is refused before, as any delete is.

// kindTagCarriers is every type a module of the program declares, an interface, a class or a type alias of an
// object type or an intersection, whose field holds the literal: by its runtime value, as tagLiteralsDistinct
// compares tags, so two enums' members with one value meet. An intersection alias (`Node & { kind: K.A; text:
// number }`) is a layout a value can be built at as much as an interface is, so leaving it out would let a valid
// program Node runs stop at run time, w3's class. A field holding a whole enum, open as Node's kind is, carries no literal
// (tagHoldsAnyNumber): a value reaching such a type was built at one that does (enum_tag_closed_kind.go's first
// half). Every file the program holds is walked, a module it only imports types from and a declaration file among
// them, never the bundled library or the prelude (load.Program.SourceFiles). The index is built for a field the
// first time a write through a cast tagged by it is judged, never before, so a program with no such write
// resolves no declared type it didn't already.
func (l *lowering) kindTagCarriersOf(field string, literal *checker.Type) ([]*checker.Type, error) {
	key, keyed := kindTagKey(l.checker, field, literal)
	if !keyed {
		return nil, nil
	}
	if !l.kindTagFields[field] {
		// Every file the program holds, never moduleOrder's runnable modules alone: a type a module declares and
		// the program only imports as a type is a layout a value can be built at, and moduleOrder leaves its
		// module out (Lower's must on head 2, the two-file Odd pin).
		modules := l.program.SourceFiles()
		if l.kindTagCarriers == nil {
			l.kindTagCarriers = map[string][]*checker.Type{}
		}
		if l.kindTagFields == nil {
			l.kindTagFields = map[string]bool{}
		}
		l.kindTagFields[field] = true
		seen := map[*ast.Symbol]bool{}
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			switch node.Kind {
			case ast.KindInterfaceDeclaration, ast.KindClassDeclaration, ast.KindTypeAliasDeclaration:
				if symbol := l.checker.GetSymbolOfDeclaration(node); symbol != nil && !seen[symbol] {
					seen[symbol] = true
					l.indexKindTagCarrier(field, l.checker.GetDeclaredTypeOfSymbol(symbol))
				}
			}
			node.ForEachChild(visit)
			return false
		}
		for _, module := range modules {
			module.AsNode().ForEachChild(visit)
		}
	}
	return l.kindTagCarriers[key], nil
}

// indexKindTagCarrier adds declared to the index under each literal its field holds.
func (l *lowering) indexKindTagCarrier(field string, declared *checker.Type) {
	if declared == nil || declared.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		return
	}
	property := l.checker.GetPropertyOfType(declared, field)
	if property == nil {
		return
	}
	held := l.checker.GetNonNullableType(l.checker.GetTypeOfSymbol(property))
	if l.tagHoldsAnyNumber(held) {
		return
	}
	for _, member := range castMembers(held) {
		if member.Flags()&(checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBooleanLiteral) == 0 {
			continue
		}
		key, keyed := kindTagKey(l.checker, field, member)
		if !keyed {
			continue
		}
		carriers := l.kindTagCarriers[key]
		// Two members of one value (an enum's alias) name the type once.
		if len(carriers) == 0 || carriers[len(carriers)-1] != declared {
			l.kindTagCarriers[key] = append(carriers, declared)
		}
	}
}

// kindTagKey is a field and a literal's runtime value as one key: the representation, then the value.
func kindTagKey(typeChecker *checker.Checker, field string, literal *checker.Type) (string, bool) {
	flags := literal.Flags()
	switch {
	case flags&checker.TypeFlagsStringLiteral != 0:
		text, isText := literal.AsLiteralType().Value().(string)
		return field + "\x00s" + text, isText
	case flags&checker.TypeFlagsNumberLiteral != 0:
		value := reflect.ValueOf(literal.AsLiteralType().Value())
		if value.Kind() != reflect.Float64 {
			return "", false
		}
		number := value.Float()
		// A test compares by value, and -0 is 0 to it.
		if number == 0 {
			number = 0
		}
		return field + "\x00n" + strconv.FormatFloat(number, 'g', -1, 64), true
	case flags&checker.TypeFlagsBooleanLiteral != 0:
		return field + "\x00b" + typeChecker.TypeToString(literal), true
	}
	return "", false
}

// kindUnionWriteHeld judges the named write part, of name, through the cast proof's union by the ruling's fourth
// condition: nil admits it to main's checked-write road. Every carrier of the cast's literals is asked, the cast's
// own members among them, since a member written inline in the cast is no declared type the index holds.
func (l *lowering) kindUnionWriteHeld(part *ast.Node, proof *kindUnion, name string) error {
	type carrier struct {
		of, literal *checker.Type
	}
	var carriers []carrier
	held := map[*checker.Type]bool{}
	for _, literal := range proof.literals {
		found, err := l.kindTagCarriersOf(proof.field, literal)
		if err != nil {
			return err
		}
		for _, member := range proof.members {
			if l.fieldLiteral(member, proof.field) == literal {
				found = append(found, member)
			}
		}
		for _, of := range found {
			if !held[of] {
				held[of] = true
				carriers = append(carriers, carrier{of: of, literal: literal})
			}
		}
	}
	write := "a write of " + name + " through a kind-checked cast's union"
	for _, carrier := range carriers {
		property := l.checker.GetPropertyOfType(carrier.of, name)
		// A carrier declaring name optional holds a slot for it where the object was built with one, which reach
		// judges once the program is lowered (checkSlotWrites).
		if property == nil || property.Flags&ast.SymbolFlagsProperty == 0 {
			return l.notYet(part, write+", where "+l.checker.TypeToString(carrier.of)+" carries "+proof.field+" "+l.checker.TypeToString(carrier.literal)+" but its layout holds no "+name)
		}
		if l.checker.IsReadonlySymbol(property) {
			return l.notYet(part, write+", where "+l.checker.TypeToString(carrier.of)+" carries "+proof.field+" "+l.checker.TypeToString(carrier.literal)+" but declares "+name+" readonly")
		}
	}
	// A pair sharing a literal is named first, since that's the pair a kind test can't tell apart.
	for _, sharing := range []bool{true, false} {
		for first := range carriers {
			for second := first + 1; second < len(carriers); second++ {
				one, other := carriers[first], carriers[second]
				if (one.literal == other.literal || !tagLiteralsDistinct(one.literal, other.literal)) != sharing {
					continue
				}
				oneType := l.checker.GetTypeOfSymbol(l.checker.GetPropertyOfType(one.of, name))
				otherType := l.checker.GetTypeOfSymbol(l.checker.GetPropertyOfType(other.of, name))
				if oneType == otherType || (l.castAssignable(oneType, otherType) && l.castAssignable(otherType, oneType)) {
					continue
				}
				declares := "declare " + name + " as " + l.checker.TypeToString(oneType) + " and " + l.checker.TypeToString(otherType)
				if sharing {
					return l.notYet(part, write+", where "+l.checker.TypeToString(one.of)+" and "+l.checker.TypeToString(other.of)+" share "+proof.field+" "+l.checker.TypeToString(one.literal)+" but "+declares)
				}
				return l.notYet(part, write+", where "+l.checker.TypeToString(one.of)+" carries "+proof.field+" "+l.checker.TypeToString(one.literal)+" and "+l.checker.TypeToString(other.of)+" "+l.checker.TypeToString(other.literal)+", but they "+declares)
			}
		}
	}
	return nil
}

// kindUnionNamedWrite is whether part, a property access a write targets, is a named write this head admits:
// the left side of an assignment, plain or compound, or the operand of an increment or a decrement.
func kindUnionNamedWrite(part *ast.Node) bool {
	if part.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	target := part
	for target.Parent.Kind == ast.KindParenthesizedExpression {
		target = target.Parent
	}
	switch target.Parent.Kind {
	case ast.KindBinaryExpression:
		binary := target.Parent.AsBinaryExpression()
		return binary.Left == target && ast.IsAssignmentOperator(binary.OperatorToken.Kind)
	case ast.KindPrefixUnaryExpression:
		operator := target.Parent.AsPrefixUnaryExpression().Operator
		return operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken
	case ast.KindPostfixUnaryExpression:
		operator := target.Parent.AsPostfixUnaryExpression().Operator
		return operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken
	}
	return false
}

// builtType is the type an object literal is built at: layout, the slot it's written into where that's an
// object type or a union or an intersection of them, whose fields lay out the literal's storage
// (declaredField), and own, the literal's own type, set only where it isn't layout. A zero builtType is a
// literal whose type can't be named.
type builtType struct {
	layout, own *checker.Type
}

// noteBuiltType keeps the type the object literal whose own type is own, of id, is built at, contextual being
// the slot it's written into, nil for none (graphAllocation). A literal's own type is fresh, so no two
// literals share one; where two lowerings of one literal see two slots, the type can't be named, and a write
// it reaches fails closed. The map is keyed by that own type's id on the premise that fresh literal types are
// never shared: if one ever were, the two literals would zero the entry and every write they reach would fail
// closed, a cost and never a hole (Lower's read of head 2).
func (l *lowering) noteBuiltType(id int, own, contextual *checker.Type) {
	built := builtType{layout: own}
	if contextual != nil {
		if contextual = l.concrete(contextual); contextual.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			built = builtType{layout: contextual, own: own}
		}
	}
	if l.builtTypes == nil {
		l.builtTypes = map[int]builtType{}
	}
	if previous, seen := l.builtTypes[id]; seen && previous != built {
		built = builtType{}
	}
	l.builtTypes[id] = built
}

// kindUnionWritten is the type the cast proof's members declare name written at, the first's, since one cast's
// members declare it alike (kindUnionWriteHeld), or nil where none declares it.
func (l *lowering) kindUnionWritten(proof *kindUnion, name string) *checker.Type {
	for _, member := range proof.members {
		if property := l.checker.GetPropertyOfType(member, name); property != nil {
			return l.checker.GetTypeOfSymbol(property)
		}
	}
	return nil
}
