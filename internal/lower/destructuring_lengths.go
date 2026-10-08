package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Array and string length are intrinsic own data properties. A string's length
// counts UTF-16 units, whereas array destructuring consumes iterator elements.
func (l *lowering) destructureLength(pattern *ast.Node, source *checker.Type, heldAs ir.Type, held int) ([]ir.Statement, error) {
	statements := []ir.Statement{}
	for _, binding := range pattern.AsBindingPattern().Elements.Nodes {
		declared := binding.AsBindingElement()
		if declared.DotDotDotToken != nil || !ast.IsIdentifier(binding.Name()) {
			return nil, l.notYet(binding, "a non-name binding from an array or string property")
		}
		name := binding.Name().Text()
		if declared.PropertyName != nil {
			if declared.PropertyName.Kind == ast.KindComputedPropertyName {
				var known bool
				name, known = l.constantFieldName(declared.PropertyName.AsComputedPropertyName().Expression)
				if !known {
					return nil, l.notYet(declared.PropertyName, "a computed field name")
				}
			} else if ast.IsIdentifier(declared.PropertyName) || declared.PropertyName.Kind == ast.KindStringLiteral {
				name = declared.PropertyName.Text()
			} else {
				return nil, l.notYet(binding, "a non-name array or string property")
			}
		}
		if name != "length" {
			return nil, l.notYet(binding, "destructuring the array or string property "+name)
		}
		property := l.checker.GetPropertyOfType(source, name)
		if property == nil {
			return nil, l.notYet(binding, "destructuring a length the type doesn't name")
		}
		read := ir.Read{Local: held, Of: heldAs}
		value := ir.Expression(ir.Length{Array: read})
		if heldAs == ir.String {
			value = ir.StringLength{Value: read}
		}
		bindings, err := l.declareDestructured(binding, l.checker.GetTypeOfSymbol(property), value)
		if err != nil {
			return nil, err
		}
		statements = append(statements, bindings...)
	}
	return statements, nil
}
