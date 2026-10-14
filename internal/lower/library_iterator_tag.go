package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// V8 installs these Symbol.toStringTag values on the shared family prototypes
// (src/init/bootstrapper.cc). Both backends read the inherited family data;
// an iterator result is a plain object, without that tag.
func libraryIteratorTag(value ir.Expression) ir.Expression {
	return ir.CallClosure{Closure: ir.Property{Object: value, Name: "__adamic_iterator_tag", Of: ir.Closure, Method: true}, Returns: ir.String}
}

func (l *lowering) libraryIteratorString(value ir.Expression) ir.Expression {
	return ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: l.constant("[object ")}, libraryIteratorTag(value), ir.StringConstant{Index: l.constant("]")}}}
}

func (l *lowering) libraryIteratorTagType(proven *checker.Type) bool {
	proven = l.concrete(proven)
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if !l.libraryIteratorTagType(member) {
				return false
			}
		}
		return true
	}
	return l.isLibraryType(proven, "MapIterator", "SetIterator", "ArrayIterator", "StringIterator")
}
