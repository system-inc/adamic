package core

func Wave09FlagsMessage(flags string, extra []string) string {
	return invalidFlagsMessage(flags, allowedConstructorFlags(NoInvalidRegexpOptions{AllowConstructorFlags: extra}))
}
