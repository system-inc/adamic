package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
)

// recordSlotNote is a seed whose written value may be a record, or a seeded captured cell that may hold
// one: the write that may close a cycle through it (nil for a cell), the cell's local, and the record
// types it can put on the cycle, a nil one a record type the checker didn't name.
type recordSlotNote struct {
	write   *fresh.Write
	local   int
	records []*checker.Type
}

// noteRecordSlot keeps a seed whose written value may be a record, with the record types it can write,
// for judgeRecordSlots to judge once graphTypes has selected the program's graph types. The types are
// read here, as the seed is found, before graphTypes adds the allocation views it selects among.
func (f *cycleFinder) noteRecordSlot(value, holder *checker.Type, write *fresh.Write) {
	records := append(f.heldRecords(value), f.writtenRecords(value, holder, write)...)
	if len(records) != 0 {
		f.recordSlots = append(f.recordSlots, recordSlotNote{write: write, local: -1, records: records})
	}
}

// noteRecordCell keeps a seeded captured cell the program assigns again with a value that can be a
// record, with the record types it can hold, for judgeRecordSlots.
func (f *cycleFinder) noteRecordCell(local int, proven *checker.Type) {
	if records := f.heldRecords(proven); len(records) != 0 && f.reassigned(local) {
		f.recordCells = append(f.recordCells, recordSlotNote{local: local, records: records})
	}
}

// judgeRecordSlots keeps in recordStop the first seed, then the first seeded captured cell, that can
// hold a record of a type graphTypes didn't select, for graphTypes to return after the boundary's
// refusals. A record of a selected type is a graph member: its literal is made in a graph region
// (ir.RecordLiteral's GraphTypes, native emit_records.go's dictionaryLiteral), its writes are seeded
// where they're made (graph_flow.go's collect), and follow plans every other maker of it or stops at
// its demand, so it's on a cycle as any member is (#1vb0x5s). A record of any other type is a counted
// allocation, so a cycle closed by writing one into a graph slot keeps it inside the region and leaks
// it: a.r = { k: a.next.slice() } freed 5 of 20 by the counted build before records joined graph
// regions (#a9e21hc). A record a cycle's types only pass near, never written into a slot on it
// (method_bind.a's dictionary, reached through function values), isn't on any cycle at run time, so it
// lowers.
func (f *cycleFinder) judgeRecordSlots() {
	for _, note := range f.recordSlots {
		if !f.unselectedRecord(note.records) {
			continue
		}
		what := "a record written into a graph slot, where " + f.writtenAt(note.write) + " may close a cycle through it, in a graph program: a record of a type the program's graph doesn't select can't join a graph region"
		if note.write.Site == 0 {
			// A write lowering didn't record: still stopped, at the program, and writtenAt names its function.
			f.recordStop = &NotYet{Where: f.l.result.Source, What: what}
			return
		}
		f.recordStop = f.l.notYet(f.l.writeSites[note.write.Site-1].node, what)
		return
	}
	// A captured cell assigned again with a value that can be a record seeds a cycle through that record.
	for _, note := range f.recordCells {
		if f.unselectedRecord(note.records) {
			f.recordStop = f.l.notYet(f.l.localNodes[note.local], "a record held by '"+f.l.result.Locals[note.local].Name+"', a variable a function value captures and the program assigns again, on a cycle, in a graph program: a record of a type the program's graph doesn't select can't join a graph region")
			return
		}
	}
}

// unselectedRecord reports whether any of records is a type graphTypes didn't select, or one the checker
// didn't name, which fails closed.
func (f *cycleFinder) unselectedRecord(records []*checker.Type) bool {
	for _, record := range records {
		if record == nil || !f.recordSelected(record) {
			return true
		}
	}
	return false
}

// recordSelected reports whether graphTypes selected a record type, by its own identity or by the
// concrete one a record literal's GraphTypes names (records.go's recordLiteral).
func (f *cycleFinder) recordSelected(record *checker.Type) bool {
	return f.l.result.GraphTypes[int(record.Id())] || f.l.result.GraphTypes[int(f.l.concrete(record).Id())]
}

// recordElements is the value types of the records a value of the type can be: its own, or each record
// arm's of a union.
func (f *cycleFinder) recordElements(proven *checker.Type) []*checker.Type {
	if element := f.l.recordElement(proven); element != nil {
		return []*checker.Type{element}
	}
	var elements []*checker.Type
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if element := f.l.recordElement(member); element != nil {
				elements = append(elements, element)
			}
		}
	}
	return elements
}

// heldRecords is the record types a value of the type can be, or a container a write makes whole around
// one can hold: a union's record members, and an array's, a tuple's, a map's or a set's type arguments
// that are or hold one. It doesn't look through an object's fields or a function's captures, where
// reachedRecords finds a record held deeper.
func (f *cycleFinder) heldRecords(proven *checker.Type) []*checker.Type {
	var records []*checker.Type
	f.heldRecordsSeen(proven, map[*checker.Type]bool{}, &records)
	return records
}

func (f *cycleFinder) heldRecordsSeen(proven *checker.Type, seen map[*checker.Type]bool, records *[]*checker.Type) {
	if proven == nil || seen[proven] {
		return
	}
	seen[proven] = true
	if f.l.recordElement(proven) != nil {
		*records = append(*records, proven)
		return
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			f.heldRecordsSeen(member, seen, records)
		}
		return
	}
	if f.l.isArrayType(proven) {
		f.heldRecordsSeen(f.l.elementTypeOfArray(proven), seen, records)
		return
	}
	if checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
		for _, argument := range f.l.checker.GetTypeArguments(proven) {
			f.heldRecordsSeen(argument, seen, records)
		}
	}
}

// reassigned reports whether a local is assigned after its declaration anywhere in the program: an
// ir.Assign to it, which can put a record into a cell a function value already captures (let r; const f =
// () => r; r = { f }). A captured local only ever bound by its declaration, such as the receiver bind
// keeps (method_bind.a's bound_value), isn't stopped here. A record that reaches its own cell through
// its initializer (a hoisted function, or a closure the literal makes) is on the cell's cycle component,
// so its type is selected and it's a graph member (#1vb0x5s).
func (f *cycleFinder) reassigned(local int) bool {
	if f.assigned == nil {
		f.assigned = map[int]bool{}
		visit := func(node any) bool {
			if assign, ok := node.(ir.Assign); ok {
				f.assigned[assign.Local] = true
			}
			return true
		}
		walk(f.l.result.Main, visit)
		for _, function := range f.l.result.Functions {
			walk(function.Body, visit)
		}
	}
	return f.assigned[local]
}

// recordWritten reports whether lowering made any write into a record: a write site whose holder is a
// record, or a union with a record among its members (set, defineHidden, ??=, a computed key, or the
// entries form's field write). Object.assign with a record as its target never lowers, so it makes no
// write site: it's NotYet in records.go's recordObjectCallArguments. It fails closed: a record written
// anywhere makes every empty literal count.
func (f *cycleFinder) recordWritten() bool {
	if f.recordWrites == 0 {
		f.recordWrites = -1
		for _, site := range f.l.writeSites {
			if site.holder != nil && len(f.recordElementsOf(site.holder)) != 0 {
				f.recordWrites = 1
				break
			}
		}
	}
	return f.recordWrites > 0
}

// recordElementsOf is the value types of the records a value of the type can be: its own, or each record
// arm's of a union.
func (f *cycleFinder) recordElementsOf(proven *checker.Type) []*checker.Type {
	if element := f.l.recordElement(proven); element != nil {
		return []*checker.Type{element}
	}
	var elements []*checker.Type
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if element := f.l.recordElement(member); element != nil {
				elements = append(elements, element)
			}
		}
	}
	return elements
}

// graphRecordSite is one record literal: where, the record type it's built as, and whether it's empty, {}
// with no entry and no spread, which holds nothing when it's made.
type graphRecordSite struct {
	node   *ast.Node
	record *checker.Type
	empty  bool
}

// recordLiteralType is the record type a record literal is built as: its slot's, as recordLiteralElement
// reads it, or else the literal's own type.
func (l *lowering) recordLiteralType(node *ast.Node) *checker.Type {
	if t := l.slotType(node); t != nil {
		if present := l.checker.GetNonNullableType(t); l.recordElement(present) != nil {
			return present
		}
	}
	return l.checker.GetTypeAtLocation(node)
}

// reachedRecords is the record types the program makes that a value written into a seed's slot can
// reach and that are themselves on the cycle, their values reaching the slot's holder back, by any road
// the finder follows: an object's field, every shape seen as the slot's type (a field a view omits),
// union members, collections' arguments, record elements and the cells function values capture.
// heldRecords asks only the slot's own type, so a.box = { r: { k: a } } and a.count = () => rec, rec a
// record holding a, lowered and leaked: 22 of 102 and 22 of 92 freed by the counted build on main
// 8e3a784a (#t5q4prw). It's asked at the seeds, the writes slotsOf finds may close a cycle, never at
// every record made: a record whose types only pass near a cycle, written into no seed, isn't on one,
// and lowers (method_bind.a's dictionary, #a9e21hc's control). A record type the checker didn't name is
// a nil one, which judgeRecordSlots fails closed on. found is the types already reached, which it adds to.
func (f *cycleFinder) reachedRecords(value, holder *checker.Type, found map[*checker.Type]bool, records *[]*checker.Type) {
	for index := range f.l.graphRecords {
		// An empty literal holds nothing when it's made, and in a program that writes into no record it
		// holds nothing ever, so nothing in it can lead back to the holder: the empty r: {} a graph object
		// starts with and never writes stops nothing (TestGraphFlowWriters' programs, a false stop at
		// 7e4d4fb9). Once any record is written, an empty one counts as any other: a write into it is a seed
		// of its own, judged by the record's selection, but it can still put a container on the cycle that
		// a later write into that container closes, which only this stop sees for a record of a type the
		// graph didn't select (Language's e05 on #t5q4prw). TestGraphEmptyRecordFilledLaterStops holds every
		// road that fills one.
		if f.l.graphRecords[index].empty && !f.recordWritten() {
			continue
		}
		record := f.l.graphRecords[index].record
		if found[record] {
			continue
		}
		// A record the slot reaches whose values can't lead back to the holder isn't on this cycle: a
		// graph object's record of strings. Both reaches are the finder's, so a view hides neither.
		var element *checker.Type
		if record != nil {
			element = f.l.recordElement(record)
		}
		if element == nil || (f.reaches(value, cycleNode{proven: record}) && f.reaches(element, cycleNode{proven: holder})) {
			found[record] = true
			*records = append(*records, record)
		}
	}
}

// writtenRecords is reachedRecords asked of what the seed's writes actually put in the slot, where that
// can be read precisely, and of the slot's type otherwise. A function type links to every closure the
// program makes that fits it, so a field of function type reaches any record a fitting closure
// captures, even one no write ever stores there: method_bind.a writes replaceable.m = (): string =>
// "second", capturing nothing, and the program's bind of a dictionary record fits the same type
// (#t5q4prw, J-1qhv2fc). The precise reading is for one form only: a field written by a plain =, every
// unproven write to it a function literal (arrow or function expression) or a named function's
// identifier. Then the slot holds those functions' own captured cells, and each cell is followed by the
// finder from there. Every other form, a closure reached through a variable, a parameter, a field read
// or a call's result (a bind among them), an element, a Set or a Map, falls back to the slot's type,
// failing closed.
func (f *cycleFinder) writtenRecords(value, holder *checker.Type, write *fresh.Write) []*checker.Type {
	var records []*checker.Type
	found := map[*checker.Type]bool{}
	cells, precise := f.writtenFunctionCells(holder, write)
	if !precise {
		f.reachedRecords(value, holder, found, &records)
		return records
	}
	seen := map[cycleNode]bool{}
	queue := make([]cycleNode, 0, len(cells))
	for _, cell := range cells {
		queue = append(queue, cycleNode{cell: cell + 1})
	}
	for len(queue) != 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if seen[node] {
			continue
		}
		seen[node] = true
		if node.proven != nil {
			f.reachedRecords(node.proven, holder, found, &records)
			continue
		}
		queue = append(queue, f.reachLinks(node)...)
	}
	return records
}

// writtenFunctionCells is the cells captured by every function value the seed's field writes store,
// with precise false when any of them isn't one writtenRecords reads precisely. Every unproven
// write of the field's name is read, a superset of the slot's own (cycleUnproven picks one of them), so
// a write to a like-named field elsewhere can only add cells. An unknown write, or one lowering kept no
// site for, isn't readable, and neither is a private field, whose writes carry a name for each native
// instantiation that cycleUnproven matches and a name match here wouldn't.
func (f *cycleFinder) writtenFunctionCells(holder *checker.Type, write *fresh.Write) ([]int, bool) {
	if write == nil || write.Kind != fresh.WriteField || f.unknownWrite >= 0 {
		return nil, false
	}
	property := f.l.checker.GetPropertyOfType(holder, write.Name)
	if property == nil || len(property.Declarations) != 1 {
		return nil, false
	}
	if declared := property.Declarations[0].Name(); declared == nil || declared.Kind == ast.KindPrivateIdentifier {
		return nil, false
	}
	var cells []int
	for index := range f.writes {
		each := &f.writes[index]
		if each.Proven || each.Kind != fresh.WriteField || each.Name != write.Name {
			continue
		}
		if each.Site == 0 {
			return nil, false
		}
		function := f.assignedFunction(f.l.writeSites[each.Site-1].node)
		if function < 0 {
			return nil, false
		}
		cells = append(cells, f.l.result.Functions[function].Environment...)
	}
	return cells, true
}

// assignedFunction is the function lowered for what a plain = stores into receiver's field, when that
// is a function literal or a named function's identifier, or -1.
func (f *cycleFinder) assignedFunction(receiver *ast.Node) int {
	if receiver == nil || receiver.Parent == nil {
		return -1
	}
	access := receiver.Parent
	switch access.Kind {
	case ast.KindPropertyAccessExpression:
		if access.AsPropertyAccessExpression().Expression != receiver {
			return -1
		}
	case ast.KindElementAccessExpression:
		if access.AsElementAccessExpression().Expression != receiver {
			return -1
		}
	default:
		return -1
	}
	assignment := access.Parent
	if assignment == nil || assignment.Kind != ast.KindBinaryExpression || assignment.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken || assignment.AsBinaryExpression().Left != access {
		return -1
	}
	written := ast.SkipParentheses(assignment.AsBinaryExpression().Right)
	target := written
	switch written.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
	case ast.KindIdentifier:
		symbol := f.l.symbol(written)
		if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindFunctionDeclaration {
			return -1
		}
		target = symbol.Declarations[0]
	default:
		return -1
	}
	for _, closure := range f.l.closureRecords {
		if closure.node == target {
			return closure.function
		}
	}
	return -1
}
