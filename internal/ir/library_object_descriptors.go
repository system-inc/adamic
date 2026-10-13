package ir

// ObjectDescriptorField records the actual plain shape's slot kind, not an ownership bit.
type ObjectDescriptorField struct {
	Name string
	Of   Type
}

// ObjectMayThrow describes catchable property failures, not terminal type checks.
func ObjectMayThrow(node any) bool {
	switch node := node.(type) {
	case SetProperty:
		return true
	case ObjectCall:
		return node.Method == "assign" || node.Method == "defineProperty" || node.Method == "defineProperties"
	}
	return false
}
