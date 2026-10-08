package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) emitViewCallableOverloadShape(property ir.Property, value, recorded string) string {
	contract := e.program.ViewContracts[property.ViewContract-1]
	signatures := make([]string, len(contract.Members))
	for index, member := range contract.Members {
		child := property
		child.ViewContract = member
		signatures[index] = e.viewCallableExpected(child)
	}
	array := e.temporary()
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_callable_signature *const %s[] = {%s};", array, strings.Join(signatures, ", ")))
	return fmt.Sprintf("((adamic_closure *)adamic_view_callable_overloads((const adamic_heap *)%s, %s, %s, %d, %s, %s, %t))", value, recorded, array, len(signatures), cString(contract.Name), cString(property.View), property.Absent || property.Optional || property.UndefinedAllowed)
}
