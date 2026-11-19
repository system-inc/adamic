package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A named write of an optional field f (`=`, compound, logical, `++` or `--`) is admitted where every
// allocation that can reach the store's receiver holds a slot for f, present or reserved, and is NotYet at
// compile time otherwise, an unknown reach among them (#w8nanzw, Compiler's accept of Nov 19). Each native
// object carries the shape it was built with, a literal that leaves f out has no slot for it, and no store
// adds one: a write of a name the shape lacks panics in adamic_object_find (object.c), where Node runs on. So
// the slot is the allocation's own, never the type a view or a cast sees it at. A class instance holds a slot
// for every field its class or a base declares, an optional one too, and none for an optional member an
// interface merges into the class. An object literal holds f only where it wrote f, `f: undefined` and a
// reserved Absent field among them, or where it spreads a value every allocation reaching which holds f. An
// array, a closure, a record and an object this can't find hold none.
//
// A slot that's there must take the store too. A literal built with no contextual type as `{ count: undefined }`
// holds an undefined-represented slot, which only a reference may be stored over (object.c's checked write), and
// a plain store would write a number's word into it, so a write of a number through `count?: number | undefined`
// would stop or misread where Node runs on, as w3 did for head 2. So every object reaching the write is judged as
// head 2 judges one (slotWriteSite): the type it's built at holds the field writable, at a type the write's
// declared type relates to both ways in Adamic's relation, and the literal's own type beside it holds a value of
// that type. A class instance is built at its class's instance type, a literal at the slot graphAllocation noted
// (builtTypes), and a spread's sources by what reaches them. `{ next: undefined }` built at `next?: Node |
// undefined` is held so, since undefined is that reference field's own representation.
//
// Three roads ask it, and reachHoldsSlot is the one answer each gets (TestSlotWritesHaveOneJudgment):
//   - a plain write to an optional own field some producer leaves out (absentOptionalFieldWrite), past the
//     admissions made where it's lowered, a presence test of the same spelling and a const literal giving it;
//   - a write through a checked view to a field the view declares optional (viewWriteReceiverTypeID), whose
//     checked write checks the slot's presence and representation but can't add one (#c7w6j0y);
//   - a named write through a kind-checked cast's union (kindUnionWrites, #b5w3ycg head 2), written at the
//     type its members declare alike, required or optional.
//
// A field held in a Union is refused where it's written (absentOptionalFieldWrite), and writing undefined or
// null into a checked slot that may hold an object stays Rule C's NotYet (recordSourceSlotWrite), since the
// checked write can't change the representation its slot holds. Reach exists only once the program is lowered, so each write is recorded where it's lowered and judged
// by checkSlotWrites after, as checkBoundArguments judges a bind.

// slotWrite is a write recordSlotWrite keeps for checkSlotWrites: part is the write's target, where its NotYet
// stands, receiver its object's node, which writeSite gives every store the write lowers to, name the field
// written, written the type the write declares it at (the field's at its receiver, a view's or the members' of a
// cast's union), and opening the NotYet's words before reachHoldsSlot's. proof is the kind-checked cast's union
// the write goes through, nil for another write. viewed says the receiver is a view, of id view, and the write is
// judged only where that view checks the field: an unchecked one is a plain store, judged as one where it was
// lowered.
type slotWrite struct {
	part, receiver *ast.Node
	name, opening  string
	written        *checker.Type
	proof          *kindUnion
	viewed         bool
	view           int
}

// recordSlotWrite keeps write for checkSlotWrites, once: a write lowered again, in another instance of a
// generic body, keeps nothing twice, and every cast of a union walks the program's writes again, each with a
// proof of its own whose members are that one union's, so the first proof is the one kept.
func (l *lowering) recordSlotWrite(write slotWrite) {
	kept := write
	kept.proof = nil
	if l.slotWritesKept[kept] {
		return
	}
	if l.slotWritesKept == nil {
		l.slotWritesKept = map[slotWrite]bool{}
	}
	l.slotWritesKept[kept] = true
	l.slotWrites = append(l.slotWrites, write)
}

// slotReach is what a slot write is judged on once the program is lowered: the allocation graph, each
// allocation site's object literal, and every named store by the receiver node its writeSite recorded.
type slotReach struct {
	graph    *allocationFlowGraph
	literals map[int]ir.ObjectLiteral
	stores   map[*ast.Node][]ir.SetProperty
}

// checkSlotWrites judges each write recordSlotWrite kept, once lowering is done: every store the write lowered
// to is found by its receiver, and every allocation reaching that store's object must hold a slot for the field
// (reachHoldsSlot), or the write is NotYet in its own words. Every store a named write builds is an
// ir.SetProperty carrying writeSite of its receiver: setProperty's fieldStore and its placeholder reset,
// updateProperty and increment, the logical assignments that take setProperty's statement, and
// finiteElementWrite's arms. A lowering that builds none stores nothing (enumNeverIdentity's Evaluate for a
// target narrowed to never, which one instance of a generic body can take while another stores), and the
// roads that build no SetProperty, a foreign send, a private static's call and super's accessor, never find a
// slot by name. So a store found in one lowering never stands in for one unjudged in another, and a write with
// no store at all is NotYet. lower.go calls it after checkBoundArguments and before checkViewLogicalWrites, so a logical
// assignment through a view is held to the slot first, and the scanning reference makes the same call.
func (l *lowering) checkSlotWrites() error {
	var writes []slotWrite
	for _, write := range l.slotWrites {
		// A view lowered after the write can still make its field checked, so the view is asked only now.
		if !write.viewed || l.result.CheckedField(write.view, write.name) {
			writes = append(writes, write)
		}
	}
	if len(writes) == 0 {
		return nil
	}
	assignAllocationSites(l.result)
	reach := &slotReach{literals: map[int]ir.ObjectLiteral{}, stores: map[*ast.Node][]ir.SetProperty{}}
	inspect := func(visited any) bool {
		switch value := visited.(type) {
		case ir.ObjectLiteral:
			if len(value.GraphTypes) > 0 {
				reach.literals[value.GraphTypes[len(value.GraphTypes)-1]] = value
			}
		case ir.SetProperty:
			if value.Site > 0 && value.Site <= len(l.writeSites) {
				receiver := ast.SkipParentheses(l.writeSites[value.Site-1].node)
				reach.stores[receiver] = append(reach.stores[receiver], value)
			}
		}
		return true
	}
	walk(l.result.Main, inspect)
	for _, function := range l.result.Functions {
		walk(function.Body, inspect)
	}
	reach.graph = l.flowGraph(l.result)
	for _, write := range writes {
		// Every object reaching the write is asked what head 2 asks, its built type's and its own type's field.
		each := func(literal ir.ObjectLiteral) string {
			return l.slotWriteSite(write.name, write.written, literal)
		}
		lowered := false
		judged := map[int]bool{}
		for _, store := range reach.stores[ast.SkipParentheses(write.receiver)] {
			if store.Name != write.name || write.written == nil {
				continue
			}
			lowered = true
			if where := l.reachHoldsSlot(reach, write.name, reach.graph.ReachingAllocations(store.Object), judged, each); where != "" {
				return l.notYet(write.part, write.opening+where)
			}
		}
		if !lowered {
			return l.notYet(write.part, write.opening+", where what reaches its receiver isn't known")
		}
	}
	return nil
}

// reachHoldsSlot is whether every allocation in reaches holds a slot for name: empty when each does, or else
// the NotYet's words after the write's. judged holds the sites already asked, so a spread reaching its own
// literal again is judged by the allocations it was first built from. each is asked of every object holding the
// slot, a spread's sources among them (slotWriteSite, whether the slot takes the write's type), and its words
// stop the write too.
func (l *lowering) reachHoldsSlot(reach *slotReach, name string, reaches ir.AllocationSet, judged map[int]bool, each func(ir.ObjectLiteral) string) string {
	if reaches.Unknown {
		// A parameter with unseen callers, a closure's call, an opaque call's result, a host value.
		return ", where what reaches its receiver isn't known"
	}
	for _, site := range reaches.Sites {
		if judged[site] {
			continue
		}
		judged[site] = true
		literal, isLiteral := reach.literals[site]
		if !isLiteral || literal.Record {
			// An array, a closure, a collection, or an object kept as counted own keys, never a field layout.
			return ", where a value reaching it isn't an object the program builds with a field layout"
		}
		holds := holdsSlot(literal.Fields, name)
		if !holds && literal.Class != 0 && (literal.Class > len(l.result.Classes) || !l.result.Classes[literal.Class-1].Literal) {
			// A class instance's fields are its class's and its bases' declared ones, an optional one reserved,
			// and never a member an interface merges into the class.
			class := "a class"
			if literal.Class <= len(l.result.Classes) {
				class = l.result.Classes[literal.Class-1].Name
			}
			return ", where a value reaching it is an instance of " + class + ", whose class holds no " + name
		}
		if !holds && literal.Spread == nil {
			return ", where a value reaching it is " + l.builtWords(literal) + " without " + name
		}
		if !holds && literal.SpreadMaybeUndefined && !holdsSlot(literal.Empty, name) {
			// A spread of undefined makes the literal's own fields and an undefined one for each field its
			// source's type declares, so a name the type doesn't declare has no slot there.
			return ", where a value reaching it is " + l.builtWords(literal) + " from a spread that may be undefined, without " + name
		}
		if where := each(literal); where != "" {
			return where
		}
		// A literal that didn't write the field copies its source's slot, so what reaches the source is judged as
		// the literal's own; one that did is judged by its sources' types too, as head 2 always judged them.
		if literal.Spread != nil {
			if where := l.reachHoldsSlot(reach, name, reach.graph.ReachingAllocations(literal.Spread), judged, each); where != "" {
				return where
			}
		}
	}
	return ""
}

// holdsSlot is whether fields lay out a slot for name: a field written, reserved Absent, or uninitialized, by
// the name a write stores it under. A private field's name is its storage's, which no named write reaches.
func holdsSlot(fields []ir.Field, name string) bool {
	return slices.ContainsFunc(fields, func(field ir.Field) bool { return field.Name == name && !field.Private })
}

// builtWords names the object literal a slot write's NotYet stops at, by the type graphAllocation noted it
// built at where it has one (noteBuiltType).
func (l *lowering) builtWords(literal ir.ObjectLiteral) string {
	if len(literal.GraphTypes) > 1 {
		if built := l.builtTypes[literal.GraphTypes[0]]; built.layout != nil {
			return "built as " + l.checker.TypeToString(built.layout)
		}
	}
	return "an object literal built"
}

// slotWriteSite judges literal, an object holding a slot for name that reaches a write of it at written, by head
// 2's amended fourth condition, every slot write's since #w8nanzw: the type it's built at holds name writable and
// at a type written relates to both ways, and its own type, beside, holds a value of written's type. It is empty
// when it does, and the NotYet's words after the write's otherwise. A class instance is built at its instance
// type, whose declared fields are its layout.
func (l *lowering) slotWriteSite(name string, written *checker.Type, literal ir.ObjectLiteral) string {
	var built builtType
	if literal.Class != 0 && (literal.Class > len(l.result.Classes) || !l.result.Classes[literal.Class-1].Literal) {
		built.layout = l.classHolders[literal.Class]
	} else if len(literal.GraphTypes) > 1 {
		// An object literal, an accessor literal's class among them, built where graphAllocation noted it.
		built = l.builtTypes[literal.GraphTypes[0]]
	}
	if built.layout == nil {
		return ", where a value reaching it is an object built at a type this can't name"
	}
	if where := l.slotWriteBuilt(name, written, built.layout, true); where != "" {
		return where
	}
	if built.own != nil {
		// The literal's own type, beside the slot it's built in: what it holds must be one of written's.
		return l.slotWriteBuilt(name, written, built.own, false)
	}
	return ""
}

// slotWriteBuilt judges built, a type an object reaching a write of name is built at, against written, the type
// the write declares name at: it holds name as a writable field, at a type written relates to both ways where
// built is its layout, or to written alone where built is a literal's own type beside its layout, since a
// literal's own type holds what it was written with (an Identifier where the layout holds a Node, undefined where
// the layout holds a Node or undefined). Both ways keeps the slot's representation the write's: number against
// number | undefined is a packed word against a plain one, and undefined alone against number | undefined is an
// undefined-represented slot. A field built declares optional is judged alike, its slot being reachHoldsSlot's
// to say. It is empty when built holds the field so, and the NotYet's words after the write's otherwise.
func (l *lowering) slotWriteBuilt(name string, written, built *checker.Type, layout bool) string {
	property := l.checker.GetPropertyOfType(built, name)
	if property == nil || property.Flags&ast.SymbolFlagsProperty == 0 {
		return ", where a value reaching it is built as " + l.checker.TypeToString(built) + ", whose layout holds no " + name
	}
	if layout && l.checker.IsReadonlySymbol(property) {
		return ", where a value reaching it is built as " + l.checker.TypeToString(built) + ", which declares " + name + " readonly"
	}
	held := l.checker.GetTypeOfSymbol(property)
	if held == written || l.castAssignable(held, written) && (!layout || l.castAssignable(written, held)) {
		return ""
	}
	return ", where a value reaching it is built as " + l.checker.TypeToString(built) + ", whose " + name + " is " + l.checker.TypeToString(held) + ", not " + l.checker.TypeToString(written)
}
