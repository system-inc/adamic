package regexp

func AdamicControlEscape(source string, index, size int, unicode, inClass bool, groups int, named bool) (int, int, int, int, bool, string) {
	result, err := decodeControlEscape(source, index, size, escapeContext{unicode: unicode, inClass: inClass, groups: groups, named: named})
	message := ""
	if err != nil {
		message = err.Error()
	}
	return int(result.kind), int(result.set), int(result.r), result.width, result.negated, message
}

func AdamicNamedBackreference(source string) (string, int, bool) { return namedBackreference(source) }
func AdamicNamedGroupOpener(source string) (string, int, bool)   { return namedGroupOpener(source) }
