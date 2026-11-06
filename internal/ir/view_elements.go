package ir

// ViewElement checks one element at the read that hands it out. Never hoist or
// cache this check: an alias can push a lie between two reads of the same array.
// A successful check preserves the value and its identity.
type ViewElement struct {
	Array, Index Expression
	Relative     bool
	Contract     ViewContractID
	Where        string
	ViewType     string
	Of           Type
	// Unrelated is a read whose element type lowering's key relates to no element type a view
	// holds, so it can't hand out a viewed element: lowering's array check puts it back as the raw
	// ArrayIndex it was, Optional included, and never stops it.
	Unrelated, Optional bool
}

func (read ViewElement) Type() Type { return read.Of }

// ViewElementCheck describes the checked load at an implicit loop or callback
// handout. It carries no array identity or cached certificate.
type ViewElementCheck struct {
	Contract        ViewContractID
	Where, ViewType string
	// Unrelated is ViewElement's: the handout reads raw, and its check is dropped.
	Unrelated bool
}

// ViewArrayUse retains an unsupported library member's source after its helper
// lowers into smaller operations. Unknown members are never implicitly raw.
type ViewArrayUse struct {
	Array         Expression
	Where, Member string
}

// A cast excludes only the view origins it introduces itself. An earlier or
// later caller can still pass a previously viewed array to its operand.
type ViewArrayCast struct {
	Value                  Expression
	Where                  string
	OriginStart, OriginEnd int
}
