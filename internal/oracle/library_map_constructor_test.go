package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
)

// Use the same private next slot as the helper under test, including c2-12's state slot.
func libraryMapConstructorNext(body []ir.Statement, parameter ir.Expression) ir.Expression {
	name := "next"
	mutateStringExpressions(reflect.ValueOf(&body).Elem(), func(value ir.Expression) ir.Expression {
		if call, ok := value.(ir.CallClosure); ok {
			if property, ok := call.Closure.(ir.Property); ok && (property.Name == "next" || property.Name == "__adamic_iterator_state") {
				name = property.Name
			}
		}
		return value
	})
	return ir.Property{Object: parameter, Name: name, Of: ir.Closure}
}
