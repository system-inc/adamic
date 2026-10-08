package regexp

// MayMatchInsideUnicodePair reports assertion paths whose Node candidate scan
// can distinguish a UTF-16 pair interior from a code-point input position.
// It is conservative about negative lookaround, including nested assertions.
func (p *Program) MayMatchInsideUnicodePair() bool {
	for _, instruction := range p.code {
		if unicodeMode(instruction.flags) && (instruction.op == opAssert && instruction.assertion == NotWordBoundary || instruction.op == opLook && instruction.negative) {
			return true
		}
		if instruction.look != nil && instruction.look.MayMatchInsideUnicodePair() {
			return true
		}
	}
	return false
}
