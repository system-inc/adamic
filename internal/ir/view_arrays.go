package ir

// Array views currently use the same conservative program-wide read policy as
// object fields. All index reads retain their own declared element contract.
func HasArrayViews(program *Program) bool {
	// Constructor schemas can intern arrays without introducing a checked view.
	// Only admitted view origins enable the conservative program-wide policy.
	if len(program.ViewOrigins) == 0 {
		return false
	}
	for _, contract := range program.ViewContracts {
		if contract.Kind == ViewArray {
			return true
		}
	}
	return false
}

// Metadata has no executable operands. It must not masquerade as an index
// expression in CFG, borrowing, reuse, or ownership traversals.
type ArrayViewRead struct {
	TupleUnion       bool
	View, ViewType   string
	ViewAllowed      []ViewLiteral
	ViewContract     ViewContractID
	ViewTypeID       int
	Element          Type
	Required         bool
	UndefinedAllowed bool
}

func (read ArrayViewRead) Index() ArrayIndex {
	return ArrayIndex{TupleUnion: read.TupleUnion, View: read.View, ViewType: read.ViewType, ViewAllowed: read.ViewAllowed, ViewContract: read.ViewContract, ViewTypeID: read.ViewTypeID, Element: read.Element, Required: read.Required, UndefinedAllowed: read.UndefinedAllowed}
}
