package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// The serializer reads after all arguments have been evaluated. A callback uses
// the shared physical adapter and the element's full logical domain at each read.
func (e *emitter) jsonArrayReader(schema *ir.JSONSchema) string {
	if schema.ArrayRead.Element == 0 || !ir.HasArrayViews(e.program) {
		return "NULL"
	}
	read := schema.ArrayRead
	primitive := ir.PrimitiveArrayContract(e.program, read.ViewContract)
	wanted := read.Element
	if primitive {
		wanted = ir.Union
	}
	name := e.temporary() + "_json_array_read"
	e.declarations = append(e.declarations, `#include "view_arrays.h"`, `#include "view_nullish.h"`)
	body := fmt.Sprintf("static adamic_value %s(const adamic_array *array,size_t index,adamic_array *owner){\n adamic_value snapshot;\n const adamic_value *slot=adamic_view_array_at(array,(double)index,false,%t,%d,%s,%s,&snapshot,owner);\n if(slot==NULL)return (adamic_value){.reference=NULL};\n", name, read.UndefinedAllowed, wanted, cString(read.ViewType), cString(read.View))
	if primitive {
		body += fmt.Sprintf(" if (!(%s)) adamic_nullish_failure(%s,%s,snapshot.reference);\n", e.arrayPrimitiveUnionTest(read.ViewContract, "snapshot.reference"), cString(read.View), cString(read.ViewType))
	}
	body += " return *slot;\n}"
	e.declarations = append(e.declarations, body)
	return name
}
