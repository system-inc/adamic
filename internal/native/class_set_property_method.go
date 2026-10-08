package native

const ownMethodMessage = "replacing an inherited method would add an own field to a fixed shape"

// Evaluated after the RHS, exactly where the store checks its receiver.
func (e *emitter) ownMethodGuard(object, name string) {
	e.line("if (adamic_object_optional_field(%s, %s, &%s) == NULL) {", object, cString(name), e.cache())
	e.line("\tstatic const char message[] = %s;", cString(ownMethodMessage))
	e.line("\tadamic_panic(message, sizeof message - 1);")
	e.line("}")
}
