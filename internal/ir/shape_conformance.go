package ir

// AllocationSet is a may-flow result. Unknown disqualifies erasure even when
// Sites is nonempty; an empty set never establishes conformance.
type AllocationSet struct {
	Sites   []int
	Unknown bool
	Reasons []string
}

// FieldTypeCertificate preserves a declared semantic type independently of the
// physical shape slot. CheckerType is scoped to this compiled program. It is
// evidence about permitted values, never evidence that a slot is initialized.
type FieldTypeCertificate struct {
	CheckerType  int
	DeclaredType string
	Optional     bool
}
