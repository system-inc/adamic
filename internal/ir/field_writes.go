package ir

// FieldWrites records known data-slot writes. An incomplete inventory cannot
// justify eliding ownership operations, even when every listed value is static.
type FieldWrites struct {
	Expressions []Expression
	Complete    bool
}
