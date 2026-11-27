package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// placeholderDestructuredRead is a destructured name's read of a field that may hold a placeholder
// (#jjf9ppe): const { parent } = node, where parent was written undefined!, read it with no check, so
// native dereferenced the null a placeholder leaves where Node throws a TypeError at parent.kind. A field
// taken out by destructuring is a placeholder use as node.parent read into a const is, so it's read as
// object.go reads that field, its placeholder storage with undefined kept, and checked as an assignment
// is (placeholderUseValue): loudly unset where it's still the placeholder. With a default, Node takes the
// default for the placeholder's undefined, which a check would stop, so a default there is not yet. The
// field is read at its placeholder storage, as typeOfSymbol says (a union, which can hold the undefined),
// never at the name's own representation: native's view read takes nothing else, and the first cut, which
// read it as the name's, stopped on a compiler bug there (placeholder view has untagged storage). read is the
// destructuring's own read of the field; one that isn't a placeholder's comes back as it is. slot says the name
// is a placeholder slot itself, as an unannotated const { parent } = node makes it, the way const parent =
// node.parent does (walkPlaceholderSlots, #0d56xpd): the unset flows into it unchecked (rulings.md:55, other
// placeholder slots), and each read of the name is the placeholder read that checks it.
func (l *lowering) placeholderDestructuredRead(binding *ast.Node, field *ast.Symbol, fieldType *checker.Type, read ir.Property, defaulted bool, slot bool) (ir.Expression, error) {
	origin := l.placeholderSymbolOrigin(field)
	direct := origin != ""
	if !direct {
		origin = l.result.PlaceholderViews[l.placeholderKey(field)]
	}
	if origin == "" {
		if slot {
			// The name took a placeholder this field doesn't hold; its storage would be the wrong one.
			return nil, l.notYet(binding, "a destructured placeholder slot whose field holds no placeholder")
		}
		return read, nil
	}
	if defaulted {
		return nil, l.notYet(binding, "a destructuring default for field "+field.Name+", which may still hold its placeholder (undefined!): Node takes the default there, where a placeholder use stops")
	}
	stored, err := l.typeOfSymbol(binding, field)
	if err != nil {
		return nil, err
	}
	path := sourceExpression(binding)
	unset := ir.Property{Object: read.Object, Name: read.Name, Of: stored, Absent: read.Absent, Unset: true}
	if !direct {
		// The field is seen through a view whose source slot may hold the placeholder, as object.go's
		// view arm reads it: its present representation is checked by kind at the read.
		present, _ := l.representation(l.checker.GetNonNullableType(fieldType))
		unset.UnsetType = present
		if present != ir.Number && present != ir.Boolean && present != ir.String && present != ir.Object && present != ir.Array && present != ir.Map {
			return nil, l.notYet(binding, "a placeholder field view without a physical-kind check for this representation")
		}
		unset.View = read.View
		unset.ViewType = read.ViewType
		unset.ViewAllowed = read.ViewAllowed
		l.result.PlaceholderChecks = append(l.result.PlaceholderChecks, ir.PlaceholderCheck{Origin: origin, Use: "field representation", Path: path, Where: l.program.Where(binding), Status: "checked"})
	}
	if slot {
		return unset, nil
	}
	of, known := l.representation(l.checker.GetNonNullableType(fieldType))
	if !known {
		of = ir.Union
	}
	return ir.PlaceholderUse{Value: unset, Origin: origin, Use: "destructured binding", Path: path, Where: l.program.Where(binding), Of: of}, nil
}
