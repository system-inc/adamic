package ir

// LibraryMayThrow records JavaScript library exceptions, not terminal soundness checks.
func LibraryMayThrow(node any) bool {
	switch n := node.(type) {
	// RegExp construction rejects invalid patterns or flags with SyntaxError.
	case Concat, ToFixed, NumberFormat, SetProperty, RegExpNew:
		return true
	case StringCall:
		return n.Method == "repeat" || n.Method == "padStart" || n.Method == "padEnd" || n.Method == "normalize"
	case ObjectCall:
		// Property definitions throw when descriptors violate existing attributes.
		return n.Method == "assign" || n.Method == "setPrototypeOf" || n.Method == "defineProperty" || n.Method == "defineProperties"
	}
	return false
}
