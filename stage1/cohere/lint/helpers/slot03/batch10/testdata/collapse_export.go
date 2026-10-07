package tailwind

func AdamicURL(value string) bool      { return isURL(value) }
func AdamicAbsolute(value string) bool { return isAbsoluteSize(value) }
func AdamicRelative(value string) bool { return isRelativeSize(value) }
