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
// default for the placeholder's undefined, which a check would stop, so a default there is not yet. A field
// whose placeholder reaches it through a view (PlaceholderViews) is not yet too: object.go's view read of it
// needs its source's storage tagged, which nothing does for a destructured read, and native stopped on a
// compiler bug there (placeholder view has untagged storage). read is the destructuring's own read of the
// field; one that isn't a placeholder's comes back as it is.
func (l *lowering) placeholderDestructuredRead(binding *ast.Node, field *ast.Symbol, fieldType *checker.Type, read ir.Property, defaulted bool) (ir.Expression, error) {
	origin := l.placeholderSymbolOrigin(field)
	direct := origin != ""
	if !direct {
		origin = l.result.PlaceholderViews[l.placeholderKey(field)]
	}
	if origin == "" {
		return read, nil
	}
	if defaulted {
		return nil, l.notYet(binding, "a destructuring default for field "+field.Name+", which may still hold its placeholder (undefined!): Node takes the default there, where a placeholder use stops")
	}
	if !direct {
		return nil, l.notYet(binding, "a destructured binding of field "+field.Name+", whose placeholder (undefined!) reaches it through a view: read it as a property instead")
	}
	path := sourceExpression(binding)
	unset := ir.Property{Object: read.Object, Name: read.Name, Of: read.Of, Absent: read.Absent, Unset: true}
	of, known := l.representation(l.checker.GetNonNullableType(fieldType))
	if !known {
		of = ir.Union
	}
	return ir.PlaceholderUse{Value: unset, Origin: origin, Use: "destructured binding", Path: path, Where: l.program.Where(binding), Of: of}, nil
}
