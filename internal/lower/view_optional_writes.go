package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// viewOptionalWrite is a field write whose receiver declares the field optional: where it's written,
// the receiver's view id and the field's name, judged by checkViewOptionalWrites once every view is in.
type viewOptionalWrite struct {
	node           *ast.Node
	receiverTypeID int
	name           string
}

// viewWriteReceiverTypeID is a field write's ViewReceiverTypeID, the one way a SetProperty gets one, so
// every write a view may check is seen here (TestViewWriteReceiverHoldsEveryWrite holds the producers to
// it). A write to a field its receiver declares optional is kept for checkViewOptionalWrites: target is
// the write, receiver its object, and names the fields it may write, one by name or a computed key's.
func (lowering *lowering) viewWriteReceiverTypeID(target, receiver *ast.Node, names ...string) int {
	id := lowering.viewReceiverTypeID(lowering.checker.GetTypeAtLocation(receiver))
	held := lowering.viewTypes[id]
	// A class instance isn't set aside: its own fields each get a slot (fieldInitializers writes every
	// property declaration), but an optional member an interface merged into the class declares has none.
	if held == nil {
		return id
	}
	for _, name := range names {
		if property := lowering.checker.GetPropertyOfType(
			held,
			name,
		); property != nil && property.Flags&ast.SymbolFlagsOptional != 0 {
			lowering.viewOptionalWrites = append(
				lowering.viewOptionalWrites,
				viewOptionalWrite{node: target, receiverTypeID: id, name: name},
			)
		}
	}
	return id
}

// checkViewOptionalWrites makes NotYet a write through a checked view to a field the view declares
// optional. A view admits a value whose native shape was fixed where it was made, and one made without
// the field has no slot for it, so the view's checked write, which can't add one, stopped a valid program
// where Node runs it: Language's v1 on f245bfe4, `n.name = "w"` through a branded supertype's predicate
// view (#c7w6j0y), exit 70 on both backends. Whether the slot is there can't be known statically, so a
// value that holds the field is NotYet too, and the rule is every checked view's, not the predicate's
// alone (Adamic, Lower and Compiler, Nov 2). A required field keeps its checked write, since every value
// a view admits holds it. It's judged once lowering is done, since a view lowered after the write can
// still make its field checked, and before checkViewLogicalWrites, so a logical assignment to such a
// field is NotYet for this reason.
func (lowering *lowering) checkViewOptionalWrites() error {
	for _, write := range lowering.viewOptionalWrites {
		if lowering.result.CheckedField(write.receiverTypeID, write.name) {
			return lowering.notYet(
				write.node,
				"a write to optional field "+write.name+" through a checked view, whose slot a fixed shape can't add",
			)
		}
	}
	return nil
}
