package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// checkViewOptionalWrites makes NotYet a write through a checked view to a field the view declares
// optional. A view admits a value whose native shape was fixed where it was made, and one made without
// the field has no slot for it, so the view's checked write, which can't add one, stopped a valid program
// where Node runs it: Language's v1 on f245bfe4, `n.name = "w"` through a branded supertype's predicate
// view (#c7w6j0y), exit 70 on both backends. Whether the slot is there can't be known statically, so a
// value that holds the field is NotYet too, and the rule is every checked view's, not the predicate's
// alone (Adamic, Lower and Compiler on #x0nd1mh, Nov 2). A required field keeps its checked write, since
// every value a view admits holds it.
//
// It walks the IR's writes once lowering is done, since a view lowered after the write can still make its
// field checked, and every write form lowers to the SetProperty both emitters check through the view by
// CheckedField: =, a compound assignment, ++ and --, and a logical assignment's write. It runs before
// checkViewLogicalWrites, so a logical assignment to such a field is NotYet for this reason.
func (l *lowering) checkViewOptionalWrites() error {
	var refused error
	visit := func(node any) bool {
		if refused != nil {
			return false
		}
		write, ok := node.(ir.SetProperty)
		if !ok || write.Name == "" || !l.result.CheckedField(write.ViewReceiverTypeID, write.Name) {
			return true
		}
		receiver := l.viewTypes[write.ViewReceiverTypeID]
		if receiver == nil || !l.optionalViewField(receiver, write.Name) {
			return true
		}
		// A write without a recorded site is held all the same, at the program's place.
		where := l.result.Source
		if write.Site != 0 {
			where = l.program.Where(l.writeSites[write.Site-1].node)
		}
		refused = &NotYet{Where: where, What: "a write to optional field " + write.Name + " through a checked view, whose slot a fixed shape can't add yet"}
		return false
	}
	walk(l.result.Main, visit)
	for _, function := range l.result.Functions {
		walk(function.Body, visit)
	}
	return refused
}

// optionalViewField is whether receiver, an object type, declares name optional. A union or an
// intersection of objects counts, since a union receiver's write is checked when a member's view checks
// the field (checkedViewReceiverField), and the checker makes a union's property optional when any
// member's is. A class instance isn't set aside: its own fields each get a slot (fieldInitializers writes
// every property declaration), but an optional member an interface merged into the class declares has
// none.
func (l *lowering) optionalViewField(receiver *checker.Type, name string) bool {
	if receiver.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsUnion|checker.TypeFlagsIntersection) == 0 {
		return false
	}
	property := l.checker.GetPropertyOfType(receiver, name)
	return property != nil && property.Flags&ast.SymbolFlagsOptional != 0
}
