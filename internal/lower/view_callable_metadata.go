package lower

// Record optional payload fields without admitting writes through wider views.
func (l *lowering) optionalViewWriteField(name string) {
	if l.result.OptionalViewFields == nil {
		l.result.OptionalViewFields = map[string]bool{}
	}
	l.result.OptionalViewFields[name] = true
}
