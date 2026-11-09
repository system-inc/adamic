package lower

import (
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
func (f *cycleFinder) noteRecordSlot(value *checker.Type, write *fresh.Write) {
	if f.recordStop != nil || !f.holdsRecord(value) {
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
