package ir

// IsGraph reports whether an allocation's concrete or contextual type belongs
// to the ownership graph selected after lowering.
func (program *Program) IsGraph(types []int) bool {
	for _, id := range types {
		if program.GraphTypes[id] {
			return true
		}
	}
	return false
}
