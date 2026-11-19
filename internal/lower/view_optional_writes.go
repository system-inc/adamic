package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// viewWriteReceiverTypeID is a field write's ViewReceiverTypeID, the one way a SetProperty gets one, so
// every write a view may check is seen here (TestViewWriteReceiverHoldsEveryWrite holds the producers to
// it). A write to a field its receiver declares optional is kept for checkSlotWrites: target is the write,
// receiver its object, and names the fields it may write, one by name or a computed key's.
//
// A view admits a value whose native shape was fixed where it was made, and one made without the field has
// no slot for it, so the view's checked write, which can't add one, stopped a valid program where Node runs
// it: Language's v1 on f245bfe4, `n.name = "w"` through a branded supertype's predicate view (#c7w6j0y),
// exit 70 on both backends. Every such write was NotYet (Adamic, Lower and Compiler, Nov 2), until reach
// could say the slot is there: where the view checks the field, the write is admitted only when every
// allocation reaching it holds a slot for it (reachHoldsSlot, #w8nanzw), judged once lowering is done, since
// a view lowered after the write can still make its field checked. A required field keeps its checked write,
// since every value a view admits holds it.
func (l *lowering) viewWriteReceiverTypeID(target, receiver *ast.Node, names ...string) int {
	id := l.viewReceiverTypeID(l.checker.GetTypeAtLocation(receiver))
	held := l.viewTypes[id]
	// A class instance isn't set aside: its own fields each get a slot (fieldInitializers writes every
	// property declaration), but an optional member an interface merged into the class declares has none.
	if held == nil {
		return id
	}
	for _, name := range names {
		if property := l.checker.GetPropertyOfType(held, name); property != nil && property.Flags&ast.SymbolFlagsOptional != 0 {
			l.recordSlotWrite(slotWrite{part: target, receiver: receiver, name: name, opening: "a write to optional field " + name + " through a checked view", written: l.checker.GetTypeOfSymbol(property), viewed: true, view: id})
		}
	}
	return id
}
