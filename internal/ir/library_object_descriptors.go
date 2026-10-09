package ir

// ObjectDescriptorField records the actual plain shape's slot kind, not an ownership bit.
type ObjectDescriptorField struct {
	Name string
	Of   Type
}
