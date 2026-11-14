package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A write of an object field through a checked view, view.f = value, puts value in a slot other aliases
// of the same object read too, each at the type its own declaration gives f. One that declares f
// narrower than value would read past what the slot holds (TestViewObjectWritesNeedSourceCertificate:
// raw declares child as { ready, other }, the view writes { ready }, and raw.child.other runs past the
// new object). The source-slot certificate is Adamic's Rule C on #2msw6t8, asked of the reads that exist:
//
//  1. A plain write `x.f = value;` through a checked view is admitted, and the type it writes is
//     recorded for the name f (recordSourceSlotWrite). Any other write keeps view's stop.
//  2. Once lowering is done, every read of f is held to every type written to it (certifySourceSlots).
//     A read whose declared type holds them all stays unchecked and costs nothing. Any other read is
//     checked under its receiver's key (CheckedFields), and the type it reads gets its view contract, so
//     the fields inside it are checked where they're read too, and a lie stops loudly at the read.
//
// Every road that reads a field by its name is a read site with its own declared type, or it keeps the
// write's stop:
//   - a property access, and a destructured field in a declaration, a parameter, a for-of, a nested
//     pattern or an assignment pattern: each an ir.Property that carries the field's declared type
//     (readObjectField, destructuredField's readers), held to the writes here;
//   - a const computed key: view refuses a viewed name read by a literal key (refuseViewElementReads),
//     and finiteElementRead's ir.Property carries no declared type, so it keeps the write's stop here;
//   - a spread or an Object.assign copies the slot into a new object, whose reads are sites of their
//     own, and checkViewCopies stops a copy of a receiver type a check was added for;
//   - JSON.stringify and encodeJson read every field by the declared type's schema, so one whose type
//     holds f at a type that doesn't hold every write keeps the write's stop (sourceSlotWholeReads);
//   - Object.values and Object.entries are admitted on a fixed object only for scalar values
//     (library_object.go), and a for-in yields keys, so neither reads an object field's value;
//   - Reflect.get, and a read of an unknown by name, give unknown, a value no type claims, which a
//     later narrowing tests as it runs;
//   - an accessor dispatcher's fallback is an unchecked load, and finishAccessors stops a checked read
//     that would go through it, so this runs before it;
//   - a read through a type parameter has no one declared type: the type its instantiation reads isn't
//     the one lowering records, so it keeps the write's stop (sourceSlotGenericReads), and so does a read
//     whose declared type is generic.
type sourceSlotWrite struct {
	target  *ast.Node
	name    string
	written *checker.Type
	stop    error
}

// recordSourceSlotWrite records target, an object field write through a checked view whose stop is
// stop, as written for certifySourceSlots, and says whether it was. Only the whole left side of a plain
// `=` statement is (hostMemberAssignment): a compound, a logical or a destructuring write keeps its stop,
// and so does one found after certifySourceSlots has judged the reads. A fresh object or array literal
// is made as the slot it's written into (freshOrWidened), so it's recorded as the field's own type.
func (l *lowering) recordSourceSlotWrite(target *ast.Node, field *ast.Symbol, stop error) bool {
	if l.sourceSlotsJudged {
		return false
	}
	assignment := hostMemberAssignment(target)
	if assignment == nil {
		return false
	}
	for _, write := range l.sourceSlotWrites {
		if write.target == target {
			return true
		}
	}
	value := assignment.AsBinaryExpression().Right
	written := l.checker.GetTypeAtLocation(value)
	switch ast.SkipParentheses(value).Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		written = l.checker.GetTypeOfSymbol(field)
	}
	l.sourceSlotWrites = append(l.sourceSlotWrites, sourceSlotWrite{target: target, name: target.Name().Text(), written: written, stop: stop})
	return true
}

// certifySourceSlots holds every read of a name recordSourceSlotWrite recorded to every type written to
// it, once lowering is done: a read whose declared type holds them all stays as it is, and any other is
// checked under its receiver's key, its declared type given a view contract. A read lowering can't name
// a declared type for, or one that has no receiver key to check it under, gives back the first write's
// stop, view's own, unchanged.
func (l *lowering) certifySourceSlots() error {
	l.sourceSlotsJudged = true
	if len(l.sourceSlotWrites) == 0 {
		return nil
	}
	written := map[string][]sourceSlotWrite{}
	for _, write := range l.sourceSlotWrites {
		written[write.name] = append(written[write.name], write)
	}
	if err := l.sourceSlotGenericReads(written); err != nil {
		return err
	}
	if err := l.sourceSlotWholeReads(written); err != nil {
		return err
	}
	if l.result.CheckedFields == nil {
		l.result.CheckedFields = map[string]bool{}
	}
	// Each declared type a checked read reads, in the order first met, with the write it fails to hold.
	var declaredTypes []*checker.Type
	failing := map[*checker.Type]sourceSlotWrite{}
	var refused error
	inspect := func(node any) bool {
		if refused != nil {
			return false
		}
		read, isProperty := node.(ir.Property)
		if !isProperty || len(written[read.Name]) == 0 {
			return true
		}
		writes := written[read.Name]
		if read.ViewTypeID == 0 {
			// A read lowering made as a check of its own (a predicate's tag test, a speculation's field
			// check) carries its view. One with neither names no type to hold the writes to.
			if read.View == "" {
				refused = writes[0].stop
			}
			return true
		}
		declared := l.viewTypes[read.ViewTypeID]
		if declared == nil || declared.Flags()&checker.TypeFlagsInstantiable != 0 {
			refused = writes[0].stop
			return false
		}
		for _, write := range writes {
			if l.sourceSlotHolds(declared, write.written) {
				continue
			}
			if read.ViewReceiverTypeID == 0 {
				refused = write.stop
				return false
			}
			l.result.CheckedFields[ir.CheckedFieldKey(read.ViewReceiverTypeID, read.Name)] = true
			if _, seen := failing[declared]; !seen {
				failing[declared] = write
				declaredTypes = append(declaredTypes, declared)
			}
			break
		}
		return true
	}
	walk(l.result.Main, inspect)
	for _, function := range l.result.Functions {
		walk(function.Body, inspect)
	}
	if refused != nil {
		return refused
	}
	// The read's own check is its field's presence and representation. What's inside it is read again
	// through the declared type, so that type's fields are checked under it too (viewSchema).
	for _, declared := range declaredTypes {
		if _, err := l.viewSchema(failing[declared].target, declared); err != nil {
			return failing[declared].stop
		}
	}
	return nil
}

// sourceSlotHolds is whether a read of a slot at type read holds a value written to it at type written:
// written is assignable to read, as the holder check relates a value to a slot (hostMemberHolder), and
// seeing it as read widens no mutable slot inside it (widened).
func (l *lowering) sourceSlotHolds(read, written *checker.Type) bool {
	return l.enumAssignable(written, read) && l.checker.IsTypeAssignableTo(written, read) && l.widened(written, read, map[[2]*checker.Type]bool{}) == nil
}

// sourceSlotGenericReads gives back a write's stop where the program reads its name through a generic
// receiver: a type parameter other than this, or an indexed access, conditional or substitution type.
// Its read is lowered with the declared type of the constraint's field, while the instantiation it's
// called with may read the field narrower than that, as T["f"] for a T of Required<...>. A read by a
// property access, a literal key or an object binding pattern is weighed, and a plain write isn't.
func (l *lowering) sourceSlotGenericReads(written map[string][]sourceSlotWrite) error {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	var refused error
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if refused != nil {
			return true
		}
		var name string
		var receiver *checker.Type
		switch part.Kind {
		case ast.KindPropertyAccessExpression:
			if hostMemberAssignment(part) == nil && ast.IsIdentifier(part.Name()) {
				name = part.Name().Text()
				receiver = l.checker.GetTypeAtLocation(part.AsPropertyAccessExpression().Expression)
			}
		case ast.KindElementAccessExpression:
			access := part.AsElementAccessExpression()
			if key := l.checker.GetTypeAtLocation(access.ArgumentExpression); key.Flags()&checker.TypeFlagsStringLiteral != 0 {
				name, _ = key.AsLiteralType().Value().(string)
				receiver = l.checker.GetTypeAtLocation(access.Expression)
			}
		case ast.KindBindingElement:
			binding := part.AsBindingElement()
			if part.Parent != nil && part.Parent.Kind == ast.KindObjectBindingPattern {
				key := binding.PropertyName
				if key == nil {
					key = part.Name()
				}
				if key != nil && (ast.IsIdentifier(key) || key.Kind == ast.KindStringLiteral) {
					name = key.Text()
					receiver = l.checker.GetTypeAtLocation(part.Parent)
				}
			}
		}
		if writes := written[name]; len(writes) != 0 && receiver != nil && l.sourceSlotGeneric(receiver) {
			refused = writes[0].stop
			return true
		}
		part.ForEachChild(visit)
		return refused != nil
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return refused
}

// sourceSlotGeneric is whether a receiver, or a member of it, is a type parameter other than this, or an
// indexed access, conditional or substitution type.
func (l *lowering) sourceSlotGeneric(receiver *checker.Type) bool {
	receiver = l.checker.GetNonNullableType(receiver)
	members := []*checker.Type{receiver}
	if receiver.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		members = receiver.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsTypeParameter != 0 && !member.AsTypeParameter().IsThisType() {
			return true
		}
		if member.Flags()&(checker.TypeFlagsIndexedAccess|checker.TypeFlagsConditional|checker.TypeFlagsSubstitution) != 0 {
			return true
		}
	}
	return false
}

// sourceSlotWholeReads gives back a write's stop where a reader that reads every field by its value's
// declared type, JSON.stringify's schema or encodeJson's (wholeObjectCopies' readers), reads a value
// whose type holds the name, at any depth, at a type that doesn't hold what the write wrote.
func (l *lowering) sourceSlotWholeReads(written map[string][]sourceSlotWrite) error {
	seen := map[*checker.Type]bool{}
	for _, copied := range l.wholeObjectCopies {
		if copied.reader == "" {
			// A copy, whose fields are read again at sites of their own.
			continue
		}
		if write := l.sourceSlotHeldNarrower(copied.copied, written, seen); write != nil {
			return write.stop
		}
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	var refused error
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if refused != nil {
			return true
		}
		if part.Kind == ast.KindCallExpression {
			call := part.AsCallExpression()
			callee := ast.SkipParentheses(call.Expression)
			if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "stringify" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "JSON") && len(call.Arguments.Nodes) != 0 {
				if write := l.sourceSlotHeldNarrower(l.checker.GetTypeAtLocation(call.Arguments.Nodes[0]), written, seen); write != nil {
					refused = write.stop
					return true
				}
			}
		}
		part.ForEachChild(visit)
		return refused != nil
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return refused
}

// sourceSlotHeldNarrower is the first write whose name holder, or anything it holds through its
// fields, elements and index signatures, declares at a type that doesn't hold what the write wrote, or
// nil. A generic holder can't be read, so it's the first write recorded. Seen holds the types already
// weighed, each answered nil the second time, since the first answered for it.
func (l *lowering) sourceSlotHeldNarrower(holder *checker.Type, written map[string][]sourceSlotWrite, seen map[*checker.Type]bool) *sourceSlotWrite {
	holder = l.checker.GetNonNullableType(l.concrete(holder))
	if seen[holder] {
		return nil
	}
	seen[holder] = true
	if holder.Flags()&checker.TypeFlagsInstantiable != 0 {
		return &l.sourceSlotWrites[0]
	}
	if holder.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range holder.Types() {
			if write := l.sourceSlotHeldNarrower(member, written, seen); write != nil {
				return write
			}
		}
		return nil
	}
	if holder.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	for _, property := range l.checker.GetPropertiesOfType(holder) {
		declared := l.checker.GetTypeOfSymbol(property)
		for index, write := range written[property.Name] {
			if !l.sourceSlotHolds(declared, write.written) {
				return &written[property.Name][index]
			}
		}
		if write := l.sourceSlotHeldNarrower(declared, written, seen); write != nil {
			return write
		}
	}
	for _, info := range l.checker.GetIndexInfosOfType(holder) {
		if write := l.sourceSlotHeldNarrower(info.ValueType(), written, seen); write != nil {
			return write
		}
	}
	return nil
}
