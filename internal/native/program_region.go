package native

import "fmt"

// Adoption moves storage. Later writes use only the returned pointer.
func (e *emitter) adoptProgram(value, size string, member bool) string {
	if member {
		e.line("%s = adamic_program_adopt_owned(%s, %s);", value, value, size)
	}
	return value
}
func (e *emitter) programArray(value string, member bool) string {
	return e.adoptProgram(value, "sizeof *"+value, member)
}
func (e *emitter) programObject(value string, member bool) string {
	return e.adoptProgram(value, fmt.Sprintf("adamic_object_size(%s->shape->count)", value), member)
}
