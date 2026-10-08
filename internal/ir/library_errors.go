package ir

// LibraryMayThrow records JavaScript library exceptions, not terminal soundness checks.
func LibraryMayThrow(node any) bool {
	switch n := node.(type) {
	case Concat, ToFixed, NumberFormat, SetProperty:
		return true
	case StringCall:
		return n.Method == "repeat" || n.Method == "padStart" || n.Method == "padEnd" || n.Method == "normalize"
	case ObjectCall:
		return n.Method == "assign" || n.Method == "setPrototypeOf"
	}
	return false
}
