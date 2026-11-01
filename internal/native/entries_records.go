package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) recordLiteral(literal ir.ObjectLiteral) string {
	object := e.own(ir.Object, "adamic_record_new(true)")
	e.programObject(object, literal.ProgramRegion)
	for _, field := range literal.Fields {
		value := e.value(field.Value)
		key := e.recordKey(field.Name)
		e.line("adamic_record_define(%s, %s, (adamic_value){.reference = %s});", object, key, e.kept(value))
	}
	return object
}

func (e *emitter) recordKey(name string) string {
	key := e.temporary()
	e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", key, cString(name)))
	return "&" + key
}

func (e *emitter) recordWrite(statement ir.SetProperty) {
	object := e.value(statement.Object)
	value := e.value(statement.Value)
	e.line("if (%s == NULL) adamic_panic(\"record write on undefined\", sizeof \"record write on undefined\" - 1);", object)
	e.line("adamic_object_check_data_write(%s, %s);", object, cString(statement.Name))
	key := e.recordKey(statement.Name)
	e.recordSet(object, key, fmt.Sprintf("(adamic_value){.reference = %s}", e.kept(value)), key, e.program.RecordWriteMayThrow(statement))
}

func (e *emitter) hasRecordStorage() bool {
	found := false
	walkExpressions(e.program, func(expression ir.Expression) {
		if literal, ok := expression.(ir.ObjectLiteral); ok && literal.Record {
			found = true
		}
	})
	return found
}
