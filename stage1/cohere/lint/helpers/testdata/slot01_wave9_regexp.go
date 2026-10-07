package regexp

func AdamicWave9Width(source string, offset int, mode string) int {
	source = source[offset:]
	switch mode {
	case "bounded":
		return boundedQuantifierWidth(source)
	case "quantifier":
		return quantifierWidth(source)
	default:
		return int(groupKindOf(source))
	}
}
