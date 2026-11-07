package native

import "fmt"

// Adoption moves storage. Later writes use only the returned pointer.
func (nativeEmitter *emitter) adoptProgram(value, size string, member bool) string {
	if member {
		nativeEmitter.line("%s = adamic_program_adopt_owned(%s, %s);", value, value, size)
	}
	return value
}
func (nativeEmitter *emitter) programArray(value string, member bool) string {
	return nativeEmitter.adoptProgram(value, "sizeof *"+value, member)
}
func (nativeEmitter *emitter) programObject(value string, member bool) string {
	return nativeEmitter.adoptProgram(value, fmt.Sprintf("adamic_object_allocation_size(%s)", value), member)
}
