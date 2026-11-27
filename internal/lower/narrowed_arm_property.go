package lower

import (
	"reflect"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// narrowedArmPropertyReadGap is narrowedArmReadGap's rule for a property read, m.n, this.n, a.m.n or m['n'],
// as narrowedNullishPropertyReadGap is the nullish rule's (#8nj5pn4). The checker narrows a property by a
// discriminant check and keeps the narrowing across a call that writes the property again:
//
//	function reset(m: M): void { m.n = b; }
//	if (m.n.kind === 'a') { reset(m); return String(m.n.a + 1); }
//
// Node prints NaN, since m.n holds b by the read, and native read b's storage through a's layout and printed 1.
// A property has no declaration that keeps its writes in one function, so it's proved by name, as the nullish
// twin proves it: a getter of the name stops the read first, since a getter answers each read itself. Then
// the read lowers when the place it was last seen narrowed (an assignment, or a check that mentions it,
// armObservationOf) is followed by nothing that can run program code before the read, or by nothing that can
// write its name or install an accessor (propertyWriteMayRunBetween). A read with no such place is judged
// from the start of its own function.
// It's a fail-closed stop: a checked view of the declared union is the read that lifts it.
func (l *lowering) narrowedArmPropertyReadGap(node *ast.Node) error {
	if !l.isExpression(node) || (ast.IsAssignmentTarget(node) && pureWrite(node)) {
		return nil
	}
	// The read's name, and its property where it has one. An element's key is read as tsc narrows it: a literal,
	// or any key whose type is a string or number literal, a const index among them (Language's e02).
	var name string
	var property *ast.Symbol
	if node.Kind == ast.KindPropertyAccessExpression {
		name = node.Name().Text()
		property = l.checker.GetSymbolAtLocation(node.Name())
	} else {
		key := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		switch key.Kind {
		case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindNumericLiteral:
			name = key.Text()
			property = l.checker.GetSymbolAtLocation(key)
		default:
			keyType := l.checker.GetTypeAtLocation(key)
			switch {
			case keyType.Flags()&checker.TypeFlagsNumberLiteral != 0:
				// Its value, never its spelling: an enum member's literal type prints as Slot.First (Language's e05).
				value := reflect.ValueOf(keyType.AsLiteralType().Value())
				if value.Kind() != reflect.Float64 {
					return nil
				}
				name = strconv.FormatFloat(value.Float(), 'f', -1, 64)
			case keyType.Flags()&checker.TypeFlagsStringLiteral != 0:
				text, isText := keyType.AsLiteralType().Value().(string)
				if !isText {
					return nil
				}
				name = text
			default:
				return nil // A key of no literal type narrows nothing.
			}
			property = l.checker.GetPropertyOfType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node.AsElementAccessExpression().Expression)), name)
		}
	}
	narrowed := l.checker.GetTypeAtLocation(node)
	var declared *checker.Type
	if property != nil {
		declared = l.checker.GetTypeOfSymbol(property)
	} else if _, numberError := strconv.ParseFloat(name, 64); node.Kind == ast.KindElementAccessExpression && numberError == nil {
		// An array's element by a literal index, arr[0], has no property of its own: its declared type is the
		// array's element type, and a write by any index or an array method may rewrite it (Language's r08).
		receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node.AsElementAccessExpression().Expression))
		if !l.isArrayType(receiver) {
			return nil
		}
		declared = l.elementTypeOfArray(receiver)
	}
	if declared == nil {
		return nil // No declared property or element, so no declared type the checker narrowed from.
	}
	if narrowed.Id() == declared.Id() || narrowed.Flags()&checker.TypeFlagsNever != 0 {
		return nil
	}
	if _, nullish := l.nullishUnit(narrowed); nullish {
		return nil // narrowedNullishPropertyReadGap's.
	}
	present := l.checker.GetNonNullableType(declared)
	if present.Id() == narrowed.Id() || l.checker.GetNonNullableType(narrowed).Id() == present.Id() {
		// A presence strip, checked or observed at run time, as the binding rule says, or the declared type with
		// only undefined added, an element read under noUncheckedIndexedAccess, which narrows nothing (Language's e01).
		return nil
	}
	if storage, known := l.representation(declared); known && storage == ir.Union {
		return nil // The held union's member is checked at run time before it's narrowed.
	}
	if narrowedStorage, narrowedKnown := l.representation(narrowed); narrowedKnown && narrowedStorage != ir.Object {
		return nil // Only an object arm reads through an object's layout: a scalar, a string, an array, a closure or a record arm doesn't. An arm whose storage isn't known is judged.
	}
	if property != nil && property.Flags&ast.SymbolFlagsEnumMember != 0 {
		return nil // An enum's member is a constant no write reaches.
	}
	stop := l.notYet(node, "a read of property "+name+" narrowed to "+l.checker.TypeToString(narrowed)+" where a call may have rewritten it")
	// A getter of the name answers each read itself, so it stops the read whatever the flow says. A program that
	// could install one some other way (writes.accessors: a descriptor, a prototype) stops it only where program
	// code can run between the narrowing and the read, since nothing else can install it there.
	writes := l.propertyWritesOf()
	if l.accessorName(name) {
		return stop
	}
	if !l.armReadMaySilentlyMisread(node, narrowed, declared) {
		// A stale arm is read by the value's own shape or by a slot every layout holding the name shares, and
		// a field the new value lacks stops loudly there (Emit's guard on #8nj5pn4), so only a field held at
		// two representations can read the wrong bits silently.
		return nil
	}
	_, notNumeric := strconv.ParseFloat(name, 64)
	numeric := notNumeric == nil
	observation := l.armObservationOf(node)
	if observation == nil {
		// No one place the narrowing was seen (a join of two checks, as a ternary chain's), so what may run is
		// judged from the start of the read's own code, where every narrowing of it begins. A read with no such
		// start stops if anything can write its name.
		boundary := callBoundary(node)
		if boundary == nil || boundary.Pos() < 0 {
			if numeric || writes.everyName || len(writes.named[name]) > 0 {
				return stop
			}
			return nil
		}
		observation = &nullishObservation{from: boundary.Pos()}
	}
	exempt := l.appendsBetween(observation, node, numeric)
	if writes.accessors && l.codeBetween(observation.from, node, exempt) {
		return stop
	}
	if l.propertyWriteMayRunBetween(node, name, numeric, observation.from, observation.target, exempt) {
		return stop
	}
	return nil
}

// armObservationOf walks the read's flow straight back to the place it was last seen narrowed: a plain
// assignment of it, after which it holds what was assigned, or a condition that mentions it, the check that
// narrowed it. A call, which codeBetween judges, an assignment of something else and a join whose every way
// in reaches the same place are walked through. A loop, a label of any other kind, a write of what holds the
// read or the start of a function ends the walk with none. A condition's target is nil, since it writes
// nothing to exempt.
func (l *lowering) armObservationOf(read *ast.Node) *nullishObservation {
	data := read.FlowNodeData()
	if data == nil || data.FlowNode == nil {
		return nil
	}
	return l.armObservationAlong(read, data.FlowNode, 0, map[*ast.FlowNode]*nullishObservation{})
}

func (l *lowering) armObservationAlong(read *ast.Node, flow *ast.FlowNode, joins int, seen map[*ast.FlowNode]*nullishObservation) *nullishObservation {
	for flow != nil {
		switch {
		case flow.Flags&ast.FlowFlagsAssignment != 0:
			if predicateMatchingReference(l.checker, read, flow.Node) {
				parent := flow.Node.Parent
				if parent != nil && parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken && parent.AsBinaryExpression().Left == flow.Node {
					return &nullishObservation{target: flow.Node, from: parent.End(), assigned: true}
				}
				return nil
			}
			if enumTagContainsReference(l.checker, read, flow.Node) || l.mentionsReference(read, flow.Node) {
				return nil
			}
		case flow.Flags&ast.FlowFlagsCondition != 0:
			if l.mentionsReference(read, flow.Node) {
				return &nullishObservation{from: flow.Node.End()}
			}
		case flow.Flags&(ast.FlowFlagsCall|ast.FlowFlagsArrayMutation) != 0:
		case flow.Flags&ast.FlowFlagsBranchLabel != 0:
			if found, walked := seen[flow]; walked {
				return found
			}
			seen[flow] = nil
			if joins >= nullishObservationDepth {
				return nil
			}
			var common *nullishObservation
			for list := flow.Antecedents; list != nil; list = list.Next {
				if list.Flow.Flags&ast.FlowFlagsUnreachable != 0 {
					continue
				}
				found := l.armObservationAlong(read, list.Flow, joins+1, seen)
				if found == nil || common != nil && found.from != common.from {
					return nil
				}
				common = found
			}
			seen[flow] = common
			return common
		default:
			return nil
		}
		flow = flow.Antecedent
	}
	return nil
}

// armReadMaySilentlyMisread is whether a read through a property narrowed to an object arm could read a stale
// value's field at the wrong representation, the one silent case Compiler left the floor on #8nj5pn4: the field
// read off it, or, where the narrowed value is handed on whole, any field its arm declares, is held at two
// representations by the types a stale value could be. Those are the types a value is made or held as, a class's
// instance type, an object literal's, an intersection the program writes, an interface's or a type literal's,
// that are assignable to the read's declared type, Node for node.parent, since only one of them can be what the
// call left there (fieldHoldersOf). Each field is read through the holder's own type, so one it inherits from a
// base class, takes from an interface it extends or from an intersection's member, or copies by a spread, counts
// at the representation the holder gives it, whatever type declared it (Compiler's must on -11). A field whose
// representation isn't known counts as one of its own, and a computed key the checker can't spell counts as one
// when the type that declares it, or a holder that may be the value and is assignable to it, may be the value,
// failing closed.
func (l *lowering) armReadMaySilentlyMisread(read *ast.Node, narrowed, declared *checker.Type) bool {
	at := read
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	var names []string
	if parent := at.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == at {
		names = []string{parent.Name().Text()}
	} else {
		for _, field := range l.checker.GetPropertiesOfType(narrowed) {
			names = append(names, field.Name)
		}
	}
	holders := l.fieldHoldersOf()
	for _, name := range names {
		held := map[ir.Type]bool{}
		var mayHold []*checker.Type
		for _, holder := range holders.named[name] {
			if !holders.assignable(l, holder, declared) {
				continue
			}
			mayHold = append(mayHold, holder)
			held[l.fieldStorage(holder, name)] = true
		}
		if len(held) > 1 {
			return true
		}
		if len(held) == 0 {
			continue
		}
		for _, owner := range holders.unspelled {
			if owner == nil || holders.assignable(l, owner, declared) {
				return true
			}
			for _, holder := range mayHold {
				if holders.assignable(l, holder, owner) {
					return true
				}
			}
		}
	}
	return false
}

// fieldStorage is the representation holder gives its field name, read through holder's own type, so an
// inherited or intersected field has the holder's view of it, with undefined added for an optional one, or 0
// where it isn't known.
func (l *lowering) fieldStorage(holder *checker.Type, name string) ir.Type {
	field := l.checker.GetPropertyOfType(holder, name)
	if field == nil {
		return 0
	}
	value := l.checker.GetTypeOfSymbol(field)
	if field.Flags&ast.SymbolFlagsOptional != 0 {
		value = l.checker.GetUnionType([]*checker.Type{value, l.checker.GetUndefinedType()})
	}
	if storage, known := l.representation(value); known {
		return storage
	}
	return 0
}

// fieldHolders is every type the program and its libraries make or hold a value as, by the name of each field
// it holds, own or not, and the type declaring each field whose name the checker can't spell, a computed key
// typed string but no literal, nil where that type can't be named. related keeps each assignability answer,
// since the same holder is asked about the same declared type at many reads.
type fieldHolders struct {
	named     map[string][]*checker.Type
	unspelled []*checker.Type
	related   map[[2]*checker.Type]bool
}

// assignable is from assignable to to, asked once per pair.
func (h *fieldHolders) assignable(l *lowering, from, to *checker.Type) bool {
	key := [2]*checker.Type{from, to}
	if answer, asked := h.related[key]; asked {
		return answer
	}
	answer := l.checker.IsTypeAssignableTo(from, to)
	h.related[key] = answer
	return answer
}

// fieldHoldersOf walks once per lowering every class, interface, object literal, type literal and written
// intersection, each holder's fields indexed by name through GetPropertiesOfType, and every field declaration
// whose name the checker can't spell, as sourceSlotDeclared reads them: a property, a method, an accessor, a
// parameter property or an object literal's member.
func (l *lowering) fieldHoldersOf() *fieldHolders {
	if l.armFieldHolders != nil {
		return l.armFieldHolders
	}
	found := &fieldHolders{named: map[string][]*checker.Type{}, related: map[[2]*checker.Type]bool{}}
	seen := map[*checker.Type]bool{}
	hold := func(holder *checker.Type) {
		if holder == nil || seen[holder] {
			return
		}
		seen[holder] = true
		for _, field := range l.checker.GetPropertiesOfType(holder) {
			found.named[field.Name] = append(found.named[field.Name], holder)
		}
	}
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		switch part.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression, ast.KindInterfaceDeclaration:
			if symbol := l.checker.GetSymbolOfDeclaration(part); symbol != nil {
				hold(l.checker.GetDeclaredTypeOfSymbol(symbol))
			}
		case ast.KindObjectLiteralExpression, ast.KindTypeLiteral:
			hold(l.checker.GetTypeAtLocation(part))
		case ast.KindIntersectionType:
			hold(l.checker.GetTypeFromTypeNode(part))
		case ast.KindPropertySignature, ast.KindPropertyDeclaration, ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment, ast.KindMethodSignature, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindParameter:
			if part.Kind == ast.KindParameter && !parameterProperty(part) {
				break
			}
			if _, named := l.sourceSlotDeclaredName(part); !named {
				found.unspelled = append(found.unspelled, l.fieldOwner(part))
			}
		}
		part.ForEachChild(visit)
		return false
	}
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		file.AsNode().ForEachChild(visit)
	}
	l.armFieldHolders = found
	return found
}

// fieldOwner is the type that declares a field's declaration, or nil where it can't be named.
func (l *lowering) fieldOwner(declaration *ast.Node) *checker.Type {
	holder := declaration.Parent
	if declaration.Kind == ast.KindParameter && holder != nil {
		holder = holder.Parent // A parameter property's constructor, then its class.
	}
	if holder == nil {
		return nil
	}
	switch holder.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression, ast.KindInterfaceDeclaration:
		if symbol := l.checker.GetSymbolOfDeclaration(holder); symbol != nil {
			return l.checker.GetDeclaredTypeOfSymbol(symbol)
		}
		return nil
	case ast.KindTypeLiteral, ast.KindObjectLiteralExpression:
		return l.checker.GetTypeAtLocation(holder)
	}
	return nil
}
