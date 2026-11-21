package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// recordLiteral makes the older record form, an ObjectLiteral with Record. One adoptGraphObject would
// adopt is made a graph member, as dictionaryLiteral's planned record is (#1vb0x5s); a Program-region
// record keeps its own road and joins no graph region. graphAdoptionBoundary refuses this form in a
// graph program today, so only a change lifting that refusal makes one.
func (e *emitter) recordLiteral(literal ir.ObjectLiteral) string {
	made := fmt.Sprintf("adamic_record_new_capacity(adamic_field_reference, %d)", len(literal.Fields))
	if e.graphObject(literal) {
		made = "adamic_record_new_graph(adamic_field_reference)"
	}
	object := e.own(ir.Object, made)
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
	if e.program.FrozenWriteMayThrow(statement) {
		e.checkThrown()
	}
	key := e.recordKey(statement.Name)
	e.recordSet(object, key, fmt.Sprintf("(adamic_value){.reference = %s}", e.kept(value)), key, e.program.RecordWriteMayThrow(statement))
}

// hasRecordStorage is whether any object literal is a record, read from the emitter's one walk of the
// program (prepareObjectMetadata).
func (e *emitter) hasRecordStorage() bool {
	e.prepareObjectMetadata()
	return e.objectMetadata.records
}
