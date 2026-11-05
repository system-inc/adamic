package ir

// ViewElement checks one element at the read that hands it out. Never hoist or
// cache this check: an alias can push a lie between two reads of the same array.
// A successful check preserves the value and its identity.
type ViewElement struct {
	Value    Expression
	Contract ViewContractID
	Where    string
	ViewType string
	Of       Type
}

func (read ViewElement) Type() Type { return read.Of }
