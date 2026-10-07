package tailwind

func AdamicJavaScriptSpace(value rune) bool { return isJavaScriptSpace(value) }
func AdamicBlank(value string) bool         { return isBlank(value) }
func AdamicValueSeparator(value byte) bool  { return isValueSeparator(value) }
