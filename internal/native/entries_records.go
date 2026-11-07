package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (nativeEmitter *emitter) recordLiteral(literal ir.ObjectLiteral) string {
	object := nativeEmitter.own(ir.Object, "adamic_record_new(adamic_field_reference)")
	nativeEmitter.programObject(object, literal.ProgramRegion)
	for _, field := range literal.Fields {
		value := nativeEmitter.value(field.Value)
		key := nativeEmitter.recordKey(field.Name)
		nativeEmitter.line(
			"adamic_record_define(%s, %s, (adamic_value){.reference = %s});",
			object,
			key,
			nativeEmitter.kept(value),
		)
	}
	return object
}

func (nativeEmitter *emitter) recordKey(name string) string {
	key := nativeEmitter.temporary()
	nativeEmitter.declarations = append(
		nativeEmitter.declarations,
		fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", key, cString(name)),
	)
	return "&" + key
}

func (nativeEmitter *emitter) recordWrite(statement ir.SetProperty) {
	object := nativeEmitter.value(statement.Object)
	value := nativeEmitter.value(statement.Value)
	nativeEmitter.line(
		"if (%s == NULL) adamic_panic(\"record write on undefined\", sizeof \"record write on undefined\" - 1);",
		object,
	)
	nativeEmitter.line("adamic_object_check_data_write(%s, %s);", object, cString(statement.Name))
	if nativeEmitter.program.FrozenWriteMayThrow(statement) {
		nativeEmitter.checkThrown()
	}
	key := nativeEmitter.recordKey(statement.Name)
	nativeEmitter.recordSet(
		object,
		key,
		fmt.Sprintf("(adamic_value){.reference = %s}", nativeEmitter.kept(value)),
		key,
		nativeEmitter.program.RecordWriteMayThrow(statement),
	)
}

func (nativeEmitter *emitter) hasRecordStorage() bool {
	found := false
	walkExpressions(nativeEmitter.program, func(expression ir.Expression) {
		if literal, ok := expression.(ir.ObjectLiteral); ok && literal.Record {
			found = true
		}
	})
	return found
}
