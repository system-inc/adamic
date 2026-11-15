package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
)

// noteRecordSlot keeps the first seed whose written value can be a record, for graphTypes to return
// after the boundary's refusals. A record is never a graph member (a literal makes it as
// ir.RecordLiteral, which nothing adopts), so a cycle closed by writing one into a graph slot keeps a
// counted allocation inside the region and leaks it: a.r = { k: a.next.slice() } freed 5 of 20 by the
// counted build (#a9e21hc). A record a cycle's types only pass near, never written into a slot on it
// (method_bind.a's dictionary, reached through function values), isn't on any cycle at run time, so
// it lowers. A cycle closed by a write into a record is refused before this, at graph_types.go:87.
func (f *cycleFinder) noteRecordSlot(value, holder *checker.Type, write *fresh.Write) {
	if f.recordStop != nil || !(f.holdsRecord(value) || f.writtenReachesRecord(value, holder, write)) {
		return
	}
	what := "a record written into a graph slot, where " + f.writtenAt(write) + " may close a cycle through it, in a graph program: a record can't join a graph region"
	if write.Site == 0 {
		// A write lowering didn't record: still stopped, at the program, and writtenAt names its function.
		f.recordStop = &NotYet{Where: f.l.result.Source, What: what}
		return
	}
	f.recordStop = f.l.notYet(f.l.writeSites[write.Site-1].node, what)
}

// holdsRecord reports whether a value of the type can be a record, or a container a write makes whole
// around one: a union with a record member, or an array, tuple, map or set whose type arguments can.
// It doesn't look through an object's fields or a function's captures, where a record held deeper is a
// gap #1vb0x5s closes by giving records their region.
func (f *cycleFinder) holdsRecord(proven *checker.Type) bool {
	return f.holdsRecordSeen(proven, map[*checker.Type]bool{})
}

func (f *cycleFinder) holdsRecordSeen(proven *checker.Type, seen map[*checker.Type]bool) bool {
	if proven == nil || seen[proven] {
		return false
	}
	seen[proven] = true
	if f.l.recordElement(proven) != nil {
		return true
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if f.holdsRecordSeen(member, seen) {
				return true
			}
		}
		return false
	}
	if f.l.isArrayType(proven) {
		return f.holdsRecordSeen(f.l.elementTypeOfArray(proven), seen)
	}
	if checker.IsTupleType(proven) || f.l.isLibraryType(proven, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
		for _, argument := range f.l.checker.GetTypeArguments(proven) {
			if f.holdsRecordSeen(argument, seen) {
				return true
			}
		}
	}
	return false
}

// reassigned reports whether a local is assigned after its declaration anywhere in the program: an
// ir.Assign to it, which can put a record into a cell a function value already captures (let r; const f =
// () => r; r = { f }). A captured local only ever bound by its declaration, such as the receiver bind
// keeps (method_bind.a's bound_value), isn't stopped here. A record that reaches its own cell through
// its initializer (a hoisted function, or a closure the literal makes) is the gap #1vb0x5s closes.
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

// graphRecordSite is one record literal: where, and the record type it's built as.
type graphRecordSite struct {
	node   *ast.Node
	record *checker.Type
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

// reachesRecord reports whether a value written into a seed's slot can reach a record the program makes
// that is itself on the cycle, its values reaching the slot's holder back, by any road the finder follows: an object's field, every shape seen as the slot's type (a field a view
// omits), union members, collections' arguments, record elements and the cells function values capture.
// holdsRecord asks only the slot's own type, so a.box = { r: { k: a } } and a.count = () => rec, rec a
// record holding a, lowered and leaked: 22 of 102 and 22 of 92 freed by the counted build on main
// 8e3a784a (#t5q4prw). It's asked at the seeds, the writes slotsOf finds may close a cycle, never at
// every record made: a record whose types only pass near a cycle, written into no seed, isn't on one,
// and lowers (method_bind.a's dictionary, #a9e21hc's control). A record type the checker didn't name
// fails closed.
func (f *cycleFinder) reachesRecord(value, holder *checker.Type) bool {
	judged := map[*checker.Type]bool{}
	for index := range f.l.graphRecords {
		record := f.l.graphRecords[index].record
		if record == nil {
			return true
		}
		if judged[record] {
			continue
		}
		judged[record] = true
		// A record the slot reaches whose values can't lead back to the holder isn't on this cycle: a
		// graph object's record of strings. Both reaches are the finder's, so a view hides neither.
		element := f.l.recordElement(record)
		if element == nil {
			return true
		}
		if f.reaches(value, cycleNode{proven: record}) && f.reaches(element, cycleNode{proven: holder}) {
			return true
		}
	}
	return false
}

// writtenReachesRecord is reachesRecord asked of what the seed's writes actually put in the slot, where
// that can be read precisely, and of the slot's type otherwise. A function type links to every closure
// the program makes that fits it, so a field of function type reaches any record a fitting closure
// captures, even one no write ever stores there: method_bind.a writes replaceable.m = (): string =>
// "second", capturing nothing, and the program's bind of a dictionary record fits the same type
// (#t5q4prw, J-1qhv2fc). The precise reading is for one form only: a field written by a plain =, every
// unproven write to it a function literal (arrow or function expression) or a named function's
// identifier. Then the slot holds those functions' own captured cells, and each cell is followed by the
// finder from there. Every other form, a closure reached through a variable, a parameter, a field read
// or a call's result (a bind among them), an element, a Set or a Map, falls back to the slot's type,
// failing closed.
func (f *cycleFinder) writtenReachesRecord(value, holder *checker.Type, write *fresh.Write) bool {
	cells, precise := f.writtenFunctionCells(holder, write)
	if !precise {
		return f.reachesRecord(value, holder)
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
			if f.reachesRecord(node.proven, holder) {
				return true
			}
			continue
		}
		queue = append(queue, f.reachLinks(node)...)
	}
	return false
}

// writtenFunctionCells is the cells captured by every function value the seed's field writes store,
// with precise false when any of them isn't one writtenReachesRecord reads precisely. Every unproven
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
