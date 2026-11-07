package native

import "fmt"

// Adoption moves storage. Later writes use only the returned pointer. graph is the site's own
// answer as if it were not a Program member: the owner the runtime falls back to once a second
// Program begins (program_region.c).
func (e *emitter) adoptProgram(value, size string, member, graph bool) string {
	if member {
		e.line("%s = adamic_program_adopt_or_fallback(%s, %s, %t);", value, value, size, graph)
	}
	return value
}
func (e *emitter) programArray(value string, member, graph bool) string {
	return e.adoptProgram(value, "sizeof *"+value, member, graph)
}
func (e *emitter) programObject(value string, member, graph bool) string {
	return e.adoptProgram(value, fmt.Sprintf("adamic_object_allocation_size(%s)", value), member, graph)
}
