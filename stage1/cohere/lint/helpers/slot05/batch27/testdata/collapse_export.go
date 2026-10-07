package tailwind

func AdamicPositive(input string, strict bool) bool {
	if strict {
		return isStrictPositiveInteger(input)
	}
	return isPositiveInteger(input)
}
