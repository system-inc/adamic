package regexp

func AdamicWave8Decimal(source string, offset int) (int, int) { return decimalEscape(source, offset) }
func AdamicWave8Octal(source string, offset int) (rune, int) {
	return decodeLegacyOctal(source, offset)
}
func AdamicWave8Covers(kind uint8, low, high, value rune) bool {
	return (classAtom{kind: classAtomKind(kind), lo: low, hi: high}).covers(value)
}
