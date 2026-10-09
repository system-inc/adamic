package ir

// IsGraph reports whether a checker view or an allocation flow site selects
// graph ownership after lowering.
func (program *Program) IsGraph(types []int) bool {
	for _, id := range types {
		if program.GraphTypes[id] {
			return true
		}
	}
	return false
}
