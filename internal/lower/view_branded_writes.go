package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// checkBrandedOptionalWrites makes NotYet every write through a branded supertype's checked view to a
// field the supertype declares optional. A checked predicate claiming S admits any object (#c7w6j0y), and
// the value's native shape is fixed where it was made: one that left the field out has no slot for it, so
// the view's checked write, which can't add one, stops a valid program where Node runs it (Language's v1
// on f245bfe4, `n.name = "w"` through Element's view on a value made without name, exit 70 on both
// backends). Whether the slot is there can't be known statically, so a value that holds the field is
// NotYet too (Optimizer's ruling, pending Lower's word as the view contract's owner). A required field
// keeps its checked write, since every value the view admits for it must hold it.
//
// It walks the IR's writes once lowering is done, since a view lowered after the write can still make its
// field checked, and every write form lowers to the SetProperty both emitters check through the view:
// =, a compound assignment, ++ and --, and a logical assignment's write. It runs before
// checkViewLogicalWrites, so a logical assignment to such a field is NotYet for this reason first.
func (l *lowering) checkBrandedOptionalWrites() error {
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
		if receiver == nil || !l.brandedOptionalField(receiver, write.Name) {
			return true
		}
		// A write without a recorded site is held all the same, at the program's place.
		where := l.result.Source
		if write.Site != 0 {
			where = l.program.Where(l.writeSites[write.Site-1].node)
		}
		refused = &NotYet{Where: where, What: "a write to optional field " + write.Name + " through a checked predicate's view, whose slot may be absent"}
		return false
	}
	walk(l.result.Main, visit)
	for _, function := range l.result.Functions {
		walk(function.Body, visit)
	}
	return refused
}

// brandedOptionalField is whether receiver is a branded supertype, an object that's no class instance
// as brandedMembership keys one, and declares name optional.
func (l *lowering) brandedOptionalField(receiver *checker.Type, name string) bool {
	if receiver.Flags()&checker.TypeFlagsObject == 0 || isClassInstance(receiver) || !l.brandedSupertype(receiver) {
		return false
	}
	property := l.checker.GetPropertyOfType(receiver, name)
	return property != nil && property.Flags&ast.SymbolFlagsOptional != 0
}
